package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	database "social-network/backend/db/sql"
)

type updateProfileRequest struct {
	Nickname *string `json:"nickname"`
	AboutMe  *string `json:"about_me"`
	IsPublic *bool   `json:"is_public"`
}

type profileResponse struct {
	User            *database.User   `json:"user"`
	FollowerCount   int              `json:"follower_count"`
	FollowingCount  int              `json:"following_count"`
	PostCount       int              `json:"post_count"`
	IsFollowing     bool             `json:"is_following"`
	IsFollowPending bool             `json:"is_follow_pending"`
	RecentPosts     []*database.Post `json:"recent_posts"`
}

func GetUserProfileHandler(w http.ResponseWriter, r *http.Request, db *database.DataBase) {
	if r.Method != http.MethodGet {
		RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userUUID := r.PathValue("id")
	if userUUID == "" {
		RespondError(w, http.StatusBadRequest, "User ID is required")
		return
	}

	targetUser, err := db.GetUserByUUID(r.Context(), userUUID)
	if err != nil {
		RespondError(w, http.StatusNotFound, "User not found")
		return
	}

	currentUserID, isAuthenticated := GetUserIDFromContext(r)
	isOwner := isAuthenticated && currentUserID == targetUser.ID

	if !targetUser.IsPublic && !isOwner {
		if !isAuthenticated {
			RespondError(w, http.StatusUnauthorized, "Private profile")
			return
		}
		isFollowing, _ := db.CheckFollowing(r.Context(), currentUserID, targetUser.ID)
		isFollowedBy, _ := db.CheckFollowing(r.Context(), targetUser.ID, currentUserID)
		if !isFollowing && !isFollowedBy {
			RespondError(w, http.StatusForbidden, "This profile is private")
			return
		}
	}

	targetUser.IsOnline = GlobalHub != nil && GlobalHub.IsUserConnected(targetUser.ID)

	followers, _ := db.GetFollowers(r.Context(), targetUser.ID)
	following, _ := db.GetFollowing(r.Context(), targetUser.ID)

	viewerID := int64(0)
	if isAuthenticated {
		viewerID = currentUserID
	}

	posts, _ := db.GetUserPostsForViewer(r.Context(), targetUser.ID, viewerID, 5, 0)
	if posts == nil {
		posts = []*database.Post{}
	}

	post_count, _ := db.CountUserPostsForViewer(r.Context(), targetUser.ID, viewerID)

	isFollowing := false
	isFollowPending := false
	if isAuthenticated {
		isFollowing, _ = db.CheckFollowing(r.Context(), currentUserID, targetUser.ID)
		if !isFollowing {
			pendingReqs, _ := db.GetPendingFollowRequests(r.Context(), targetUser.ID)
			for _, req := range pendingReqs {
				if req.ID == currentUserID {
					isFollowPending = true
					break
				}
			}
		}
	}

	RespondSuccess(w, http.StatusOK, "Profile retrieved", profileResponse{
		User:            targetUser,
		FollowerCount:   len(followers),
		FollowingCount:  len(following),
		PostCount:       post_count,
		IsFollowing:     isFollowing,
		IsFollowPending: isFollowPending,
		RecentPosts:     posts,
	})
}

// GetUserPostsForViewerHandler returns a user's posts (paginated) as visible to
// the requesting user, applying profile-privacy rules.
func GetUserPostsForViewerHandler(w http.ResponseWriter, r *http.Request, db *database.DataBase) {
	if r.Method != http.MethodGet {
		RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	viewerID, _ := GetUserIDFromContext(r)

	userUUID := r.PathValue("id")
	if userUUID == "" {
		RespondError(w, http.StatusBadRequest, "User ID is required")
		return
	}

	targetUser, err := db.GetUserByUUID(r.Context(), userUUID)
	if err != nil {
		RespondError(w, http.StatusNotFound, "User not found")
		return
	}

	isOwner := viewerID != 0 && viewerID == targetUser.ID

	// Same private-profile gate as the profile endpoint.
	if !targetUser.IsPublic && !isOwner {
		if viewerID == 0 {
			RespondError(w, http.StatusUnauthorized, "Private profile")
			return
		}
		isFollowing, _ := db.CheckFollowing(r.Context(), viewerID, targetUser.ID)
		isFollowedBy, _ := db.CheckFollowing(r.Context(), targetUser.ID, viewerID)
		if !isFollowing && !isFollowedBy {
			RespondError(w, http.StatusForbidden, "This profile is private")
			return
		}
	}

	limit := 5
	offset := 0
	if v, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && v > 0 && v <= 50 {
		limit = v
	}
	if v, err := strconv.Atoi(r.URL.Query().Get("offset")); err == nil && v >= 0 {
		offset = v
	}

	posts, err := db.GetUserPostsForViewer(r.Context(), targetUser.ID, viewerID, limit, offset)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to fetch posts")
		return
	}
	if posts == nil {
		posts = []*database.Post{}
	}

	RespondSuccess(w, http.StatusOK, "Posts retrieved", posts)
}

func UpdateUserProfileHandler(w http.ResponseWriter, r *http.Request, db *database.DataBase) {
	if r.Method != http.MethodPut {
		RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := GetUserIDFromContext(r)
	if !ok {
		RespondError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	userUUID := r.PathValue("id")
	if userUUID == "" {
		RespondError(w, http.StatusBadRequest, "User ID is required")
		return
	}

	targetUser, err := db.GetUserByUUID(r.Context(), userUUID)
	if err != nil {
		RespondError(w, http.StatusNotFound, "User not found")
		return
	}

	if targetUser.ID != userID {
		RespondError(w, http.StatusForbidden, "You can only edit your own profile")
		return
	}

	var req updateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Nickname != nil && *req.Nickname != "" {
		existing, _ := db.GetUserByNickname(r.Context(), *req.Nickname)
		if existing != nil && existing.ID != userID {
			RespondError(w, http.StatusConflict, "Nickname already taken")
			return
		}
	}

	user := &database.User{
		Nickname: req.Nickname,
		AboutMe:  req.AboutMe,
	}

	if err := db.UpdateUserProfile(r.Context(), userID, user); err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to update profile")
		return
	}

	if req.IsPublic != nil {
		if err := db.UpdateUserPrivacy(r.Context(), userID, *req.IsPublic); err != nil {
			RespondError(w, http.StatusInternalServerError, "Failed to update privacy setting")
			return
		}
	}

	updatedUser, err := db.GetUserByID(r.Context(), userID)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to fetch updated profile")
		return
	}

	RespondSuccess(w, http.StatusOK, "Profile updated", updatedUser)
}

func UpdateUserAvatarHandler(w http.ResponseWriter, r *http.Request, db *database.DataBase) {
	if r.Method != http.MethodPost {
		RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := GetUserIDFromContext(r)
	if !ok {
		RespondError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	userUUID := r.PathValue("id")
	if userUUID == "" {
		RespondError(w, http.StatusBadRequest, "User ID is required")
		return
	}

	targetUser, err := db.GetUserByUUID(r.Context(), userUUID)
	if err != nil {
		RespondError(w, http.StatusNotFound, "User not found")
		return
	}

	if targetUser.ID != userID {
		RespondError(w, http.StatusForbidden, "You can only change your own avatar")
		return
	}

	avatarPath, ok := SaveMultipartImage(w, r)
	if !ok {
		return
	}

	if err := db.UpdateUserAvatar(r.Context(), userID, avatarPath); err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to update avatar")
		return
	}

	RespondSuccess(w, http.StatusOK, "Avatar updated", map[string]interface{}{
		"avatar_path": avatarPath,
	})
}

func SearchUsersHandler(w http.ResponseWriter, r *http.Request, db *database.DataBase) {
	if r.Method != http.MethodGet {
		RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := GetUserIDFromContext(r)
	if !ok {
		RespondError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if query == "" {
		RespondSuccess(w, http.StatusOK, "Search results", []map[string]interface{}{})
		return
	}

	limit := 20
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if v, err := strconv.Atoi(limitStr); err == nil && v > 0 && v <= 50 {
			limit = v
		}
	}

	// Escape LIKE wildcards so user input is matched literally.
	escaped := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(query)
	pattern := "%" + escaped + "%"

	results, err := db.SearchUsers(r.Context(), userID, pattern, limit)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to search users")
		return
	}

	type searchUser struct {
		ID              string  `json:"id"`
		FirstName       string  `json:"first_name"`
		LastName        string  `json:"last_name"`
		Nickname        *string `json:"nickname"`
		AvatarPath      *string `json:"avatar_path,omitempty"`
		IsPublic        bool    `json:"is_public"`
		IsOnline        bool    `json:"is_online"`
		IsFollowing     bool    `json:"is_following"`
		IsFollowPending bool    `json:"is_follow_pending"`
	}

	payload := make([]searchUser, 0, len(results))
	for _, res := range results {
		payload = append(payload, searchUser{
			ID:              res.UUID,
			FirstName:       res.FirstName,
			LastName:        res.LastName,
			Nickname:        res.Nickname,
			AvatarPath:      res.AvatarPath,
			IsPublic:        res.IsPublic,
			IsOnline:        GlobalHub != nil && GlobalHub.IsUserConnected(res.ID),
			IsFollowing:     res.IsFollowing,
			IsFollowPending: res.IsFollowPending,
		})
	}

	RespondSuccess(w, http.StatusOK, "Search results", payload)
}
