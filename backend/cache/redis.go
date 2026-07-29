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
	PostTTL    = 10 * time.Minute
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

func (rc *RedisClient) Client() *redis.Client {
	return rc.client
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

// ---- User cache ----

func UserKey(userID int64) string {
	return "user:" + fmt.Sprint(userID)
}

func UserByEmailKey(email string) string {
	return "user:email:" + email
}

func UserByUUIDKey(uuid string) string {
	return "user:uuid:" + uuid
}

func (rc *RedisClient) CacheUser(ctx context.Context, userID int64, data interface{}, ttl time.Duration) error {
	return rc.SetJSON(ctx, UserKey(userID), data, ttl)
}

func (rc *RedisClient) GetCachedUser(ctx context.Context, userID int64, dest interface{}) error {
	return rc.GetJSON(ctx, UserKey(userID), dest)
}

func (rc *RedisClient) InvalidateUser(ctx context.Context, userID int64) error {
	return rc.client.Del(ctx, UserKey(userID), UserByUUIDKey("")).Err()
}

// ---- Post cache ----

func PostKey(postID int64) string {
	return "post:" + fmt.Sprint(postID)
}

func UserPostsKey(userID int64) string {
	return "user_posts:" + fmt.Sprint(userID)
}

func (rc *RedisClient) CachePost(ctx context.Context, postID int64, data interface{}, ttl time.Duration) error {
	return rc.SetJSON(ctx, PostKey(postID), data, ttl)
}

func (rc *RedisClient) GetCachedPost(ctx context.Context, postID int64, dest interface{}) error {
	return rc.GetJSON(ctx, PostKey(postID), dest)
}

func (rc *RedisClient) InvalidatePost(ctx context.Context, postID int64) error {
	return rc.client.Del(ctx, PostKey(postID)).Err()
}

// ---- Group cache ----

func GroupKey(groupID int64) string {
	return "group:" + fmt.Sprint(groupID)
}

func UserGroupsKey(userID int64) string {
	return "user_groups:" + fmt.Sprint(userID)
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

// Pre-defined channel names for future use
const (
	ChannelNewMessage      = "chat:new_message"
	ChannelNewNotification = "notification:new"
	ChannelGroupMessage    = "group:message"
	ChannelUserOnline      = "presence:online"
	ChannelUserOffline     = "presence:offline"
)
