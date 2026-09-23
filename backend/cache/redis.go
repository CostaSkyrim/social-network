package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisConfig struct {
	Addr         string `json:"addr"`
	Password     string `json:"password"`
	DB           int    `json:"db"`
	PoolSize     int    `json:"pool_size"`
	MinIdleConns int    `json:"min_idle_conns"`
	MaxRetries   int    `json:"max_retries"`
	DialTimeout  string `json:"dial_timeout"`
	ReadTimeout  string `json:"read_timeout"`
	WriteTimeout string `json:"write_timeout"`
}

type RedisClient struct {
	client *redis.Client
	cfg    *RedisConfig
}

var (
	SessionTTL = 24 * time.Hour
	UserTTL    = 15 * time.Minute
	GroupTTL   = 10 * time.Minute
)

func NewRedisClient(ctx context.Context, cfg *RedisConfig) (*RedisClient, error) {
	if cfg.Addr == "" {
		cfg.Addr = "localhost:6379"
	}
	if cfg.PoolSize == 0 {
		cfg.PoolSize = 10
	}
	if cfg.MinIdleConns == 0 {
		cfg.MinIdleConns = 5
	}
	if cfg.MaxRetries == 0 {
		cfg.MaxRetries = 3
	}

	dialTimeout := 5 * time.Second
	readTimeout := 3 * time.Second
	writeTimeout := 3 * time.Second

	if cfg.DialTimeout != "" {
		if d, err := time.ParseDuration(cfg.DialTimeout); err == nil {
			dialTimeout = d
		}
	}
	if cfg.ReadTimeout != "" {
		if d, err := time.ParseDuration(cfg.ReadTimeout); err == nil {
			readTimeout = d
		}
	}
	if cfg.WriteTimeout != "" {
		if d, err := time.ParseDuration(cfg.WriteTimeout); err == nil {
			writeTimeout = d
		}
	}

	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		MaxRetries:   cfg.MaxRetries,
		DialTimeout:  dialTimeout,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Printf("✅ Redis connected at %s (db %d, pool %d)", cfg.Addr, cfg.DB, cfg.PoolSize)
	return &RedisClient{client: client, cfg: cfg}, nil
}

func (rc *RedisClient) Close() error {
	log.Println("Closing Redis connection...")
	return rc.client.Close()
}

// ---- Generic helpers ----

func (rc *RedisClient) SetJSON(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("redis set marshal: %w", err)
	}
	return rc.client.Set(ctx, key, data, ttl).Err()
}

func (rc *RedisClient) GetJSON(ctx context.Context, key string, dest interface{}) error {
	data, err := rc.client.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

func (rc *RedisClient) Delete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	return rc.client.Del(ctx, keys...).Err()
}

// ---- Session cache ----

func SessionKey(sessionID string) string {
	return "session:" + sessionID
}

func UserSessionsKey(userID int64) string {
	return "user_sessions:" + fmt.Sprint(userID)
}

func (rc *RedisClient) CacheSession(ctx context.Context, key string, userID int64, ttl time.Duration) error {
	return rc.client.Set(ctx, key, userID, ttl).Err()
}

func (rc *RedisClient) GetSession(ctx context.Context, key string) (int64, error) {
	userID, err := rc.client.Get(ctx, key).Int64()
	if err != nil {
		return 0, err
	}
	return userID, nil
}

// InvalidateSessionLookup drops the cached session→user lookup used by the
// SQLite session backend, so a session that has expired (or been revoked) stops
// authenticating immediately instead of at the end of its cache TTL.
func (rc *RedisClient) InvalidateSessionLookup(ctx context.Context, sessionID string) error {
	return rc.client.Del(ctx, SessionKey(sessionID)).Err()
}

// ---- Session store (Redis as source of truth) ----

// SessionRecord is the full session payload stored in Redis when
// sessions.storage = "redis". It uses a distinct "rsession:" namespace so it
// never collides with the int64 session cache used by the sqlite backend.
type SessionRecord struct {
	UserID    int64     `json:"user_id"`
	IP        string    `json:"ip"`
	UserAgent string    `json:"user_agent"`
	CreatedAt time.Time `json:"created_at"`
}

func SessionRecordKey(sessionID string) string {
	return "rsession:" + sessionID
}

func (rc *RedisClient) SetSessionRecord(ctx context.Context, sessionID string, rec SessionRecord, ttl time.Duration) error {
	return rc.SetJSON(ctx, SessionRecordKey(sessionID), rec, ttl)
}

func (rc *RedisClient) GetSessionRecord(ctx context.Context, sessionID string) (*SessionRecord, error) {
	var rec SessionRecord
	if err := rc.GetJSON(ctx, SessionRecordKey(sessionID), &rec); err != nil {
		return nil, err
	}
	return &rec, nil
}

