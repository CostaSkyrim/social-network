package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	database "social-network/backend/db/sql"
)

func TestClientIP(t *testing.T) {
	tests := []struct {
		name       string
		xff        string
		xRealIP    string
		remoteAddr string
		want       string
	}{
		{
			name:       "falls back to RemoteAddr host when no proxy headers",
			remoteAddr: "203.0.113.9:54321",
			want:       "203.0.113.9",
		},
		{
			name:       "single X-Forwarded-For entry wins over RemoteAddr",
			xff:        "198.51.100.7",
			remoteAddr: "172.20.0.5:54321",
			want:       "198.51.100.7",
		},
		{
			name:       "takes the right-most X-Forwarded-For entry",
			xff:        "1.2.3.4, 198.51.100.7",
			remoteAddr: "172.20.0.5:54321",
			want:       "198.51.100.7",
		},
		{
			name:       "trims whitespace around X-Forwarded-For entries",
			xff:        "  1.2.3.4 ,  198.51.100.7  ",
			remoteAddr: "172.20.0.5:54321",
			want:       "198.51.100.7",
		},
		{
			name:       "X-Real-IP is used when X-Forwarded-For is absent",
			xRealIP:    "198.51.100.7",
			remoteAddr: "172.20.0.5:54321",
			want:       "198.51.100.7",
		},
		{
			name:       "RemoteAddr without a port is returned as-is",
			remoteAddr: "198.51.100.7",
			want:       "198.51.100.7",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/api/health", nil)
			r.RemoteAddr = tc.remoteAddr
			if tc.xff != "" {
				r.Header.Set("X-Forwarded-For", tc.xff)
			}
			if tc.xRealIP != "" {
				r.Header.Set("X-Real-IP", tc.xRealIP)
			}

			if got := clientIP(r); got != tc.want {
				t.Errorf("clientIP() = %q, want %q", got, tc.want)
			}
		})
	}
}

// TestAuthMiddlewareRecoversHandlerPanic is a regression test for the deferred
// recover being registered after nextHandler had already returned, which meant
// a panicking handler was never recovered by the middleware.
func TestAuthMiddlewareRecoversHandlerPanic(t *testing.T) {
	tests := []struct {
		name        string
		target      string
		contentType string
		wantJSON    bool
	}{
		{name: "api path gets a JSON error", target: "/api/boom", wantJSON: true},
		{name: "json content type gets a JSON error", target: "/boom", contentType: "application/json", wantJSON: true},
		{name: "plain path gets a plain error", target: "/boom", wantJSON: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, tc.target, nil)
			if tc.contentType != "" {
				r.Header.Set("Content-Type", tc.contentType)
			}

			panicking := func(w http.ResponseWriter, r *http.Request, db *database.DataBase) {
				panic("boom")
			}

			// nil db/redis is fine: without a session cookie
			// GetUserFromCookie returns before touching either, and the
			// config getters fall back to defaults when unconfigured.
			AuthMiddleware(false, rec, r, nil, nil, panicking)

			if rec.Code != http.StatusInternalServerError {
				t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
			}
			body := rec.Body.String()
			if !strings.Contains(body, "Internal server error") {
				t.Errorf("body = %q, want it to mention the error", body)
			}
			if gotJSON := strings.Contains(body, `"error"`); gotJSON != tc.wantJSON {
				t.Errorf("json body = %v, want %v (body = %q)", gotJSON, tc.wantJSON, body)
			}
		})
	}
}
