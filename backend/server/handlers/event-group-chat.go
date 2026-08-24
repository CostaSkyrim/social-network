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

// isGroupMember checks whether the user is an accepted member (or creator) of
// the group and writes a 403 error response if not.
func isGroupMember(w http.ResponseWriter, r *http.Request, db *database.DataBase, groupID, userID int64) bool {
	group, err := db.GetGroup(r.Context(), groupID)
	if err != nil {
		RespondError(w, http.StatusNotFound, "Group not found")
		return false
	}
	if group.CreatorID == userID {
		return true
	}

	members, err := db.GetGroupMembers(r.Context(), groupID)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to check membership")
		return false
	}
	for _, m := range members {
		if m.UserID == userID && m.Status == "accepted" {
			return true
		}
	}

	RespondError(w, http.StatusForbidden, "You must be a member of this group to access its chat")
	return false
}

// GetGroupMessagesHandler returns the messages for a group chat. Only accepted
// members (or the creator) can read them.
func GetGroupMessagesHandler(w http.ResponseWriter, r *http.Request, db *database.DataBase) {
	if r.Method != http.MethodGet {
		RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := GetUserIDFromContext(r)
	if !ok {
		RespondError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	groupID, ok := resolveGroupID(w, r, db)
	if !ok {
		return
	}

	if !isGroupMember(w, r, db, groupID, userID) {
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 50
	if limitStr != "" {
		if v, err := strconv.Atoi(limitStr); err == nil && v > 0 && v <= 100 {
			limit = v
		}
	}

	messages, err := db.GetGroupMessages(r.Context(), groupID, limit)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to fetch messages")
		return
	}

	if messages == nil {
		messages = []*database.Message{}
	}

	RespondSuccess(w, http.StatusOK, "Messages retrieved", messages)
}

// SendGroupMessageHandler sends a message to a group chat. Only accepted
// members (or the creator) can send, and the message is broadcast to all other
// members over the WebSocket hub.
func SendGroupMessageHandler(w http.ResponseWriter, r *http.Request, db *database.DataBase) {
	if r.Method != http.MethodPost {
		RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	currentUserID, ok := GetUserIDFromContext(r)
	if !ok {
		RespondError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	groupID, ok := resolveGroupID(w, r, db)
	if !ok {
		return
	}

	if !isGroupMember(w, r, db, groupID, currentUserID) {
		return
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

	msg := &database.Message{
		UUID:      uuid.New().String(),
		SenderID:  currentUserID,
		GroupID:   &groupID,
		Content:   req.Content,
		ImagePath: imagePath,
	}

	msgID, err := db.CreateMessage(r.Context(), msg)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to send message")
		return
	}

	db.UpdateGroupLastMessageTime(r.Context(), groupID)

	now := time.Now().UTC()

	if GlobalHub != nil {
		sender, _ := db.GetUserByID(r.Context(), currentUserID)
		group, _ := db.GetGroup(r.Context(), groupID)
		var groupUUID string
		if group != nil {
			groupUUID = group.UUID
		}

		senderInfo := map[string]interface{}{
			"id":          sender.UUID,
			"first_name":  sender.FirstName,
			"last_name":   sender.LastName,
			"nickname":    sender.Nickname,
			"avatar_path": sender.AvatarPath,
		}

		groupPayload, _ := json.Marshal(map[string]interface{}{
			"message_id": msgID,
			"group_id":   groupID,
			"group_uuid": groupUUID,
			"content":    req.Content,
			"image_path": imagePath,
			"sender":     senderInfo,
			"created_at": now.Format(time.RFC3339),
		})

		GlobalHub.BroadcastGroup(groupID, currentUserID, &ws.WSMessage{
			Type:      ws.TypeGroupMessage,
			Payload:   groupPayload,
			SenderID:  currentUserID,
			Timestamp: now,
		})
	}

	RespondSuccess(w, http.StatusCreated, "Message sent", map[string]interface{}{
		"id":         msgID,
		"group_id":   groupID,
		"content":    req.Content,
		"image_path": imagePath,
		"sender_id":  currentUserID,
	})
}
