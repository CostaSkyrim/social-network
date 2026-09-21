package websocket

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 4096
	// presencePeriod is how often a connected client refreshes its Redis
	// presence key. It must stay well below the presence TTL (30s) so idle
	// connections don't expire and appear offline.
	presencePeriod = 15 * time.Second
)

type Client struct {
	Hub      *Hub
	Conn     *websocket.Conn
	UserID   int64
	UserUUID string
	Send     chan []byte
	mu       sync.Mutex
}

func NewClient(hub *Hub, conn *websocket.Conn, userID int64) *Client {
	return &Client{
		Hub:    hub,
		Conn:   conn,
		UserID: userID,
		Send:   make(chan []byte, 256),
	}
}

func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket read error for user %d: %v", c.UserID, err)
			}
			break
		}

		var msg WSMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("WebSocket invalid message from user %d: %v", c.UserID, err)
			continue
		}

		msg.SenderID = c.UserID
		msg.Timestamp = time.Now()

		c.Hub.HandleMessage(c, &msg)
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	presenceTicker := time.NewTicker(presencePeriod)
	defer func() {
		ticker.Stop()
		presenceTicker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}

		case <-presenceTicker.C:
			// Keep the Redis presence key alive while the socket is open.
			c.Hub.HeartbeatUser(c.UserID)
		}
	}
}

func (c *Client) SendMessage(msg *WSMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	select {
	case c.Send <- data:
	default:
		c.Hub.Unregister <- c
	}
	return nil
}

func (c *Client) SendJSON(v interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
	return c.Conn.WriteJSON(v)
}

func (c *Client) Context() context.Context {
	return context.Background()
}
