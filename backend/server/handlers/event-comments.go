package handlers

import (
	"encoding/json"
	"net/http"
	database "social-network/backend/db/sql"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

type CreateCommentRequest struct {
	PostID          int64   `json:"post_id"`
	ParentCommentID *int64  `json:"parent_comment_id"`
	Content         string  `json:"content"`
	ImagePath       *string `json:"image_path,omitempty"`
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

	if req.PostID <= 0 {
		RespondError(w, http.StatusBadRequest, "Invalid post ID")
		return
	}

	comment := &database.Comment{
		UUID:            uuid.New().String(),
		PostID:          req.PostID,
		AuthorID:        userID,
		ParentCommentID: req.ParentCommentID,
		Content:         req.Content,
		ImagePath:       req.ImagePath,
	}

	_, err := db.CreateComment(r.Context(), comment)
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

	postIDstr := r.PathValue("id")
	postID, err := strconv.ParseInt(postIDstr, 10, 64)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "Invalid post ID")
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

	commentIDstr := r.PathValue("id")
	commentID, err := strconv.ParseInt(commentIDstr, 10, 64)

	if err != nil {
		RespondError(w, http.StatusBadRequest, "Invalid comment ID")
		return
	}

	if err := db.DeleteComment(r.Context(), commentID, userID); err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to delete comment")
		return
	}

	RespondSuccess(w, http.StatusOK, "Comment deleted", nil)
}
