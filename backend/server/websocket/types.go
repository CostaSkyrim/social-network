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

	// SenderUUID identifies the sender for server-side routing only. It is
	// deliberately never serialized, so the client-facing wire format (which
	// still carries the numeric sender_id) is unchanged.
	SenderUUID string `json:"-"`
}

// FanoutEnvelope wraps a WSMessage for cross-instance delivery over Redis
// pub/sub. Origin is the instance that produced the message, so the publisher
// can ignore its own echo (it already delivered locally). Target and Exclude
// are user UUIDs; GroupID is the internal group id.
type FanoutEnvelope struct {
	Origin  string    `json:"origin"`
	Kind    string    `json:"kind"` // "user" | "all" | "group"
	Target  int64     `json:"target,omitempty"`
	GroupID int64     `json:"group_id,omitempty"`
	Exclude string    `json:"exclude,omitempty"`
	Message WSMessage `json:"message"`
}

type ChatPayload struct {
	DMID     int64  `json:"dm_id"`
	DMUserID int64  `json:"dm_user_id"`
	Content  string `json:"content"`
}

type GroupChatPayload struct {
	GroupID int64  `json:"group_id"`
	Content string `json:"content"`
}

type NotificationPayload struct {
	ID           int64   `json:"id"`
	Type         string  `json:"type"`
	Content      string  `json:"content"`
	RelatedID    *int64  `json:"related_id,omitempty"`
	RelatedUUID  *string `json:"related_id_uuid,omitempty"`
	FromUserID   *int64  `json:"from_user_id,omitempty"`
	FromUserUUID *string `json:"from_user_uuid,omitempty"`
	TargetID     int64   `json:"-"`
	IsRead       bool    `json:"is_read"`
	CreatedAt    string  `json:"created_at"`
}

type PresencePayload struct {
	UserID   int64  `json:"user_id"`
	UserUUID string `json:"user_uuid"`
	IsOnline bool   `json:"is_online"`
}

type TypingPayload struct {
	TargetID int64 `json:"target_id"`
	IsGroup  bool  `json:"is_group"`
	IsTyping bool  `json:"is_typing"`
}

type ErrorPayload struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
