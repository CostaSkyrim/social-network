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
		if v, err := strconv.Atoi(limitStr); err == nil && v > 0 && v <= maxPageLimit() {
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

func GetFollowingPostsHandler(w http.ResponseWriter, r *http.Request, db *database.DataBase) {
	if r.Method != http.MethodGet {
		RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := GetUserIDFromContext(r)
	if !ok {
		RespondError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	limit := 10
	offset := 0
	if v, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && v > 0 && v <= maxPageLimit() {
		limit = v
	}
	if v, err := strconv.Atoi(r.URL.Query().Get("offset")); err == nil && v >= 0 {
		offset = v
	}

	posts, err := db.GetFollowingPosts(r.Context(), userID, limit, offset)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to fetch following posts")
		return
	}

	if posts == nil {
		posts = []*database.Post{}
	}

	RespondSuccess(w, http.StatusOK, "Following posts retrieved", posts)
}

func GetExplorePostsHandler(w http.ResponseWriter, r *http.Request, db *database.DataBase) {
	if r.Method != http.MethodGet {
		RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := GetUserIDFromContext(r)
	if !ok {
		RespondError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	limit := 10
	offset := 0
	if v, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && v > 0 && v <= maxPageLimit() {
		limit = v
	}
	if v, err := strconv.Atoi(r.URL.Query().Get("offset")); err == nil && v >= 0 {
		offset = v
	}

	posts, err := db.GetExplorePosts(r.Context(), userID, limit, offset)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to fetch explore posts")
		return
	}

	if posts == nil {
		posts = []*database.Post{}
	}

	RespondSuccess(w, http.StatusOK, "Explore posts retrieved", posts)
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

	can, err := db.CanViewPost(r.Context(), userID, postID)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to check post access")
		return
	}
	if !can {
		RespondError(w, http.StatusNotFound, "Post not found")
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

	// Only the author of a private post gets the list of allowed users.
	if post.AuthorID == userID && post.PrivacyLevel == "private" {
		if uuids, err := db.GetPostVisibleUserUUIDs(r.Context(), post.ID); err == nil {
			post.VisibleUserIDs = uuids
		}
	}

	RespondSuccess(w, http.StatusOK, "Post retrieved", post)
}
