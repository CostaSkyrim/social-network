package handlers

import (
	"encoding/json"
	"net/http"
	database "social-network/backend/db/sql"
	"strings"

	"github.com/google/uuid"
)

type CreateCommentRequest struct {
	PostID          string  `json:"post_id"`
	ParentCommentID *string `json:"parent_comment_id"`
	Content         string  `json:"content"`
	ImagePath       *string `json:"image_path,omitempty"`
}

// resolveCommentID looks up a comment by its UUID (from the URL path) and
// returns the numeric internal ID. Responds with an error and returns false on failure.
func resolveCommentID(w http.ResponseWriter, r *http.Request, db *database.DataBase) (int64, bool) {
	commentUUID := r.PathValue("id")
	if commentUUID == "" {
		RespondError(w, http.StatusBadRequest, "Comment ID required")
		return 0, false
	}

	comment, err := db.GetCommentByUUID(r.Context(), commentUUID)
	if err != nil {
		RespondError(w, http.StatusNotFound, "Comment not found")
		return 0, false
	}

	return comment.ID, true
}

func CreateCommentHandler(w http.ResponseWriter, r *http.Request, db *database.DataBase) {
	if r.Method != http.MethodPost {
		RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := GetUserIDFromContext(r)
	if !ok {
		RespondError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	var req CreateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if strings.TrimSpace(req.Content) == "" {
		RespondError(w, http.StatusBadRequest, "Content cannot be empty")
		return
	}

	post, err := db.GetPostByUUID(r.Context(), req.PostID)
	if err != nil {
		RespondError(w, http.StatusNotFound, "Post not found")
		return
	}

	var parentCommentID *int64
	if req.ParentCommentID != nil {
		parent, err := db.GetCommentByUUID(r.Context(), *req.ParentCommentID)
		if err != nil {
			RespondError(w, http.StatusNotFound, "Parent comment not found")
			return
		}
		parentCommentID = &parent.ID
	}

	comment := &database.Comment{
		UUID:            uuid.New().String(),
		PostID:          post.ID,
		AuthorID:        userID,
		ParentCommentID: parentCommentID,
		Content:         req.Content,
		ImagePath:       req.ImagePath,
	}

	_, err = db.CreateComment(r.Context(), comment)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to create comment")
		return
	}

	RespondSuccess(w, http.StatusCreated, "Comment created", nil)
}

func GetPostCommentsHandler(w http.ResponseWriter, r *http.Request, db *database.DataBase) {
	if r.Method != http.MethodGet {
		RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	postID, ok := resolvePostID(w, r, db)
	if !ok {
		return
	}

	comments, err := db.GetPostComments(r.Context(), postID)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to fetch comments")
		return
	}

	RespondSuccess(w, http.StatusOK, "Comments retrieved", comments)
}

func DeleteCommentHandler(w http.ResponseWriter, r *http.Request, db *database.DataBase) {
	if r.Method != http.MethodDelete {
		RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := GetUserIDFromContext(r)
	if !ok {
		RespondError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	commentID, ok := resolveCommentID(w, r, db)
	if !ok {
		return
	}

	if err := db.DeleteComment(r.Context(), commentID, userID); err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to delete comment")
		return
	}

	RespondSuccess(w, http.StatusOK, "Comment deleted", nil)
}

type EditCommentRequest struct {
	Content   string  `json:"content"`
	ImagePath *string `json:"image_path,omitempty"`
}

func EditCommentHandler(w http.ResponseWriter, r *http.Request, db *database.DataBase) {
	if r.Method != http.MethodPut {
		RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := GetUserIDFromContext(r)
	if !ok {
		RespondError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	commentID, ok := resolveCommentID(w, r, db)
	if !ok {
		return
	}

	var req EditCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if strings.TrimSpace(req.Content) == "" {
		RespondError(w, http.StatusBadRequest, "Content cannot be empty")
		return
	}

	if err := db.UpdateComment(r.Context(), commentID, userID, req.Content, req.ImagePath); err != nil {
		RespondError(w, http.StatusNotFound, "Comment not found or not authorized")
		return
	}

	RespondSuccess(w, http.StatusOK, "Comment updated", nil)
}
