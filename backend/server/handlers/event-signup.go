package handlers

import (
	"encoding/json"
	"fmt"
	"math/rand"
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
	Nickname    string    `json:"nickname"`
	AboutMe     *string   `json:"about_me"`
}

func SignupHandler(w http.ResponseWriter, r *http.Request, db *database.DataBase) {
	if r.Method != http.MethodPost {
		RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req SignupRequest
	var avatarPath *string

	if isMultipart(r) {
		if err := parseMultipartForm(r); err != nil {
			RespondError(w, http.StatusBadRequest, "Invalid multipart form")
			return
		}
		req.Email = r.FormValue("email")
		req.Password = r.FormValue("password")
		req.FirstName = r.FormValue("first_name")
		req.LastName = r.FormValue("last_name")
		req.Nickname = r.FormValue("nickname")
		if about := r.FormValue("about_me"); about != "" {
			req.AboutMe = &about
		}
		if dob := r.FormValue("date_of_birth"); dob != "" {
			if parsed, err := time.Parse(time.RFC3339, dob); err == nil {
				req.DateOfBirth = parsed
			}
		}
		if img, ok := multipartImage(w, r); ok && img != nil {
			avatarPath = img
		}
	} else {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			RespondError(w, http.StatusBadRequest, "Invalid request body")
			return
		}
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

	nickname := ""
	if req.Nickname != "" {
		nickname = sanitizeNickname(req.Nickname, limits.MaxUsername)
		if len(nickname) < limits.MinUsername {
			nickname = ""
		}
	}
	if nickname == "" {
		nickname = generateNickname(req.Email, limits.MinUsername, limits.MaxUsername)
	}

	for {
		_, err := db.GetUserByNickname(r.Context(), nickname)
		if err != nil {
			break
		}
		nickname = generateNicknameWithSuffix(req.Email, limits.MinUsername, limits.MaxUsername, rand.Intn(9000)+1000)
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
		Nickname:     &nickname,
		DateOfBirth:  req.DateOfBirth,
		AboutMe:      req.AboutMe,
		AvatarPath:   avatarPath,
		IsPublic:     true,
	}

	userID, err := db.AddUser(r.Context(), user)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to create account")
		return
	}

	if err := CreateUserSession(w, r, db, getRedis(), userID); err != nil {
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

var nonAlphaNum = regexp.MustCompile(`[^a-z0-9_]`)

func sanitizeNickname(raw string, maxLen int) string {
	lower := strings.ToLower(raw)
	sanitized := nonAlphaNum.ReplaceAllString(lower, "_")
	sanitized = strings.Trim(sanitized, "_")
	if len(sanitized) > maxLen {
		sanitized = sanitized[:maxLen]
	}
	if sanitized == "" {
		sanitized = "user"
	}
	return sanitized
}

func generateNickname(email string, minLen, maxLen int) string {
	parts := strings.SplitN(email, "@", 2)
	base := sanitizeNickname(parts[0], maxLen)
	if len(base) < minLen {
		padding := minLen - len(base)
		base = base + "_" + strings.Repeat("0", padding-1)
		if len(base) > maxLen {
			base = base[:maxLen]
		}
	}
	return base
}

func generateNicknameWithSuffix(email string, minLen, maxLen int, suffix int) string {
	base := generateNickname(email, minLen, maxLen)
	suffixStr := fmt.Sprintf("%d", suffix)
	maxBaseLen := maxLen - len(suffixStr)
	if maxBaseLen < 1 {
		maxBaseLen = 1
	}
	if len(base) > maxBaseLen {
		base = base[:maxBaseLen]
	}
	return base + suffixStr
}
