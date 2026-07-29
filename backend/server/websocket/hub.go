package websocket

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	"social-network/backend/cache"
	database "social-network/backend/db/sql"
)

type Hub struct {
	clients    map[int64]map[*Client]bool
	register   chan *Client
	Unregister chan *Client
	broadcast  chan *WSMessage
	mu         sync.RWMutex

	db          *database.DataBase
	redisClient *cache.RedisClient

	ctx    context.Context
	cancel context.CancelFunc
}

func NewHub(db *database.DataBase, redisClient *cache.RedisClient) *Hub {
	ctx, cancel := context.WithCancel(context.Background())

	return &Hub{
		clients:     make(map[int64]map[*Client]bool),
		register:    make(chan *Client),
		Unregister:  make(chan *Client),
		broadcast:   make(chan *WSMessage, 256),
		db:          db,
		redisClient: redisClient,
		ctx:         ctx,
		cancel:      cancel,
	}
}

func NewHubForTest(db *database.DataBase) *Hub {
	ctx, cancel := context.WithCancel(context.Background())

	return &Hub{
		clients:    make(map[int64]map[*Client]bool),
		register:   make(chan *Client),
		Unregister: make(chan *Client),
		broadcast:  make(chan *WSMessage, 256),
		db:         db,
		ctx:        ctx,
		cancel:     cancel,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if _, ok := h.clients[client.UserID]; !ok {
				h.clients[client.UserID] = make(map[*Client]bool)
			}
			h.clients[client.UserID][client] = true
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
					}
				}
			}
			h.mu.Unlock()

			h.handleUserOffline(client)

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

func (h *Hub) IsUserConnected(userID int64) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	clients, ok := h.clients[userID]
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
			members, err := h.db.GetGroupMembers(client.Context(), payload.TargetID)
			if err != nil {
				return
			}
			for _, member := range members {
				if member.UserID != client.UserID && member.Status == "accepted" {
					h.SendToUser(member.UserID, msg)
				}
			}
		} else {
			h.SendToUser(payload.TargetID, msg)
		}

	default:
	}
}

func (h *Hub) handleUserOnline(client *Client) {
	clientCount := 0
	h.mu.RLock()
	if c, ok := h.clients[client.UserID]; ok {
		clientCount = len(c)
	}
	h.mu.RUnlock()

	if clientCount == 1 {
		if h.redisClient != nil {
			h.redisClient.SetUserOnline(client.Context(), client.UserID)

			presencePayload, _ := json.Marshal(PresencePayload{
				UserID:   client.UserID,
				IsOnline: true,
			})
			h.redisClient.Publish(client.Context(), cache.ChannelUserOnline, presencePayload)
		}

		log.Printf("User %d is now online", client.UserID)
	}
}

func (h *Hub) handleUserOffline(client *Client) {
	stillConnected := false
	h.mu.RLock()
	if c, ok := h.clients[client.UserID]; ok {
		stillConnected = len(c) > 0
	}
	h.mu.RUnlock()

	if !stillConnected {
		if h.redisClient != nil {
			h.redisClient.SetUserOffline(client.Context(), client.UserID)

			presencePayload, _ := json.Marshal(PresencePayload{
				UserID:   client.UserID,
				IsOnline: false,
			})
			h.redisClient.Publish(client.Context(), cache.ChannelUserOffline, presencePayload)
		}

		log.Printf("User %d is now offline", client.UserID)
	}
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

func (h *Hub) SendToUsers(userIDs []int64, msg *WSMessage) {
	for _, userID := range userIDs {
		h.SendToUser(userID, msg)
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
		h.SendToUser(payload.DMID, msg)
		h.SendToUser(msg.SenderID, msg)

	case TypeGroupMessage:
		var payload GroupChatPayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			return
		}
		members, err := h.db.GetGroupMembers(context.Background(), payload.GroupID)
		if err != nil {
			return
		}
		for _, member := range members {
			if member.UserID != msg.SenderID && member.Status == "accepted" {
				h.SendToUser(member.UserID, msg)
			}
		}

	case TypeNotification:
		h.BroadcastToAll(msg)

	case TypePresenceUpdate:
		h.BroadcastToAll(msg)

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

func (h *Hub) GetOnlineUsers(userIDs []int64) ([]int64, error) {
	if h.redisClient == nil {
		var online []int64
		for _, id := range userIDs {
			if h.IsUserConnected(id) {
				online = append(online, id)
			}
		}
		return online, nil
	}

	return h.redisClient.GetOnlineUsers(context.Background(), userIDs)
}

func (h *Hub) HeartbeatUser(userID int64) {
	if h.redisClient != nil {
		h.redisClient.SetUserOnline(context.Background(), userID)
	}
}
