package handlers

import (
	"encoding/json"
	"net/http"

	database "social-network/backend/db/sql"
)

type updateProfileRequest struct {
	Nickname *string `json:"nickname"`
	AboutMe  *string `json:"about_me"`
	IsPublic *bool   `json:"is_public"`
}

type profileResponse struct {
	User             *database.User  `json:"user"`
	FollowerCount    int             `json:"follower_count"`
	FollowingCount   int             `json:"following_count"`
	PostCount        int             `json:"post_count"`
	IsFollowing      bool            `json:"is_following"`
	IsFollowPending  bool            `json:"is_follow_pending"`
	RecentPosts      []*database.Post `json:"recent_posts"`
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

	posts, _ := db.GetUserPosts(r.Context(), targetUser.ID, 5, 0)
	if posts == nil {
		posts = []*database.Post{}
	}

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
		PostCount:       len(posts),
		IsFollowing:     isFollowing,
		IsFollowPending: isFollowPending,
		RecentPosts:     posts,
	})
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
