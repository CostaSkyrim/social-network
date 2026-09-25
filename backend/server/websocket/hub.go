package websocket

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"social-network/backend/cache"
	database "social-network/backend/db/sql"

	"github.com/google/uuid"
)

type Hub struct {
	clients    map[int64]map[*Client]bool
	uuidToID   map[string]int64
	register   chan *Client
	Unregister chan *Client
	broadcast  chan *WSMessage
	mu         sync.RWMutex

	db          *database.DataBase
	redisClient *cache.RedisClient

	// instanceID identifies this process so cross-instance fan-out can ignore
	// its own echoed messages.
	instanceID string

	offlineTimers map[int64]*time.Timer
	timersMu      sync.Mutex

	ctx    context.Context
	cancel context.CancelFunc
}

const offlineDebounce = 3 * time.Second

func NewHub(db *database.DataBase, redisClient *cache.RedisClient) *Hub {
	ctx, cancel := context.WithCancel(context.Background())

	return &Hub{
		clients:       make(map[int64]map[*Client]bool),
		uuidToID:      make(map[string]int64),
		register:      make(chan *Client),
		Unregister:    make(chan *Client),
		broadcast:     make(chan *WSMessage, 256),
		db:            db,
		redisClient:   redisClient,
		instanceID:    uuid.New().String(),
		offlineTimers: make(map[int64]*time.Timer),
		ctx:           ctx,
		cancel:        cancel,
	}
}

func NewHubForTest(db *database.DataBase) *Hub {
	ctx, cancel := context.WithCancel(context.Background())

	return &Hub{
		clients:       make(map[int64]map[*Client]bool),
		uuidToID:      make(map[string]int64),
		register:      make(chan *Client),
		Unregister:    make(chan *Client),
		broadcast:     make(chan *WSMessage, 256),
		db:            db,
		instanceID:    uuid.New().String(),
		offlineTimers: make(map[int64]*time.Timer),
		ctx:           ctx,
		cancel:        cancel,
	}
}

func (h *Hub) Run() {
	if h.redisClient != nil {
		go h.listenRedisFanout()
	}

	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if _, ok := h.clients[client.UserID]; !ok {
				h.clients[client.UserID] = make(map[*Client]bool)
			}
			h.clients[client.UserID][client] = true
			if client.UserUUID != "" {
				h.uuidToID[client.UserUUID] = client.UserID
			}
			h.mu.Unlock()

			h.handleUserOnline(client)

			log.Printf("WebSocket client connected: user %d (total clients: %d)", client.UserID, h.ClientCount())

		case client := <-h.Unregister:
			h.mu.Lock()
			if clients, ok := h.clients[client.UserID]; ok {
				if _, exists := clients[client]; exists {
					delete(clients, client)
					close(client.Send)
					if len(clients) == 0 {
						delete(h.clients, client.UserID)
						if client.UserUUID != "" {
							delete(h.uuidToID, client.UserUUID)
						}
					}
				}
			}
			stillConnected := h.isUserConnectedLocked(client.UserID)
			h.mu.Unlock()

			if !stillConnected {
				h.scheduleOffline(client)
			}

			log.Printf("WebSocket client disconnected: user %d (total clients: %d)", client.UserID, h.ClientCount())

		case message := <-h.broadcast:
			h.dispatchMessage(message)

		case <-h.ctx.Done():
			return
		}
	}
}

func (h *Hub) Shutdown() {
	h.cancel()

	h.timersMu.Lock()
	for _, t := range h.offlineTimers {
		t.Stop()
	}
	h.offlineTimers = nil
	h.timersMu.Unlock()

	h.mu.Lock()
	defer h.mu.Unlock()

	for userID, clients := range h.clients {
		for client := range clients {
			client.Conn.Close()
			close(client.Send)
		}
		delete(h.clients, userID)
	}

	log.Println("WebSocket hub shut down")
}

func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	count := 0
	for _, clients := range h.clients {
		count += len(clients)
	}
	return count
}