func (rc *RedisClient) DeleteSessionRecord(ctx context.Context, sessionID string) error {
	return rc.client.Del(ctx, SessionRecordKey(sessionID)).Err()
}

// AddUserSession tracks a session id in the user's session set. The set gets
// the session TTL because nothing prunes it when a session simply expires, so
// without an expiry it would grow without bound.
func (rc *RedisClient) AddUserSession(ctx context.Context, userID int64, sessionID string, ttl time.Duration) error {
	key := UserSessionsKey(userID)

	pipe := rc.client.Pipeline()
	pipe.SAdd(ctx, key, sessionID)
	if ttl > 0 {
		pipe.Expire(ctx, key, ttl)
	}
	_, err := pipe.Exec(ctx)
	return err
}

func (rc *RedisClient) RemoveUserSession(ctx context.Context, userID int64, sessionID string) error {
	return rc.client.SRem(ctx, UserSessionsKey(userID), sessionID).Err()
}

// GetUserSessionRecords returns the live session ids tracked for a user,
// dropping entries whose session record has already expired. The set is
// otherwise only pruned on an explicit logout, so expired sessions would
// accumulate in it.
func (rc *RedisClient) GetUserSessionRecords(ctx context.Context, userID int64) ([]string, error) {
	key := UserSessionsKey(userID)

	tracked, err := rc.client.SMembers(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	live := make([]string, 0, len(tracked))
	for _, sessionID := range tracked {
		exists, err := rc.client.Exists(ctx, SessionRecordKey(sessionID)).Result()
		if err != nil {
			return nil, err
		}
		if exists == 1 {
			live = append(live, sessionID)
			continue
		}
		if err := rc.client.SRem(ctx, key, sessionID).Err(); err != nil {
			return nil, err
		}
	}

	return live, nil
}

func (rc *RedisClient) ClearUserSessions(ctx context.Context, userID int64) error {
	return rc.client.Del(ctx, UserSessionsKey(userID)).Err()
}

// ---- User cache ----

func UserKey(userID int64) string {
	return "user:" + fmt.Sprint(userID)
}

func (rc *RedisClient) CacheUser(ctx context.Context, userID int64, data interface{}, ttl time.Duration) error {
	return rc.SetJSON(ctx, UserKey(userID), data, ttl)
}

func (rc *RedisClient) GetCachedUser(ctx context.Context, userID int64, dest interface{}) error {
	return rc.GetJSON(ctx, UserKey(userID), dest)
}

func (rc *RedisClient) InvalidateUser(ctx context.Context, userID int64) error {
	return rc.client.Del(ctx, UserKey(userID)).Err()
}

// ---- User lookup-index cache (email / nickname → user ID) ----
//
// These store only a numeric user ID (never the password hash or other
// sensitive fields) so the hot login/signup/invite/nickname-collision lookups
// can skip a full table scan on the email/nickname column.

func UserEmailKey(email string) string {
	return "user_email:" + email
}

func UserNicknameKey(nickname string) string {
	return "user_nickname:" + nickname
}

func (rc *RedisClient) CacheUserIDByEmail(ctx context.Context, email string, userID int64, ttl time.Duration) error {
	return rc.client.Set(ctx, UserEmailKey(email), userID, ttl).Err()
}

func (rc *RedisClient) GetUserIDByEmail(ctx context.Context, email string) (int64, error) {
	return rc.client.Get(ctx, UserEmailKey(email)).Int64()
}

func (rc *RedisClient) CacheUserIDByNickname(ctx context.Context, nickname string, userID int64, ttl time.Duration) error {
	return rc.client.Set(ctx, UserNicknameKey(nickname), userID, ttl).Err()
}

func (rc *RedisClient) GetUserIDByNickname(ctx context.Context, nickname string) (int64, error) {
	return rc.client.Get(ctx, UserNicknameKey(nickname)).Int64()
}

func (rc *RedisClient) InvalidateUserEmail(ctx context.Context, email string) error {
	return rc.client.Del(ctx, UserEmailKey(email)).Err()
}

func (rc *RedisClient) InvalidateUserNickname(ctx context.Context, nickname string) error {
	return rc.client.Del(ctx, UserNicknameKey(nickname)).Err()
}

// ---- Group cache ----

func GroupKey(groupID int64) string {
	// v2: the payload switched from Group (which drops the internal numeric
	// ids) to the id-preserving groupCacheEntry DTO.
	return "group:v2:" + fmt.Sprint(groupID)
}

func (rc *RedisClient) CacheGroup(ctx context.Context, groupID int64, data interface{}, ttl time.Duration) error {
	return rc.SetJSON(ctx, GroupKey(groupID), data, ttl)
}

func (rc *RedisClient) GetCachedGroup(ctx context.Context, groupID int64, dest interface{}) error {
	return rc.GetJSON(ctx, GroupKey(groupID), dest)
}

func (rc *RedisClient) InvalidateGroup(ctx context.Context, groupID int64) error {
	return rc.client.Del(ctx, GroupKey(groupID)).Err()
}

// ---- Group member-list cache ----

func GroupMembersKey(groupID int64) string {
	// v2: the payload switched from []GroupMember (which drops the internal
	// numeric ids) to the id-preserving groupMemberCacheEntry DTO.
	return "group_members:v2:" + fmt.Sprint(groupID)
}

func (rc *RedisClient) CacheGroupMembers(ctx context.Context, groupID int64, members interface{}, ttl time.Duration) error {
	return rc.SetJSON(ctx, GroupMembersKey(groupID), members, ttl)
}

func (rc *RedisClient) GetCachedGroupMembers(ctx context.Context, groupID int64, dest interface{}) error {
	return rc.GetJSON(ctx, GroupMembersKey(groupID), dest)
}

func (rc *RedisClient) InvalidateGroupMembers(ctx context.Context, groupID int64) error {
	return rc.client.Del(ctx, GroupMembersKey(groupID)).Err()
}

// ---- Group lookup-index cache (uuid → group ID) ----
//
// The group object cache is keyed by numeric id, but requests arrive with the
// group's UUID. Without this index GetGroupByUUID has to read the database on
// every request just to translate the UUID, which makes the group cache
// pointless on the hot path.

func GroupUUIDKey(uuid string) string {
	return "group_uuid:" + uuid
}

func (rc *RedisClient) CacheGroupIDByUUID(ctx context.Context, uuid string, groupID int64, ttl time.Duration) error {
	return rc.client.Set(ctx, GroupUUIDKey(uuid), groupID, ttl).Err()
}

func (rc *RedisClient) GetGroupIDByUUID(ctx context.Context, uuid string) (int64, error) {
	return rc.client.Get(ctx, GroupUUIDKey(uuid)).Int64()
}

func (rc *RedisClient) InvalidateGroupUUID(ctx context.Context, uuid string) error {
	return rc.client.Del(ctx, GroupUUIDKey(uuid)).Err()
}

// ---- Rate limiting ----

const rateLimitPrefix = "ratelimit:"

func (rc *RedisClient) CheckRateLimit(ctx context.Context, key string, limit int64, windowSeconds float64) (bool, error) {
	redisKey := rateLimitPrefix + key
	now := time.Now().UnixMilli()
	windowStart := now - int64(windowSeconds*1000)

	pipe := rc.client.Pipeline()
	pipe.ZRemRangeByScore(ctx, redisKey, "0", fmt.Sprint(windowStart))
	pipe.ZCard(ctx, redisKey)
	pipe.ZAdd(ctx, redisKey, redis.Z{Score: float64(now), Member: fmt.Sprint(now)})
	pipe.Expire(ctx, redisKey, time.Duration(windowSeconds*2)*time.Second)

	cmds, err := pipe.Exec(ctx)
	if err != nil {
		return false, fmt.Errorf("rate limit pipeline: %w", err)
	}

	count := cmds[1].(*redis.IntCmd).Val()
	if count >= limit {
		return true, nil
	}

	return false, nil
}

// ---- Pub/Sub for future WebSocket + RabbitMQ integration ----

func (rc *RedisClient) Publish(ctx context.Context, channel string, message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("redis publish marshal: %w", err)
	}
	return rc.client.Publish(ctx, channel, data).Err()
}

