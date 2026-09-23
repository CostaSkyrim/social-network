package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
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
