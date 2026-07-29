package websocket

import (
	"encoding/json"
	"time"
)

type MessageType string

const (
	TypeChatMessage    MessageType = "chat_message"
	TypeGroupMessage   MessageType = "group_message"
	TypeNotification   MessageType = "notification"
	TypePresenceUpdate MessageType = "presence_update"
	TypeTyping         MessageType = "typing"
	TypeError          MessageType = "error"
	TypePing           MessageType = "ping"
	TypePong           MessageType = "pong"
)

type WSMessage struct {
	Type      MessageType     `json:"type"`
	Payload   json.RawMessage `json:"payload,omitempty"`
	SenderID  int64           `json:"sender_id,omitempty"`
	GroupID   *int64          `json:"group_id,omitempty"`
	Timestamp time.Time       `json:"timestamp"`
}

type ChatPayload struct {
	DMID    int64  `json:"dm_id"`
	Content string `json:"content"`
}

type GroupChatPayload struct {
	GroupID int64  `json:"group_id"`
	Content string `json:"content"`
}

type NotificationPayload struct {
	ID        int64  `json:"id"`
	Type      string `json:"type"`
	Content   string `json:"content"`
	RelatedID *int64 `json:"related_id,omitempty"`
}

type PresencePayload struct {
	UserID   int64 `json:"user_id"`
	IsOnline bool  `json:"is_online"`
}

type TypingPayload struct {
	TargetID int64  `json:"target_id"`
	IsGroup  bool   `json:"is_group"`
	IsTyping bool   `json:"is_typing"`
}

type ErrorPayload struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