// IsUserConnectedUUID reports whether a user with the given public UUID has a
// live connection, checking Redis presence first (so it works across instances)
// and falling back to this instance's local registry.
func (h *Hub) IsUserConnectedUUID(userUUID string) bool {
	if userUUID == "" {
		return false
	}

	if h.redisClient != nil {
		online, err := h.redisClient.IsUserOnline(context.Background(), userUUID)
		if err == nil {
			return online
		}
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	id, ok := h.uuidToID[userUUID]
	if !ok {
		return false
	}
	clients, ok := h.clients[id]
	return ok && len(clients) > 0
}

func (h *Hub) RegisterClient(client *Client) {
	h.register <- client
}

func (h *Hub) HandleMessage(client *Client, msg *WSMessage) {
	switch msg.Type {
	case TypePing:
		pong := &WSMessage{
			Type:      TypePong,
			SenderID:  0,
			Timestamp: msg.Timestamp,
		}
		client.SendMessage(pong)

	case TypeChatMessage, TypeGroupMessage:
		h.broadcast <- msg

	case TypeTyping:
		var payload TypingPayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			return
		}
		if payload.IsGroup {
			h.PublishBroadcastGroup(payload.TargetID, client.UserUUID, msg)
		} else {
			h.PublishToUser(payload.TargetID, msg)
		}

	default:
	}
}

func (h *Hub) handleUserOnline(client *Client) {
	// Cancel any pending offline timer
	h.timersMu.Lock()
	if t, ok := h.offlineTimers[client.UserID]; ok {
		t.Stop()
		delete(h.offlineTimers, client.UserID)
	}
	h.timersMu.Unlock()

	clientCount := 0
	h.mu.RLock()
	if c, ok := h.clients[client.UserID]; ok {
		clientCount = len(c)
	}
	h.mu.RUnlock()

	if clientCount == 1 {
		presencePayload, _ := json.Marshal(PresencePayload{
			UserID:   client.UserID,
			UserUUID: client.UserUUID,
			IsOnline: true,
		})
		h.PublishBroadcastAll(&WSMessage{
			Type:      TypePresenceUpdate,
			Payload:   presencePayload,
			Timestamp: time.Now(),
		})

		if h.redisClient != nil {
			h.redisClient.SetUserOnline(client.Context(), client.UserUUID)
		}

		log.Printf("User %d is now online", client.UserID)
	}
}

func (h *Hub) isUserConnectedLocked(userID int64) bool {
	if c, ok := h.clients[userID]; ok {
		return len(c) > 0
	}
	return false
}

func (h *Hub) scheduleOffline(client *Client) {
	userUUID := client.UserUUID
	userID := client.UserID

	h.timersMu.Lock()
	if t, ok := h.offlineTimers[userID]; ok {
		t.Stop()
	}
	h.offlineTimers[userID] = time.AfterFunc(offlineDebounce, func() {
		h.mu.RLock()
		stillConnected := h.isUserConnectedLocked(userID)
		h.mu.RUnlock()

		if stillConnected {
			return
		}

		presencePayload, _ := json.Marshal(PresencePayload{
			UserID:   userID,
			UserUUID: userUUID,
			IsOnline: false,
		})
		h.PublishBroadcastAll(&WSMessage{
			Type:      TypePresenceUpdate,
			Payload:   presencePayload,
			Timestamp: time.Now(),
		})

		if h.redisClient != nil {
			h.redisClient.SetUserOffline(context.Background(), userUUID)
		}

		h.timersMu.Lock()
		delete(h.offlineTimers, userID)
		h.timersMu.Unlock()

		log.Printf("User %d is now offline", userID)
	})
	h.timersMu.Unlock()
}

func (h *Hub) SendToUser(userID int64, msg *WSMessage) {
	h.mu.RLock()
	clients, ok := h.clients[userID]
	h.mu.RUnlock()

	if !ok {
		return
	}

	for client := range clients {
		client.SendMessage(msg)
	}
}

// SendToUserUUID delivers a message to the local connections of the user with
// the given public UUID. Relaying to other instances is the caller's job via
// the fan-out envelope.
func (h *Hub) SendToUserUUID(userUUID string, msg *WSMessage) {
	if userUUID == "" {
		return
	}

	h.mu.RLock()
	id, ok := h.uuidToID[userUUID]
	h.mu.RUnlock()

	if !ok {
		return
	}

	h.SendToUser(id, msg)
}

func (h *Hub) SendToUsers(userIDs []int64, msg *WSMessage) {
	for _, userID := range userIDs {
		h.SendToUser(userID, msg)
	}
}

