package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"social-network/backend/config"
	database "social-network/backend/db/sql"
	"social-network/backend/global"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const oauthStateCookie = "oauth_state"

func getOAuthProvider(name string) *config.OAuthProvider {
	cfg := config.GetConfig()
	if cfg == nil {
		return nil
	}
	switch name {
	case "google":
		return &cfg.OAuth.Google
	case "github":
		return &cfg.OAuth.Github
	default:
		return nil
	}
}

// isKnownOAuthProvider reports whether the given provider name is recognized
// (google/github) even if its credentials are not yet configured.
func isKnownOAuthProvider(name string) bool {
	return name == "google" || name == "github"
}

// OAuthLoginHandler redirects the user to the provider's authorization page.
func OAuthLoginHandler(w http.ResponseWriter, r *http.Request, db *database.DataBase) {
	providerName := r.PathValue("provider")
	if !isKnownOAuthProvider(providerName) {
		RespondError(w, http.StatusNotFound, "Unknown OAuth provider")
		return
	}
	provider := getOAuthProvider(providerName)
	if provider == nil || !provider.IsConfigured() {
		RespondError(w, http.StatusServiceUnavailable, "OAuth provider not configured")
		return
	}

	// Generate a random state value and store it in a short-lived cookie to
	// prevent CSRF on the callback.
	state := uuid.New().String()
	http.SetCookie(w, &http.Cookie{
		Name:     oauthStateCookie,
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   600,
	})

	http.Redirect(w, r, provider.GetAuthURL(state), http.StatusFound)
}

// OAuthCallbackHandler handles the provider redirect after the user authorizes
// the app. It exchanges the code for a token, fetches the user's profile,
// finds or creates the local user, and starts a session before redirecting to
// the frontend.
func OAuthCallbackHandler(w http.ResponseWriter, r *http.Request, db *database.DataBase) {
	providerName := r.PathValue("provider")
	if !isKnownOAuthProvider(providerName) {
		RespondError(w, http.StatusNotFound, "Unknown OAuth provider")
		return
	}
	provider := getOAuthProvider(providerName)
	if provider == nil || !provider.IsConfigured() {
		RespondError(w, http.StatusServiceUnavailable, "OAuth provider not configured")
		return
	}

	if errMsg := r.URL.Query().Get("error"); errMsg != "" {
		oauthRedirectError(w, r, "OAuth authorization failed: "+errMsg)
		return
	}

	// Validate the state parameter against the stored cookie.
	state, err := r.Cookie(oauthStateCookie)
	if err != nil || state.Value == "" || state.Value != r.URL.Query().Get("state") {
		oauthRedirectError(w, r, "Invalid OAuth state")
		return
	}
	http.SetCookie(w, &http.Cookie{Name: oauthStateCookie, Path: "/", MaxAge: -1})

	code := r.URL.Query().Get("code")
	if code == "" {
		oauthRedirectError(w, r, "Missing authorization code")
		return
	}

	accessToken, err := provider.ExchangeCodeForToken(code)
	if err != nil {
		oauthRedirectError(w, r, "Failed to exchange authorization code")
		return
	}

	userInfo, err := provider.FetchUserInfo(accessToken)
	if err != nil {
		oauthRedirectError(w, r, "Failed to fetch account information")
		return
	}

	oauthUser, err := parseOAuthUser(providerName, userInfo)
	if err != nil {
		oauthRedirectError(w, r, "Failed to parse account information")
		return
	}

	// GitHub: only trust a verified email for account linking. Fetch the
	// email list and prefer the primary verified address; if none exists,
	// fall back to a synthetic noreply address so we never auto-link to an
	// existing account using an unverified email.
	if providerName == "github" {
		if raw, fetchErr := provider.FetchUserEmails(accessToken); fetchErr == nil {
			if verified := githubVerifiedEmail(raw); verified != "" {
				oauthUser.Email = verified
			} else {
				oauthUser.Email = fmt.Sprintf("%s@users.noreply.github.com", oauthUser.Nickname)
			}
		} else {
			oauthUser.Email = fmt.Sprintf("%s@users.noreply.github.com", oauthUser.Nickname)
		}
	}

	user, err := findOrCreateOAuthUser(r, db, providerName, oauthUser)
	if err != nil {
		oauthRedirectError(w, r, "Failed to sign in")
		return
	}

	if err := CreateUserSession(w, r, db, getRedis(), user.ID); err != nil {
		oauthRedirectError(w, r, "Failed to create session")
		return
	}

	http.Redirect(w, r, config.GetFrontendURL()+"/home", http.StatusFound)
}

// oauthUserInfo holds the normalized profile fields returned by a provider.
type oauthUserInfo struct {
	ProviderID string
	Email      string
	FirstName  string
	LastName   string
	Nickname   string
	AvatarURL  string
}

