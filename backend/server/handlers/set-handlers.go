package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"path"
)

type Config struct {
	Images                *ImgConfig                `json:"image"`
	MaxPostSize           int64                     `json:"max_post_size"`
	CookieExpirationHours int                       `json:"cookie_expiration_hours"`
	RateLimits            map[string]map[string]any `json:"rate_limits"`
}

type ImgConfig struct {
	MaxSize     int64    `json:"max_size"`
	AllowedExts []string `json:"allowed_extensions"`
}

type endpoint struct {
	path                 string
	rateLimitMaxRequests int
	rateLimitInterval    float64
	requireAuth          bool // this will be true if the endpoint needs you to be logged in
	nextHandler          func(http.ResponseWriter, *http.Request, *sql.DB)
}

func makeEndpoint(path string, rateLimitMaxRequests int, rateLimitInterval float64, requireAuth bool, nextHandler func(http.ResponseWriter, *http.Request, *sql.DB)) endpoint {
	return endpoint{
		path:                 path,
		rateLimitMaxRequests: rateLimitMaxRequests,
		rateLimitInterval:    rateLimitInterval,
		requireAuth:          requireAuth,
		nextHandler:          nextHandler,
	}
}

var Configuration *Config

func SetHandlers(db *sql.DB) *http.ServeMux {
	hub := GetWebSocketHub(db)
	go hub.Run()

	go syncMapCleaner()

	mux := http.NewServeMux()

	staticDirCss := path.Join("web", "static")
	staticDirImg := path.Join("web", "images")
	staticDirJs := path.Join("web", "js")

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDirCss))))
	mux.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir(staticDirImg))))
	mux.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir(staticDirJs))))

	endpoints := []endpoint{
		// This is where all our endpoints will go
	}

	for _, ep := range endpoints {
		limitCount := ep.rateLimitMaxRequests
		limitInterval := ep.rateLimitInterval

		if Configuration != nil {
			if limits, ok := Configuration.RateLimits[ep.path]; ok {
				if lCount, ok1 := limits["rate_limit_count"].(float64); ok1 {
					limitCount = int(lCount)
				}
				if lSeconds, ok2 := limits["rate_limit_second_interval"].(float64); ok2 {
					limitInterval = lSeconds
				}
			}
		}

		universalCount := 30
		universalSeconds := 2.0

		if Configuration != nil {
			if universal, ok := Configuration.RateLimits["universal"]; ok {
				if lCount, ok1 := universal["rate_limit_count"].(float64); ok1 {
					universalCount = int(lCount)
				}
				if lSeconds, ok2 := universal["rate_limit_second_interval"].(float64); ok2 {
					universalSeconds = lSeconds
				}
			}
		}

		handler := ep.nextHandler
		mux.HandleFunc(ep.path, func(w http.ResponseWriter, r *http.Request) {
			AuthMiddleware(
				ep.requireAuth,
				w, r, db,
				handler,
				limitCount, limitInterval,
				universalCount, universalSeconds,
			)
		})
	}

	fmt.Println("✅ Routes registered successfully")
	return mux
}