// BroadcastGroup sends a message to all accepted members of a group except the
// sender (who is the origin of the message and already has it locally). Members
// are addressed by their public UUID; the sender is excluded by UUID too.
func (h *Hub) BroadcastGroup(groupID int64, excludeUUID string, msg *WSMessage) {
	members, err := h.db.GetGroupMembers(context.Background(), groupID)
	if err != nil {
		return
	}

	for _, member := range members {
		if member.Status != "accepted" || member.User == nil {
			continue
		}
		if member.User.UUID == excludeUUID {
			continue
		}
		h.SendToUserUUID(member.User.UUID, msg)
	}
}

func (h *Hub) BroadcastToAll(msg *WSMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, clients := range h.clients {
		for client := range clients {
			client.SendMessage(msg)
		}
	}
}

func (h *Hub) dispatchMessage(msg *WSMessage) {
	switch msg.Type {
	case TypeChatMessage:
		var payload ChatPayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			return
		}
		h.PublishToUser(payload.DMUserID, msg)
		h.PublishToUser(msg.SenderID, msg)

	case TypeGroupMessage:
		var payload GroupChatPayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			return
		}
		h.PublishBroadcastGroup(payload.GroupID, msg.SenderUUID, msg)

	case TypeNotification:
		var payload NotificationPayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			return
		}
		h.PublishToUser(payload.TargetID, msg)

	case TypePresenceUpdate:
		h.PublishBroadcastAll(msg)

	default:
		BroadcastMessage <- msg
	}
}

var BroadcastMessage = make(chan *WSMessage, 256)

func init() {
	go func() {
		for range BroadcastMessage {
		}
	}()
}

func (h *Hub) GetOnlineUsers(userUUIDs []string) ([]string, error) {
	if h.redisClient == nil {
		var online []string
		for _, id := range userUUIDs {
			if h.IsUserConnectedUUID(id) {
				online = append(online, id)
			}
		}
		return online, nil
	}

	return h.redisClient.GetOnlineUsers(context.Background(), userUUIDs)
}

func (h *Hub) HeartbeatUser(userUUID string) {
	if h.redisClient != nil {
		h.redisClient.SetUserOnline(context.Background(), userUUID)
	}
}

// ---- Cross-instance fan-out ----

// PublishToUser delivers a message to a user's local connections and fans it
// out to other backend instances.
func (h *Hub) PublishToUser(userID int64, msg *WSMessage) {
	h.SendToUser(userID, msg)
	h.publishFanout(FanoutEnvelope{Kind: "user", Target: userID, Message: *msg})
}

// PublishBroadcastAll delivers a message to every local connection and fans it
// out to other backend instances.
func (h *Hub) PublishBroadcastAll(msg *WSMessage) {
	h.BroadcastToAll(msg)
	h.publishFanout(FanoutEnvelope{Kind: "all", Message: *msg})
}

// PublishBroadcastGroup delivers a message to a group's local members (except
// excludeUUID) and fans it out to other backend instances.
func (h *Hub) PublishBroadcastGroup(groupID int64, excludeUUID string, msg *WSMessage) {
	h.BroadcastGroup(groupID, excludeUUID, msg)
	h.publishFanout(FanoutEnvelope{Kind: "group", GroupID: groupID, Exclude: excludeUUID, Message: *msg})
}

func (h *Hub) publishFanout(env FanoutEnvelope) {
	if h.redisClient == nil {
		return
	}
	env.Origin = h.instanceID
	if err := h.redisClient.Publish(context.Background(), cache.ChannelWSFanout, env); err != nil {
		log.Printf("ws fanout publish error: %v", err)
	}
}

// listenRedisFanout subscribes to the cross-instance channel and delivers
// messages produced by other instances to this instance's local clients.
func (h *Hub) listenRedisFanout() {
	ps := h.redisClient.Subscribe(context.Background(), cache.ChannelWSFanout)
	defer ps.Close()

	ch := ps.Channel()
	for {
		select {
		case msg, ok := <-ch:
			if !ok {
				return
			}
			var env FanoutEnvelope
			if err := json.Unmarshal([]byte(msg.Payload), &env); err != nil {
				continue
			}
			// Ignore our own echo — the origin instance already delivered locally.
			if env.Origin == h.instanceID {
				continue
			}

			m := env.Message
			switch env.Kind {
			case "user":
				h.SendToUser(env.Target, &m)
			case "all":
				h.BroadcastToAll(&m)
			case "group":
				h.BroadcastGroup(env.GroupID, env.Exclude, &m)
			}
		case <-h.ctx.Done():
			return
		}
	}
}