// parseOAuthUser normalizes the provider-specific user info JSON.
func parseOAuthUser(provider string, raw []byte) (*oauthUserInfo, error) {
	if provider == "google" {
		var g struct {
			ID         string `json:"id"`
			Email      string `json:"email"`
			GivenName  string `json:"given_name"`
			FamilyName string `json:"family_name"`
			Name       string `json:"name"`
			Picture    string `json:"picture"`
		}
		if err := json.Unmarshal(raw, &g); err != nil {
			return nil, err
		}
		firstName, lastName := splitName(g.GivenName, g.FamilyName, g.Name)
		return &oauthUserInfo{
			ProviderID: g.ID,
			Email:      g.Email,
			FirstName:  firstName,
			LastName:   lastName,
			Nickname:   g.Name,
			AvatarURL:  g.Picture,
		}, nil
	}

	// GitHub
	var gh struct {
		ID        int64  `json:"id"`
		Login     string `json:"login"`
		Name      string `json:"name"`
		Email     string `json:"email"`
		AvatarURL string `json:"avatar_url"`
	}
	if err := json.Unmarshal(raw, &gh); err != nil {
		return nil, err
	}

	// GitHub may not return a public email in the primary response; if empty,
	// fall back to a synthetic identifier so the account remains unique.
	email := gh.Email
	if email == "" {
		email = fmt.Sprintf("%s@users.noreply.github.com", gh.Login)
	}

	firstName, lastName := splitName("", "", gh.Name)
	if gh.Name == "" {
		firstName = gh.Login
	}

	return &oauthUserInfo{
		ProviderID: fmt.Sprintf("%d", gh.ID),
		Email:      email,
		FirstName:  firstName,
		LastName:   lastName,
		Nickname:   gh.Login,
		AvatarURL:  gh.AvatarURL,
	}, nil
}

// githubVerifiedEmail returns the primary verified email address from a GitHub
// /user/emails response, falling back to any verified address.
func githubVerifiedEmail(raw []byte) string {
	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}
	if err := json.Unmarshal(raw, &emails); err != nil {
		return ""
	}

	for _, e := range emails {
		if e.Primary && e.Verified {
			return e.Email
		}
	}
	for _, e := range emails {
		if e.Verified {
			return e.Email
		}
	}
	return ""
}

// splitName derives first/last name from the given/name fields.
func splitName(given, family, full string) (string, string) {
	if given != "" {
		if family == "" {
			return given, ""
		}
		return given, family
	}
	if full != "" {
		parts := strings.Fields(full)
		if len(parts) == 1 {
			return parts[0], ""
		}
		return parts[0], parts[len(parts)-1]
	}
	return "", ""
}

// findOrCreateOAuthUser looks up an existing user linked to the provider
// account, or creates a new user (and link) on first login.
func findOrCreateOAuthUser(r *http.Request, db *database.DataBase, provider string, info *oauthUserInfo) (*database.User, error) {
	// Existing linked account → just log in.
	if userID, err := db.GetUserIDByOAuthAccount(r.Context(), provider, info.ProviderID); err == nil {
		if user, err := db.GetUserByID(r.Context(), userID); err == nil {
			return user, nil
		}
	}

	// Email already registered (e.g. via password signup) → link the account.
	if existing, err := db.GetUserByEmail(r.Context(), info.Email); err == nil {
		if err := db.CreateOAuthAccount(r.Context(), existing.ID, provider, info.ProviderID); err == nil {
			return existing, nil
		}
	}

	cfg := config.GetConfig()
	limits := cfg.DatabaseConfiguration.Limits

	// Generate a unique nickname.
	base := sanitizeOAuthNickname(info.Nickname, info.Email, limits.MaxUsername)
	nickname := base
	for i := 0; ; i++ {
		if _, err := db.GetUserByNickname(r.Context(), nickname); err != nil {
			break
		}
		nickname = fmt.Sprintf("%s_%d", base, 1000+i)
	}

	// Generate a random password (the user authenticates via OAuth, so this
	// password is never used directly).
	randomPassword := uuid.New().String() + uuid.New().String()
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(randomPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Download and store the provider avatar, if any.
	avatarPath, err := downloadAvatar(info.AvatarURL)
	if err != nil {
		avatarPath = nil
	}

	user := &database.User{
		UUID:         uuid.New().String(),
		Email:        info.Email,
		PasswordHash: string(hashedPassword),
		FirstName:    info.FirstName,
		LastName:     info.LastName,
		Nickname:     &nickname,
		DateOfBirth:  time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
		AvatarPath:   avatarPath,
		IsPublic:     true,
	}

	userID, err := db.AddUser(r.Context(), user)
	if err != nil {
		return nil, err
	}

	if err := db.CreateOAuthAccount(r.Context(), userID, provider, info.ProviderID); err != nil {
		return nil, err
	}

	user.ID = userID
	return user, nil
}

// sanitizeOAuthNickname derives a valid nickname from the provider name/email.
func sanitizeOAuthNickname(name, email string, maxLen int) string {
	base := ""
	if name != "" {
		base = sanitizeNickname(name, maxLen)
	}
	if base == "" || base == "user" {
		base = generateNickname(email, 1, maxLen)
	}
	if len(base) < 3 {
		base = base + "_" + strings.Repeat("0", 3-len(base))
	}
	return base
}

// downloadAvatar fetches the provider avatar image and saves it locally,
// returning the URL-relative path (e.g. "images/{uuid}.jpg").
func downloadAvatar(url string) (*string, error) {
	if url == "" {
		return nil, fmt.Errorf("no avatar URL")
	}

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("avatar download failed with status %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	ext, ok := extensionForType(contentType, []string{"image/jpeg", "image/jpg", "image/png", "image/gif"})
	if !ok {
		return nil, fmt.Errorf("unsupported avatar type %s", contentType)
	}

	cfg := config.GetConfig()
	dir := global.GetImagePath(cfg.Handlers.Image.PathPrefix)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}

	filename := generateUUID() + ext
	dst, err := os.Create(filepath.Join(dir, filename))
	if err != nil {
		return nil, err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, resp.Body); err != nil {
		return nil, err
	}

	path := "images/" + filename
	return &path, nil
}

// oauthRedirectError redirects back to the frontend login page with an error
// query parameter so the frontend can surface the failure.
func oauthRedirectError(w http.ResponseWriter, r *http.Request, msg string) {
	frontend := config.GetFrontendURL()
	redirectURL := frontend + "/login?oauth_error=" + url.QueryEscape(msg)
	http.Redirect(w, r, redirectURL, http.StatusFound)
}
