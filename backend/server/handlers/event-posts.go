package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	database "social-network/backend/db/sql"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

type CreatePostRequest struct {
	Content      string  `json:"content"`
	ImagePath    *string `json:"image_path,omitempty"`
	PrivacyLevel string  `json:"privacy_level"`
	GroupID      *string `json:"group_id,omitempty"`
}

// resolvePostID looks up a post by its UUID (from the URL path) and returns
// the numeric internal ID. Responds with an error and returns false on failure.
func resolvePostID(w http.ResponseWriter, r *http.Request, db *database.DataBase) (int64, bool) {
	postUUID := r.PathValue("id")
	if postUUID == "" {
		RespondError(w, http.StatusBadRequest, "Post ID required")
		return 0, false
	}

	post, err := db.GetPostByUUID(r.Context(), postUUID)
	if err != nil {
		RespondError(w, http.StatusNotFound, "Post not found")
		return 0, false
	}

	return post.ID, true
}

type PostResponse struct {
	ID           string  `json:"id"`
	AuthorID     int64   `json:"author_id"`
	GroupID      *int64  `json:"group_id,omitempty"`
	Content      string  `json:"content"`
	ImagePath    *string `json:"image_path,omitempty"`
	PrivacyLevel string  `json:"privacy_level"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
}

func CreatePostHandler(w http.ResponseWriter, r *http.Request, db *database.DataBase) {
	if r.Method != http.MethodPost {
		RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := GetUserIDFromContext(r)
	if !ok {
		RespondError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	var req CreatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if strings.TrimSpace(req.Content) == "" {
		RespondError(w, http.StatusBadRequest, "Content cannot be empty")
		return
	}

	if req.PrivacyLevel == "" {
		req.PrivacyLevel = "public"
	}

	var groupID *int64
	if req.GroupID != nil {
		group, err := db.GetGroupByUUID(r.Context(), *req.GroupID)
		if err != nil {
			RespondError(w, http.StatusNotFound, "Group not found")
			return
		}
		groupID = &group.ID
	}

	post := &database.Post{
		UUID:         uuid.New().String(),
		AuthorID:     userID,
		GroupID:      groupID,
		Content:      req.Content,
		ImagePath:    req.ImagePath,
		PrivacyLevel: req.PrivacyLevel,
	}

	id, err := db.CreatePost(r.Context(), post)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to create post")
		return
	}

	post.ID = id
	notifyFollowersOfNewPost(db, userID, post)

	RespondSuccess(w, http.StatusCreated, "Post created", map[string]any{
		"id":            post.UUID,
		"content":       post.Content,
		"image_path":    post.ImagePath,
		"privacy_level": post.PrivacyLevel,
		"group_id":      req.GroupID,
	})
}

func GetUserPostsHandler(w http.ResponseWriter, r *http.Request, db *database.DataBase) {
	if r.Method != http.MethodGet {
		RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := GetUserIDFromContext(r)
	if !ok {
		RespondError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 20
	offset := 0

	if limitStr != "" {
		if v, err := strconv.Atoi(limitStr); err == nil && v > 0 && v <= 50 {
			limit = v
		}
	}
	if offsetStr != "" {
		if v, err := strconv.Atoi(offsetStr); err == nil && v >= 0 {
			offset = v
		}
	}

	posts, err := db.GetUserPosts(r.Context(), userID, limit, offset)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to fetch posts")
		return
	}

	RespondSuccess(w, http.StatusOK, "Posts retrieved", posts)
}

func DeletePostHandler(w http.ResponseWriter, r *http.Request, db *database.DataBase) {
	if r.Method != http.MethodDelete {
		RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := GetUserIDFromContext(r)
	if !ok {
		RespondError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	postID, ok := resolvePostID(w, r, db)
	if !ok {
		return
	}

	if err := db.DeletePost(r.Context(), postID, userID); err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to delete post")
		return
	}

	RespondSuccess(w, http.StatusOK, "Post deleted", nil)
}

type EditPostRequest struct {
	Content      string  `json:"content"`
	ImagePath    *string `json:"image_path,omitempty"`
	PrivacyLevel string  `json:"privacy_level"`
}

func EditPostHandler(w http.ResponseWriter, r *http.Request, db *database.DataBase) {
	if r.Method != http.MethodPut {
		RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := GetUserIDFromContext(r)
	if !ok {
		RespondError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	postID, ok := resolvePostID(w, r, db)
	if !ok {
		return
	}

	var req EditPostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if strings.TrimSpace(req.Content) == "" {
		RespondError(w, http.StatusBadRequest, "Content cannot be empty")
		return
	}

	if req.PrivacyLevel == "" {
		req.PrivacyLevel = "public"
	}

	if err := db.UpdatePost(r.Context(), postID, userID, req.Content, req.ImagePath, req.PrivacyLevel); err != nil {
		RespondError(w, http.StatusNotFound, "Post not found or not authorized")
		return
	}

	RespondSuccess(w, http.StatusOK, "Post updated", nil)
}

func notifyFollowersOfNewPost(db *database.DataBase, authorID int64, post *database.Post) {
	followerIDs, err := db.GetFollowerIDs(context.Background(), authorID)
	if err != nil || len(followerIDs) == 0 {
		return
	}

	for _, followerID := range followerIDs {
		sendNotification(db, followerID, authorID, NotifNewPost, "shared a new post", &post.ID, nil)
	}

	fmt.Printf("Sent new_post notification to %d followers\n", len(followerIDs))
}
