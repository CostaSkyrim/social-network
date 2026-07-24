package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"social-network/backend/config"
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

	cfg := config.GetConfig()
	if cfg == nil {
		RespondError(w, http.StatusInternalServerError, "Server configuration error")
		return
	}
	limits := cfg.DatabaseConfiguration.Limits

	if req.Email == "" || req.Password == "" || req.FirstName == "" ||
		req.LastName == "" {
		RespondError(w, http.StatusBadRequest, "Required fields must be filled in")
		return
	}

	if !isValidEmail(req.Email) {
		RespondError(w, http.StatusBadRequest, "Invalid email format")
		return
	}

	if len(req.Password) < limits.MinPass {
		RespondError(w, http.StatusBadRequest,
			fmt.Sprintf("Password must be at least %d characters long", limits.MaxPass))
		return
	}

	firstName := strings.TrimSpace(req.FirstName)
	if len(firstName) < limits.MinFirstName {
		RespondError(w, http.StatusBadRequest,
			fmt.Sprintf("First name must be at least %d characters long", limits.MinFirstName))
		return
	}
	if len(firstName) > limits.MaxFirstName {
		RespondError(w, http.StatusBadRequest,
			fmt.Sprintf("First name must be at most %d characters long", limits.MaxFirstName))
		return
	}

	if req.Nickname != nil && *req.Nickname != "" {
		nickname := strings.TrimSpace(*req.Nickname)
		if len(nickname) < limits.MinUsername {
			RespondError(w, http.StatusBadRequest,
				fmt.Sprintf("Username must be at least %d characters long", limits.MinUsername))
			return
		}
		if len(nickname) > limits.MaxUsername {
			RespondError(w, http.StatusBadRequest,
				fmt.Sprintf("Username must be at most %d characters long", limits.MaxUsername))
			return
		}
	}

	if req.AboutMe != nil && len(*req.AboutMe) > limits.MaxBio {
		RespondError(w, http.StatusBadRequest,
			fmt.Sprintf("About me must be at most %d characters long", limits.MaxBio))
		return
	}

	if !req.DateOfBirth.IsZero() {
		age := calculateAge(req.DateOfBirth)
		if age < 13 {
			RespondError(w, http.StatusBadRequest, "You must be at least 13 years old to register")
			return
		}
		if age > 120 {
			RespondError(w, http.StatusBadRequest, "Invalid date of birth")
			return
		}
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

func isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return false
	}

	if len(email) > 254 {
		return false
	}

	if strings.Contains(email, "..") {
		return false
	}

	return true
}

func calculateAge(dob time.Time) int {
	now := time.Now()
	age := now.Year() - dob.Year()

	if now.Month() < dob.Month() || (now.Month() == dob.Month() && now.Day() < dob.Day()) {
		age--
	}

	return age
}
