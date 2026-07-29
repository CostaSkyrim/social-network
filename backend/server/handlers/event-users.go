package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	database "social-network/backend/db/sql"
)

type updateProfileRequest struct {
	Nickname *string `json:"nickname"`
	AboutMe  *string `json:"about_me"`
	IsPublic *bool   `json:"is_public"`
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

	user, err := db.GetUserByUUID(r.Context(), userUUID)
	if err != nil {
		RespondError(w, http.StatusNotFound, "User not found")
		return
	}

	RespondSuccess(w, http.StatusOK, "User profile retrieved", user)
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

	targetUserIDStr := strconv.FormatInt(targetUser.ID, 10)
	currentUserIDStr := strconv.FormatInt(userID, 10)
	if targetUserIDStr != currentUserIDStr {
		RespondError(w, http.StatusForbidden, "You can only edit your own profile")
		return
	}

	var req updateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
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
