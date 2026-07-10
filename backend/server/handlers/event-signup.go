package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	database "social-network/backend/db/sql"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type SignupRequest struct {
	Email       string    `json:"email"`
	Password    string    `json:"password"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	DateOfBirth time.Time `json:"date_of_birth"`
	Nickname    *string   `json:"nickname"`
	AboutMe     *string   `json:"about_me"`
}

func SignupHandler(w http.ResponseWriter, r *http.Request, db *database.DataBase) {
	if r.Method != http.MethodPost {
		RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" || req.FirstName == "" ||
		req.LastName == "" {
		RespondError(w, http.StatusBadRequest, "Required fields must be filled in")
		return
	}

	if len(req.Password) < 8 {
		RespondError(w, http.StatusBadRequest, "Password must be at least 8 characters long")
		return
	}

	_, err := db.GetUserByEmail(r.Context(), req.Email)
	if err == nil {
		RespondError(w, http.StatusConflict, "Email already registered")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to process registration")
		return
	}

	user := &database.User{
		UUID:         uuid.New().String(),
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Nickname:     req.Nickname,
		DateOfBirth:  req.DateOfBirth,
		AboutMe:      req.AboutMe,
		IsPublic:     true,
	}

	userID, err := db.AddUser(r.Context(), user)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to create account")
		return
	}

	if err := CreateUserSession(w, r, db, userID); err != nil {
		RespondError(w, http.StatusInternalServerError, "Account created but failed to login")
		return
	}

	user.ID = userID
	RespondSuccess(w, http.StatusCreated, "Registration successful", userToResponse(user))
}