func (rc *RedisClient) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	return rc.client.Subscribe(ctx, channels...)
}

// ---- Presence tracking ----

const presencePrefix = "presence:"
const presenceTTL = 30 * time.Second

func (rc *RedisClient) SetUserOnline(ctx context.Context, userID int64) error {
	return rc.client.Set(ctx, presencePrefix+fmt.Sprint(userID), "1", presenceTTL).Err()
}

func (rc *RedisClient) SetUserOffline(ctx context.Context, userID int64) error {
	return rc.client.Del(ctx, presencePrefix+fmt.Sprint(userID)).Err()
}

func (rc *RedisClient) IsUserOnline(ctx context.Context, userID int64) (bool, error) {
	n, err := rc.client.Exists(ctx, presencePrefix+fmt.Sprint(userID)).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (rc *RedisClient) GetOnlineUsers(ctx context.Context, userIDs []int64) ([]int64, error) {
	if len(userIDs) == 0 {
		return nil, nil
	}
	keys := make([]string, len(userIDs))
	for i, id := range userIDs {
		keys[i] = presencePrefix + fmt.Sprint(id)
	}
	results, err := rc.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}
	var online []int64
	for i, val := range results {
		if val != nil {
			online = append(online, userIDs[i])
		}
	}
	return online, nil
}

// ChannelWSFanout carries WebSocket messages between backend instances so a
// message delivered on one instance reaches clients connected to another.
const ChannelWSFanout = "ws:fanout"
