package config

import "testing"

func TestApplyOAuthEnvOverrides(t *testing.T) {
	base := func() *Config {
		return &Config{
			OAuth: OAuthConfig{
				Google: OAuthProvider{
					ClientID:     "file-google-id",
					ClientSecret: "file-google-secret",
					RedirectURI:  "http://file/google/callback",
				},
				Github: OAuthProvider{
					ClientID:     "file-github-id",
					ClientSecret: "file-github-secret",
					RedirectURI:  "http://file/github/callback",
				},
			},
		}
	}

	tests := []struct {
		name          string
		env           map[string]string
		wantGoogleID  string
		wantGoogleSec string
		wantGoogleURI string
		wantGithubID  string
		wantGithubSec string
		wantGithubURI string
	}{
		{
			name:          "no env keeps file values",
			env:           nil,
			wantGoogleID:  "file-google-id",
			wantGoogleSec: "file-google-secret",
			wantGoogleURI: "http://file/google/callback",
			wantGithubID:  "file-github-id",
			wantGithubSec: "file-github-secret",
			wantGithubURI: "http://file/github/callback",
		},
		{
			name: "env overrides all values",
			env: map[string]string{
				"GOOGLE_CLIENT_ID":     "env-google-id",
				"GOOGLE_CLIENT_SECRET": "env-google-secret",
				"GOOGLE_REDIRECT_URI":  "http://env/google/callback",
				"GITHUB_CLIENT_ID":     "env-github-id",
				"GITHUB_CLIENT_SECRET": "env-github-secret",
				"GITHUB_REDIRECT_URI":  "http://env/github/callback",
			},
			wantGoogleID:  "env-google-id",
			wantGoogleSec: "env-google-secret",
			wantGoogleURI: "http://env/google/callback",
			wantGithubID:  "env-github-id",
			wantGithubSec: "env-github-secret",
			wantGithubURI: "http://env/github/callback",
		},
		{
			name:          "partial env overrides only what is set",
			env:           map[string]string{"GITHUB_CLIENT_SECRET": "env-github-secret", "GOOGLE_REDIRECT_URI": "http://env/google/callback"},
			wantGoogleID:  "file-google-id",
			wantGoogleSec: "file-google-secret",
			wantGoogleURI: "http://env/google/callback",
			wantGithubID:  "file-github-id",
			wantGithubSec: "env-github-secret",
			wantGithubURI: "http://file/github/callback",
		},
		{
			name:          "empty env value does not clobber file value",
			env:           map[string]string{"GOOGLE_CLIENT_ID": ""},
			wantGoogleID:  "file-google-id",
			wantGoogleSec: "file-google-secret",
			wantGoogleURI: "http://file/google/callback",
			wantGithubID:  "file-github-id",
			wantGithubSec: "file-github-secret",
			wantGithubURI: "http://file/github/callback",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, k := range []string{
				"GOOGLE_CLIENT_ID", "GOOGLE_CLIENT_SECRET", "GOOGLE_REDIRECT_URI",
				"GITHUB_CLIENT_ID", "GITHUB_CLIENT_SECRET", "GITHUB_REDIRECT_URI",
			} {
				t.Setenv(k, "")
			}
			for k, v := range tt.env {
				t.Setenv(k, v)
			}

			c := base()
			applyOAuthEnvOverrides(c)

			if c.OAuth.Google.ClientID != tt.wantGoogleID {
				t.Errorf("google client_id = %q, want %q", c.OAuth.Google.ClientID, tt.wantGoogleID)
			}
			if c.OAuth.Google.ClientSecret != tt.wantGoogleSec {
				t.Errorf("google client_secret = %q, want %q", c.OAuth.Google.ClientSecret, tt.wantGoogleSec)
			}
			if c.OAuth.Google.RedirectURI != tt.wantGoogleURI {
				t.Errorf("google redirect_uri = %q, want %q", c.OAuth.Google.RedirectURI, tt.wantGoogleURI)
			}
			if c.OAuth.Github.ClientID != tt.wantGithubID {
				t.Errorf("github client_id = %q, want %q", c.OAuth.Github.ClientID, tt.wantGithubID)
			}
			if c.OAuth.Github.ClientSecret != tt.wantGithubSec {
				t.Errorf("github client_secret = %q, want %q", c.OAuth.Github.ClientSecret, tt.wantGithubSec)
			}
			if c.OAuth.Github.RedirectURI != tt.wantGithubURI {
				t.Errorf("github redirect_uri = %q, want %q", c.OAuth.Github.RedirectURI, tt.wantGithubURI)
			}
		})
	}
}
