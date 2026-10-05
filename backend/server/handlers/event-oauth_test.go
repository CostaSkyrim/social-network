package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"social-network/backend/config"
)

func TestOAuthProvidersHandler(t *testing.T) {
	previous := config.AppConfig
	t.Cleanup(func() { config.AppConfig = previous })

	config.AppConfig = &config.Config{
		OAuth: config.OAuthConfig{
			Google: config.OAuthProvider{ClientID: "google-id", ClientSecret: "google-secret"},
			Github: config.OAuthProvider{}, // not configured
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/auth/providers", nil)
	rec := httptest.NewRecorder()

	OAuthProvidersHandler(rec, req, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var body struct {
		Data map[string]bool `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !body.Data["google"] {
		t.Errorf("google = false, want true when client id and secret are set")
	}
	if body.Data["github"] {
		t.Errorf("github = true, want false when credentials are empty")
	}
}
