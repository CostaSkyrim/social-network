package handlers

import (
	"encoding/json"
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
	GroupID      *int64  `json:"group_id,omitempty"`
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

	post := &database.Post{
		UUID:         uuid.New().String(),
		AuthorID:     userID,
		GroupID:      req.GroupID,
		Content:      req.Content,
		ImagePath:    req.ImagePath,
		PrivacyLevel: req.PrivacyLevel,
	}

	_, err := db.CreatePost(r.Context(), post)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to create post")
		return
	}

	RespondSuccess(w, http.StatusCreated, "Post created", nil)
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

	postIDstr := r.PathValue("id")
	postID, err := strconv.ParseInt(postIDstr, 10, 64)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "Invalid post ID")
		return
	}

	if err := db.DeletePost(r.Context(), postID, userID); err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to delete post")
		return
	}

	RespondSuccess(w, http.StatusOK, "Post deleted", nil)
}
