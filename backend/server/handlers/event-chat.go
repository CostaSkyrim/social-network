package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	database "social-network/backend/db/sql"
	ws "social-network/backend/server/websocket"

	"github.com/google/uuid"
)

type SendMessageRequest struct {
	Content string `json:"content"`
}

type DMListItem struct {
	ID            int64          `json:"id"`
	OtherUser     *database.User `json:"other_user"`
	LastMessage   *string        `json:"last_message,omitempty"`
	LastMessageAt interface{}    `json:"last_message_at,omitempty"`
	UnreadCount   int            `json:"unread_count"`
	CanSend       bool           `json:"can_send"`
}

func GetDMsHandler(w http.ResponseWriter, r *http.Request, db *database.DataBase) {
	if r.Method != http.MethodGet {
		RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := GetUserIDFromContext(r)
	if !ok {
		RespondError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	dms, err := db.GetAllDMs(r.Context(), userID)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to fetch conversations")
		return
	}

	result := make([]DMListItem, 0, len(dms))
	for _, dm := range dms {
		item := DMListItem{
			ID:            dm.ID,
			OtherUser:     dm.OtherUser,
			LastMessage:   dm.LastMessage,
			LastMessageAt: dm.LastMessageAt,
		}

		if dm.OtherUser != nil {
			unread, _ := db.GetUnreadCountForDM(r.Context(), dm.ID, userID)
			item.UnreadCount = unread
			item.OtherUser.IsOnline = GlobalHub != nil && GlobalHub.IsUserConnected(dm.OtherUser.ID)

			// History stays viewable, but sending requires at least one
			// direction of following.
			follows, _ := db.CheckFollowing(r.Context(), userID, dm.OtherUser.ID)
			if follows {
				item.CanSend = true
			} else {
				followedBy, _ := db.CheckFollowing(r.Context(), dm.OtherUser.ID, userID)
				item.CanSend = followedBy
			}
		}

		result = append(result, item)
	}

	RespondSuccess(w, http.StatusOK, "Conversations retrieved", result)
}

func GetMessagesHandler(w http.ResponseWriter, r *http.Request, db *database.DataBase) {
	if r.Method != http.MethodGet {
		RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := GetUserIDFromContext(r)
	if !ok {
		RespondError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	dmIDStr := r.PathValue("id")
	dmID, err := strconv.ParseInt(dmIDStr, 10, 64)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "Invalid DM ID")
		return
	}

	dm, err := db.GetDMByID(r.Context(), dmID)
	if err != nil || dm == nil {
		RespondError(w, http.StatusNotFound, "Conversation not found")
		return
	}

	// Conversation history remains visible to both participants; only sending
	// new messages requires an active follow relationship.
	if userID != dm.User1ID && userID != dm.User2ID {
		RespondError(w, http.StatusForbidden, "You are not part of this conversation")
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 50
	if limitStr != "" {
		if v, err := strconv.Atoi(limitStr); err == nil && v > 0 && v <= 100 {
			limit = v
		}
	}

	messages, err := db.GetMessages(r.Context(), dmID, limit)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to fetch messages")
		return
	}

	if messages == nil {
		messages = []*database.Message{}
	}

	for _, msg := range messages {
		db.MarkMessageRead(r.Context(), msg.ID, userID)
	}

	RespondSuccess(w, http.StatusOK, "Messages retrieved", messages)
}

func SendMessageHandler(w http.ResponseWriter, r *http.Request, db *database.DataBase) {
	if r.Method != http.MethodPost {
		RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	currentUserID, ok := GetUserIDFromContext(r)
	if !ok {
		RespondError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	targetUserIDStr := r.PathValue("id")
	targetUserID, err := resolveUserID(r, db, targetUserIDStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	if currentUserID == targetUserID {
		RespondError(w, http.StatusBadRequest, "Cannot message yourself")
		return
	}

	if _, err := db.GetUserByID(r.Context(), targetUserID); err != nil {
		RespondError(w, http.StatusNotFound, "User not found")
		return
	}

	// Access requires at least one direction of following at all times, so
	// unfollowing revokes the ability to keep messaging an existing chat.
	isFollowing, _ := db.CheckFollowing(r.Context(), currentUserID, targetUserID)
	if !isFollowing {
		isFollowedBy, _ := db.CheckFollowing(r.Context(), targetUserID, currentUserID)
		if !isFollowedBy {
			RespondError(w, http.StatusForbidden, "You must follow this user to send messages")
			return
		}
	}

	var req SendMessageRequest
	var imagePath *string

	if isMultipart(r) {
		if err := parseMultipartForm(r); err != nil {
			RespondError(w, http.StatusBadRequest, "Invalid multipart form")
			return
		}
		req.Content = r.FormValue("content")
		imgPath, ok := multipartImage(w, r)
		if !ok {
			return
		}
		imagePath = imgPath
	} else {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			RespondError(w, http.StatusBadRequest, "Invalid request body")
			return
		}
	}

	if strings.TrimSpace(req.Content) == "" && imagePath == nil {
		RespondError(w, http.StatusBadRequest, "Content cannot be empty")
		return
	}

	dm := &database.DirectMessage{
		User1ID: currentUserID,
		User2ID: targetUserID,
	}

	dmID, err := db.CreateOrGetDirectMessage(r.Context(), dm)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to create conversation")
		return
	}

	msg := &database.Message{
		UUID:            uuid.New().String(),
		SenderID:        currentUserID,
		DirectMessageID: &dmID,
		Content:         req.Content,
		ImagePath:       imagePath,
	}

	msgID, err := db.CreateMessage(r.Context(), msg)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to send message")
		return
	}

	db.UpdateDMTime(r.Context(), dmID)
	db.MarkMessageRead(r.Context(), msgID, currentUserID)

	now := time.Now().UTC()

	if GlobalHub != nil {
		sender, _ := db.GetUserByID(r.Context(), currentUserID)
		senderInfo := map[string]interface{}{
			"id":          sender.UUID,
			"first_name":  sender.FirstName,
			"last_name":   sender.LastName,
			"nickname":    sender.Nickname,
			"avatar_path": sender.AvatarPath,
		}

		chatPayload, _ := json.Marshal(map[string]interface{}{
			"message_id": msgID,
			"dm_id":      dmID,
			"content":    req.Content,
			"image_path": imagePath,
			"sender":     senderInfo,
			"created_at": now.Format(time.RFC3339),
		})

		GlobalHub.SendToUser(targetUserID, &ws.WSMessage{
			Type:      ws.TypeChatMessage,
			Payload:   chatPayload,
			SenderID:  currentUserID,
			Timestamp: now,
		})
	}

	RespondSuccess(w, http.StatusCreated, "Message sent", map[string]interface{}{
		"id":         msgID,
		"dm_id":      dmID,
		"content":    req.Content,
		"image_path": imagePath,
		"sender_id":  currentUserID,
	})
}

func GetUnreadMessageCountHandler(w http.ResponseWriter, r *http.Request, db *database.DataBase) {
	if r.Method != http.MethodGet {
		RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := GetUserIDFromContext(r)
	if !ok {
		RespondError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	count, err := db.GetUnreadMessageCount(r.Context(), userID)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to get unread count")
		return
	}

	RespondSuccess(w, http.StatusOK, "Unread count retrieved", map[string]int{
		"count": count,
	})
}
