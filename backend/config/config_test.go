package config

import "testing"

func TestApplySecretOverrides(t *testing.T) {
	base := func() *Config {
		return &Config{
			OAuth: OAuthConfig{
				Google: OAuthProvider{ClientID: "file-google-id", ClientSecret: "file-google-secret"},
				Github: OAuthProvider{ClientID: "file-github-id", ClientSecret: "file-github-secret"},
			},
		}
	}

	tests := []struct {
		name          string
		env           map[string]string
		wantGoogleID  string
		wantGoogleSec string
		wantGithubID  string
		wantGithubSec string
	}{
		{
			name:          "no env keeps file values",
			env:           nil,
			wantGoogleID:  "file-google-id",
			wantGoogleSec: "file-google-secret",
			wantGithubID:  "file-github-id",
			wantGithubSec: "file-github-secret",
		},
		{
			name: "env overrides all providers",
			env: map[string]string{
				"GOOGLE_CLIENT_ID":     "env-google-id",
				"GOOGLE_CLIENT_SECRET": "env-google-secret",
				"GITHUB_CLIENT_ID":     "env-github-id",
				"GITHUB_CLIENT_SECRET": "env-github-secret",
			},
			wantGoogleID:  "env-google-id",
			wantGoogleSec: "env-google-secret",
			wantGithubID:  "env-github-id",
			wantGithubSec: "env-github-secret",
		},
		{
			name:          "partial env overrides only what is set",
			env:           map[string]string{"GITHUB_CLIENT_SECRET": "env-github-secret"},
			wantGoogleID:  "file-google-id",
			wantGoogleSec: "file-google-secret",
			wantGithubID:  "file-github-id",
			wantGithubSec: "env-github-secret",
		},
		{
			name:          "empty env value does not clobber file value",
			env:           map[string]string{"GOOGLE_CLIENT_ID": ""},
			wantGoogleID:  "file-google-id",
			wantGoogleSec: "file-google-secret",
			wantGithubID:  "file-github-id",
			wantGithubSec: "file-github-secret",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, k := range []string{
				"GOOGLE_CLIENT_ID", "GOOGLE_CLIENT_SECRET",
				"GITHUB_CLIENT_ID", "GITHUB_CLIENT_SECRET",
			} {
				t.Setenv(k, "")
			}
			for k, v := range tt.env {
				t.Setenv(k, v)
			}

			c := base()
			applySecretOverrides(c)

			if c.OAuth.Google.ClientID != tt.wantGoogleID {
				t.Errorf("google client_id = %q, want %q", c.OAuth.Google.ClientID, tt.wantGoogleID)
			}
			if c.OAuth.Google.ClientSecret != tt.wantGoogleSec {
				t.Errorf("google client_secret = %q, want %q", c.OAuth.Google.ClientSecret, tt.wantGoogleSec)
			}
			if c.OAuth.Github.ClientID != tt.wantGithubID {
				t.Errorf("github client_id = %q, want %q", c.OAuth.Github.ClientID, tt.wantGithubID)
			}
			if c.OAuth.Github.ClientSecret != tt.wantGithubSec {
				t.Errorf("github client_secret = %q, want %q", c.OAuth.Github.ClientSecret, tt.wantGithubSec)
			}
		})
	}
}
