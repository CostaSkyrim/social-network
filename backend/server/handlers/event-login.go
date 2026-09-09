package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"social-network/backend/cache"
	database "social-network/backend/db/sql"
	ws "social-network/backend/server/websocket"

	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Message string        `json:"message"`
	User    *UserResponse `json:"user"`
}

type UserResponse struct {
	ID          string    `json:"id"`
	Email       string    `json:"email"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	Nickname    *string   `json:"nickname"`
	DateOfBirth time.Time `json:"date_of_birth"`
	AboutMe     *string   `json:"about_me"`
	AvatarPath  *string   `json:"avatar_path"`
	IsPublic    bool      `json:"is_public"`
}

func LoginHandler(w http.ResponseWriter, r *http.Request, db *database.DataBase) {
	if r.Method != http.MethodPost {
		RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" {
		RespondError(w, http.StatusUnauthorized, "Invalid email/nickname or password")
		return
	}

	user, err := db.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		user, err = db.GetUserByNickname(r.Context(), req.Email)
		if err != nil {
			RespondError(w, http.StatusUnauthorized, "Invalid email/nickname or password")
			return
		}
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		RespondError(w, http.StatusUnauthorized, "Invalid email/nickname or password")
		return
	}

	if err := CreateUserSession(w, r, db, getRedis(), user.ID); err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to create session")
		return
	}

	RespondSuccess(w, http.StatusOK, "Login successful", userToResponse(user))
}

func userToResponse(user *database.User) *UserResponse {
	return &UserResponse{
		ID:          user.UUID,
		Email:       user.Email,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		Nickname:    user.Nickname,
		DateOfBirth: user.DateOfBirth,
		AboutMe:     user.AboutMe,
		AvatarPath:  user.AvatarPath,
		IsPublic:    user.IsPublic,
	}
}

func CheckAuthHandler(w http.ResponseWriter, r *http.Request, db *database.DataBase) {
	if r.Method != http.MethodGet {
		RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	user, err := GetUserFromCookie(r, db, getRedis())
	if err != nil {
		RespondError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	RespondSuccess(w, http.StatusOK, "Authenticated", userToResponse(user))
}

func LogoutHandler(w http.ResponseWriter, r *http.Request, db *database.DataBase) {
	if r.Method != http.MethodPost {
		RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := GetUserIDFromContext(r)
	if ok && GlobalHub != nil {
		user, err := db.GetUserByID(r.Context(), userID)
		if err == nil {
			payload, _ := json.Marshal(map[string]interface{}{
				"user_id":   user.ID,
				"user_uuid": user.UUID,
				"is_online": false,
			})
			GlobalHub.BroadcastToAll(&ws.WSMessage{
				Type:      ws.TypePresenceUpdate,
				Payload:   payload,
				Timestamp: time.Now(),
			})
		}
	}

	cookie, err := r.Cookie("session_token")
	if err == nil {
		db.DeleteSession(r.Context(), cookie.Value)
		if rc := getRedis(); rc != nil {
			rc.Delete(r.Context(), cache.SessionKey(cookie.Value))
		}
	}

	ClearSessionCookie(w)

	RespondSuccess(w, http.StatusOK, "Logged out successfuly", nil)
}

func LogoutAllHandler(w http.ResponseWriter, r *http.Request, db *database.DataBase) {
	if r.Method != http.MethodPost {
		RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, authenticated := GetUserIDFromContext(r)
	if !authenticated {
		RespondError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	sessionIDs, err := db.GetUserSessionIDs(r.Context(), userID)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to logout")
		return
	}

	if err := db.DeleteAllUserSessions(r.Context(), userID); err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to logout")
		return
	}

	if rc := getRedis(); rc != nil {
		for _, sessionID := range sessionIDs {
			rc.Delete(r.Context(), cache.SessionKey(sessionID))
		}
	}

	ClearSessionCookie(w)

	RespondSuccess(w, http.StatusOK, "Logged out from all devices successfuly", nil)
}
