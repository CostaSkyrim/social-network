package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	database "social-network/backend/db/sql"
	"social-network/backend/config"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

type CreatePostRequest struct {
	Content        string    `json:"content"`
	ImagePath      *string   `json:"image_path,omitempty"`
	PrivacyLevel   string    `json:"privacy_level"`
	GroupID        *string   `json:"group_id,omitempty"`
	VisibleUserIDs *[]string `json:"visible_user_ids,omitempty"`
}

// resolveVisibleUserIDs converts a list of user UUIDs into internal user IDs.
func resolveVisibleUserIDs(ctx context.Context, db *database.DataBase, uuids []string) ([]int64, error) {
	ids := make([]int64, 0, len(uuids))
	for _, u := range uuids {
		u = strings.TrimSpace(u)
		if u == "" {
			continue
		}
		user, err := db.GetUserByUUID(ctx, u)
		if err != nil {
			return nil, fmt.Errorf("user not found: %s", u)
		}
		ids = append(ids, user.ID)
	}
	return ids, nil
}

// splitCommaList flattens form values that may contain comma-separated IDs.
func splitCommaList(values []string) []string {
	out := []string{}
	for _, v := range values {
		for _, part := range strings.Split(v, ",") {
			part = strings.TrimSpace(part)
			if part != "" {
				out = append(out, part)
			}
		}
	}
	return out
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
	if isMultipart(r) {
		if err := parseMultipartForm(r); err != nil {
			RespondError(w, http.StatusBadRequest, "Invalid multipart form")
			return
		}
		req.Content = r.FormValue("content")
		req.PrivacyLevel = r.FormValue("privacy_level")
		if g := r.FormValue("group_id"); g != "" {
			req.GroupID = &g
		}
		if vals, ok := r.Form["visible_user_ids"]; ok {
			list := splitCommaList(vals)
			req.VisibleUserIDs = &list
		}
		imgPath, ok := multipartImage(w, r)
		if !ok {
			return
		}
		req.ImagePath = imgPath
	} else {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			RespondError(w, http.StatusBadRequest, "Invalid request body")
			return
		}
	}

	if strings.TrimSpace(req.Content) == "" && req.ImagePath == nil {
		RespondError(w, http.StatusBadRequest, "Content cannot be empty")
		return
	}

	if cfg := config.GetConfig(); cfg != nil {
		limits := cfg.DatabaseConfiguration.Limits
		if content := strings.TrimSpace(req.Content); limits.MinBody > 0 && content != "" && len(content) < limits.MinBody {
			RespondError(w, http.StatusBadRequest,
				fmt.Sprintf("Post body must be at least %d characters long", limits.MinBody))
			return
		}
		if limits.MaxPostBody > 0 && len(req.Content) > limits.MaxPostBody {
			RespondError(w, http.StatusBadRequest,
				fmt.Sprintf("Post body must be at most %d characters long", limits.MaxPostBody))
			return
		}
	}

	var groupID *int64
	if req.GroupID != nil {
		group, err := db.GetGroupByUUID(r.Context(), *req.GroupID)
		if err != nil {
			RespondError(w, http.StatusNotFound, "Group not found")
			return
		}
		if !requireGroupMember(w, r, db, group.ID, userID) {
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

	if req.PrivacyLevel == "private" && req.VisibleUserIDs != nil {
		ids, err := resolveVisibleUserIDs(r.Context(), db, *req.VisibleUserIDs)
		if err != nil {
			RespondError(w, http.StatusBadRequest, "Invalid visible user")
			return
		}
		if err := db.SetPostVisibility(r.Context(), id, ids); err != nil {
			RespondError(w, http.StatusInternalServerError, "Failed to set post visibility")
			return
		}
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
		if v, err := strconv.Atoi(limitStr); err == nil && v > 0 && v <= maxPageLimit() {
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

func GetGroupPostsHandler(w http.ResponseWriter, r *http.Request, db *database.DataBase) {
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

	if !requireGroupMember(w, r, db, groupID, userID) {
		return
	}

	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 20
	offset := 0

	if limitStr != "" {
		if v, err := strconv.Atoi(limitStr); err == nil && v > 0 && v <= maxPageLimit() {
			limit = v
		}
	}
	if offsetStr != "" {
		if v, err := strconv.Atoi(offsetStr); err == nil && v >= 0 {
			offset = v
		}
	}

	posts, err := db.GetGroupPosts(r.Context(), groupID, limit, offset)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to fetch group posts")
		return
	}

	RespondSuccess(w, http.StatusOK, "Group posts retrieved", posts)
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
	Content        string    `json:"content"`
	ImagePath      *string   `json:"image_path,omitempty"`
	PrivacyLevel   string    `json:"privacy_level"`
	VisibleUserIDs *[]string `json:"visible_user_ids,omitempty"`
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

	post, err := db.GetPostByUUID(r.Context(), r.PathValue("id"))
	if err != nil {
		RespondError(w, http.StatusNotFound, "Post not found")
		return
	}

	var req EditPostRequest
	var imagePath *string

	if isMultipart(r) {
		if err := parseMultipartForm(r); err != nil {
			RespondError(w, http.StatusBadRequest, "Invalid multipart form")
			return
		}
		req.Content = r.FormValue("content")
		req.PrivacyLevel = r.FormValue("privacy_level")
		if vals, ok := r.Form["visible_user_ids"]; ok {
			list := splitCommaList(vals)
			req.VisibleUserIDs = &list
		}
		imgPath, ok := multipartImage(w, r)
		if !ok {
			return
		}
		imagePath = resolveImageUpdate(post.ImagePath, imgPath, r)
	} else {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			RespondError(w, http.StatusBadRequest, "Invalid request body")
			return
		}
		imagePath = req.ImagePath
	}

	if strings.TrimSpace(req.Content) == "" && imagePath == nil {
		RespondError(w, http.StatusBadRequest, "Content cannot be empty")
		return
	}

	if req.PrivacyLevel == "" {
		req.PrivacyLevel = "public"
	}

	if err := db.UpdatePost(r.Context(), postID, userID, req.Content, imagePath, req.PrivacyLevel); err != nil {
		RespondError(w, http.StatusNotFound, "Post not found or not authorized")
		return
	}

	// Keep post visibility in sync with the privacy level.
	switch {
	case req.PrivacyLevel != "private":
		if err := db.SetPostVisibility(r.Context(), postID, nil); err != nil {
			RespondError(w, http.StatusInternalServerError, "Failed to clear post visibility")
			return
		}
	case req.VisibleUserIDs != nil:
		ids, err := resolveVisibleUserIDs(r.Context(), db, *req.VisibleUserIDs)
		if err != nil {
			RespondError(w, http.StatusBadRequest, "Invalid visible user")
			return
		}
		if err := db.SetPostVisibility(r.Context(), postID, ids); err != nil {
			RespondError(w, http.StatusInternalServerError, "Failed to set post visibility")
			return
		}
	}

	RespondSuccess(w, http.StatusOK, "Post updated", nil)
}

func notifyFollowersOfNewPost(db *database.DataBase, authorID int64, post *database.Post) {
	followerIDs, err := db.GetFollowerIDs(context.Background(), authorID)
	if err != nil || len(followerIDs) == 0 {
		return
	}

	authorName := "A user"
	if author, err := db.GetUserByID(context.Background(), authorID); err == nil && author != nil {
		authorName = getDisplayName(author)
	}

	for _, followerID := range followerIDs {
		sendNotification(db, followerID, authorID, NotifNewPost, authorName+" shared a new post", &post.ID, &post.UUID)
	}

	fmt.Printf("Sent new_post notification to %d followers\n", len(followerIDs))
}
