package handlers

import (
	"context"
	"fmt"
	"math/rand/v2"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"social-network/backend/cache"
	"social-network/backend/config"
	database "social-network/backend/db/sql"
)

type contextKey string

const (
	userIDKey   contextKey = "userID"
	userUUIDKey contextKey = "userUUID"
)

type (
	writer  = http.ResponseWriter
	request = *http.Request
)

func AuthMiddleware(
	requireAuth bool,
	w writer,
	r request,
	db *database.DataBase,
	redisClient *cache.RedisClient,
	nextHandler func(writer, request, *database.DataBase),
) {
	// Recover from panics raised by the handler (or by the middleware itself).
	// This must be registered before any work happens: a defer placed after
	// nextHandler has already returned can never recover it.
	isJSON := r.Header.Get("Content-Type") == "application/json" || (len(r.URL.Path) >= 4 && r.URL.Path[:4] == "/api")

	defer func() {
		if rec := recover(); rec != nil {
			if isJSON {
				RespondError(w, http.StatusInternalServerError, "Internal server error")
			} else {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}

			for range 10 {
				fmt.Fprintln(os.Stderr, "UNEXPECTED PANIC IN HANDLER:")
			}
			fmt.Fprintf(os.Stderr, "Panic: %v\n", rec)
		}
	}()

	// CORS compliant headers
	frontendURL := config.GetFrontendURL()
	w.Header().Set("Access-Control-Allow-Origin", frontendURL)
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Resolve the session before rate limiting so the limits can be applied
	// per user. Everything is proxied through Caddy, so keying on RemoteAddr
	// alone would put every user in the same bucket.
	user, err := GetUserFromCookie(r, db, redisClient)
	isAuthenticated := err == nil && user != nil

	rateLimitID := "ip:" + clientIP(r)
	if isAuthenticated {
		rateLimitID = fmt.Sprintf("user:%d", user.ID)
	}
	rateLimitTag := fmt.Sprint(rateLimitID, r.URL.Path)

	universalCount, universalInterval := config.GetUniversalRateLimit()
	pathCount, pathInterval := config.GetRateLimit(r.URL.Path)

	blocked := false

	if redisClient != nil {
		if isBlocked, err := redisClient.CheckRateLimit(r.Context(), rateLimitID, int64(universalCount), universalInterval); err == nil && isBlocked {
			blocked = true
		} else if isBlocked, err := redisClient.CheckRateLimit(r.Context(), rateLimitTag, int64(pathCount), pathInterval); err == nil && isBlocked {
			blocked = true
		}
	} else {
		if BlockRequest(int64(universalCount), int64(universalInterval*1000), rateLimitID) {
			blocked = true
		} else if BlockRequest(int64(pathCount), int64(pathInterval*1000), rateLimitTag) {
			blocked = true
		}
	}

	if blocked {
		RespondError(w, http.StatusTooManyRequests, "Too many requests")
		return
	}

	if requireAuth && !isAuthenticated {
		if isJSON {
			RespondError(w, http.StatusUnauthorized, "Authentication required")
		} else {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
		}
		return
	}

	if isAuthenticated {
		ctx := context.WithValue(r.Context(), userIDKey, user.ID)
		ctx = context.WithValue(ctx, userUUIDKey, user.UUID)
		r = r.WithContext(ctx)
	}

	nextHandler(w, r, db)
}

func GetUserIDFromContext(r *http.Request) (int64, bool) {
	userID, ok := r.Context().Value(userIDKey).(int64)
	return userID, ok
}

// GetUserUUIDFromContext returns the authenticated user's public identifier.
// Authorization decisions compare this rather than the internal numeric id, so
// they use the same identifier the API exposes. A missing or empty value is
// reported as not-ok so a blank comparison can never authorise.
func GetUserUUIDFromContext(r *http.Request) (string, bool) {
	userUUID, ok := r.Context().Value(userUUIDKey).(string)
	return userUUID, ok && userUUID != ""
}

// clientIP returns the best-effort client address for rate limiting.
// Requests are proxied through Caddy, so RemoteAddr is always the proxy.
// Caddy appends the address it saw to X-Forwarded-For, which makes the
// right-most entry the real peer (anything earlier can be supplied by the
// client, so it is not trusted).
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if ip := strings.TrimSpace(parts[len(parts)-1]); ip != "" {
			return ip
		}
	}
	if xri := strings.TrimSpace(r.Header.Get("X-Real-IP")); xri != "" {
		return xri
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

type requestEntry struct {
	count              atomic.Int64
	timeSinceLastReset atomic.Int64
}

var KeyAttempts = initSyncMap()

func initSyncMap() *sync.Map {
	var m sync.Map
	return &m
}

func syncMapCleaner() {
	for {
		time.Sleep(time.Hour)
		time.Sleep(time.Minute * time.Duration(rand.IntN(20)))
		KeyAttempts.Clear()
	}
}

func BlockRequest(limit int64, intervalMilli int64, key string) bool {
	entryAny, ok := KeyAttempts.Load(key)

	if !ok {
		newEntry := requestEntry{}
		newEntry.count.Store(1)
		newEntry.timeSinceLastReset.Store(time.Now().UnixMilli())
		entryAny, ok = KeyAttempts.LoadOrStore(key, &newEntry)

		if !ok {
			return false
		}
	}

	entry := entryAny.(*requestEntry)

	count := entry.count.Load()
	timesOver := (time.Now().UnixMilli() - entry.timeSinceLastReset.Load()) / intervalMilli

	if timesOver > 0 {
		entry.count.Store(max(1, count-(timesOver*limit)))
		entry.timeSinceLastReset.Store(time.Now().UnixMilli())
	} else {
		entry.count.Add(1)
	}

	if entry.count.Load() > limit {
		fmt.Println(entry.count.Load(), " more than ", limit)
		return true
	}

	return false
}
