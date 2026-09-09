package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	database "social-network/backend/db/sql"

	"github.com/google/uuid"
)

type FollowRequest struct {
	UserID string `json:"user_id"`
}

func FollowRequestHandler(w http.ResponseWriter, r *http.Request, db *database.DataBase) {
	if r.Method != http.MethodPost {
		RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	currentUserID, ok := GetUserIDFromContext(r)
	if !ok {
		RespondError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	var req FollowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	targetID, err := resolveUserID(r, db, req.UserID)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	if targetID == currentUserID {
		RespondError(w, http.StatusBadRequest, "Cannot follow yourself")
		return
	}

	targetUser, err := db.GetUserByID(r.Context(), targetID)
	if err != nil {
		RespondError(w, http.StatusNotFound, "User not found")
		return
	}

	if targetUser.IsPublic {
		if err := db.CreateFollowRequest(r.Context(), currentUserID, targetID); err != nil {
			RespondError(w, http.StatusInternalServerError, "Failed to send follow request")
			return
		}

		db.AcceptFollowRequest(r.Context(), currentUserID, targetID)

		sendNotification(db, targetID, currentUserID, NotifNewFollower, "started following you", &currentUserID, nil)

		RespondSuccess(w, http.StatusOK, "Now following user", nil)
		return
	}

	isFollowing, _ := db.CheckFollowing(r.Context(), currentUserID, targetID)
	if isFollowing {
		RespondError(w, http.StatusConflict, "Already following this user")
		return
	}

	if err := db.CreateFollowRequest(r.Context(), currentUserID, targetID); err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to send follow request")
		return
	}

	sendNotification(db, targetID, currentUserID, NotifFollowRequest, "sent you a follow request", &currentUserID, nil)

	RespondSuccess(w, http.StatusOK, "Follow request sent", nil)
}

func AcceptFollowHandler(w http.ResponseWriter, r *http.Request, db *database.DataBase) {
	if r.Method != http.MethodPost {
		RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	currentUserID, ok := GetUserIDFromContext(r)
	if !ok {
		RespondError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	var req FollowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	targetID, err := resolveUserID(r, db, req.UserID)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	if err := db.AcceptFollowRequest(r.Context(), targetID, currentUserID); err != nil {
		RespondError(w, http.StatusNotFound, "Follow request not found")
		return
	}

	follower, _ := db.GetUserByID(r.Context(), targetID)
	content := "accepted your follow request"
	if follower != nil {
		name := getDisplayName(follower)
		content = name + " accepted your follow request"
	}

	sendNotification(db, targetID, currentUserID, NotifFollowAccepted, content, &currentUserID, nil)

	RespondSuccess(w, http.StatusOK, "Follow request accepted", nil)
}

func DeclineFollowHandler(w http.ResponseWriter, r *http.Request, db *database.DataBase) {
	if r.Method != http.MethodPost {
		RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	currentUserID, ok := GetUserIDFromContext(r)
	if !ok {
		RespondError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	var req FollowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	targetID, err := resolveUserID(r, db, req.UserID)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	if err := db.DeclineFollowRequest(r.Context(), targetID, currentUserID); err != nil {
		RespondError(w, http.StatusNotFound, "Follow request not found")
		return
	}

	RespondSuccess(w, http.StatusOK, "Follow request declined", nil)
}

func UnfollowHandler(w http.ResponseWriter, r *http.Request, db *database.DataBase) {
	if r.Method != http.MethodPost {
		RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	currentUserID, ok := GetUserIDFromContext(r)
	if !ok {
		RespondError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	var req FollowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	targetID, err := resolveUserID(r, db, req.UserID)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	if err := db.RemoveFollow(r.Context(), currentUserID, targetID); err != nil {
		RespondError(w, http.StatusNotFound, "Not following this user")
		return
	}

	RespondSuccess(w, http.StatusOK, "Unfollowed successfuly", nil)
}

func GetFollowersHandler(w http.ResponseWriter, r *http.Request, db *database.DataBase) {
	if r.Method != http.MethodGet {
		RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userIDStr := r.URL.Query().Get("user_id")
	log.Printf("[GET /api/followers] raw user_id=%q", userIDStr)
	if userIDStr == "" {
		RespondError(w, http.StatusBadRequest, "user_id query parameter required")
		return
	}

	userID, err := resolveUserID(r, db, userIDStr)
	if err != nil {
		log.Printf("[GET /api/followers] resolveUserID failed: %v", err)
		RespondError(w, http.StatusBadRequest, "Invalid user_id")
		return
	}
	log.Printf("[GET /api/followers] resolved userID=%d", userID)

	followers, err := db.GetFollowersWithDM(r.Context(), userID)
	if err != nil {
		log.Printf("[GET /api/followers] GetFollowersWithDM error: %v", err)
		RespondError(w, http.StatusInternalServerError, "Failed to get followers")
		return
	}

	log.Printf("[GET /api/followers] got %d followers for user %d", len(followers), userID)

	if followers == nil {
		followers = []database.FollowerWithDM{}
	}

	for i := range followers {
		followers[i].User.IsOnline = GlobalHub != nil && GlobalHub.IsUserConnected(followers[i].User.ID)
	}

	RespondSuccess(w, http.StatusOK, "Followers retrieved", followers)
}

func GetFollowingHandler(w http.ResponseWriter, r *http.Request, db *database.DataBase) {
	if r.Method != http.MethodGet {
		RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		RespondError(w, http.StatusBadRequest, "user_id query parameter required")
		return
	}

	userID, err := resolveUserID(r, db, userIDStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "Invalid user_id")
		return
	}

	following, err := db.GetFollowingWithDM(r.Context(), userID)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to get following")
		return
	}

	if following == nil {
		following = []database.FollowerWithDM{}
	}

	for i := range following {
		following[i].User.IsOnline = GlobalHub != nil && GlobalHub.IsUserConnected(following[i].User.ID)
	}

	RespondSuccess(w, http.StatusOK, "Following retrieved", following)
}

func GetPendingFollowsHandler(w http.ResponseWriter, r *http.Request, db *database.DataBase) {
	if r.Method != http.MethodGet {
		RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	currentUserID, ok := GetUserIDFromContext(r)
	if !ok {
		RespondError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	pending, err := db.GetPendingFollowRequests(r.Context(), currentUserID)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to get pending requests")
		return
	}

	if pending == nil {
		pending = []database.User{}
	}

	RespondSuccess(w, http.StatusOK, "Pending requests retrieved", pending)
}

func getDisplayName(user *database.User) string {
	if user.Nickname != nil && *user.Nickname != "" {
		return *user.Nickname
	}
	return user.FirstName + " " + user.LastName
}

func generateUUID() string {
	return uuid.New().String()
}
