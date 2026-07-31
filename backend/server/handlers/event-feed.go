package handlers

import (
	"database/sql"
	"net/http"
	database "social-network/backend/db/sql"
	"strconv"
)

func GetFeedHandler(w http.ResponseWriter, r *http.Request, db *database.DataBase) {
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

	limit := 10
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

	posts, err := db.GetFeed(r.Context(), userID, limit, offset)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to fetch feed")
		return
	}

	if posts == nil {
		posts = []*database.Post{}
	}

	RespondSuccess(w, http.StatusOK, "Feed retrieved", posts)
}

func GetPostHandler(w http.ResponseWriter, r *http.Request, db *database.DataBase) {
	if r.Method != http.MethodGet {
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

	post, err := db.GetPostWithAuthor(r.Context(), postID, userID)
	if err != nil {
		if err == sql.ErrNoRows || err.Error() == "post not found" {
			RespondError(w, http.StatusNotFound, "Post not found")
			return
		}
		RespondError(w, http.StatusInternalServerError, "Failed to fetch post")
		return
	}

	RespondSuccess(w, http.StatusOK, "Post retrieved", post)
}
