package handlers

import (
	"context"
	"fmt"
	"math/rand/v2"
	"net"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"social-network/backend/config"
	database "social-network/backend/db/sql"
)

type contextKey string

const userIDKey contextKey = "userID"

type (
	writer  = http.ResponseWriter
	request = *http.Request
)

func AuthMiddleware(
	requireAuth bool,
	w writer,
	r request,
	db *database.DataBase,
	nextHandler func(writer, request, *database.DataBase),
) {
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

	remoteAddr, _, _ := net.SplitHostPort(r.RemoteAddr)
	rateLimitTag := fmt.Sprint(remoteAddr, r.URL.Path)

	universalCount, universalInterval := config.GetUniversalRateLimit()
	pathCount, pathInterval := config.GetRateLimit(r.URL.Path)

	// universal rate limit, applies to all requests from an IP
	if BlockRequest(int64(universalCount), int64(universalInterval*1000), remoteAddr) {
		fmt.Println("blocking universal rate limit: ", remoteAddr)
		RespondError(w, http.StatusTooManyRequests, "Too many requests")
		return
	}

	// path specific rate limit
	if BlockRequest(int64(pathCount), int64(pathInterval*1000), rateLimitTag) {
		fmt.Println("blocking endpoint rate limit: ", rateLimitTag)
		RespondError(w, http.StatusTooManyRequests, "Too many requests to endpoint")
		return
	}

	user, err := GetUserFromCookie(r, db)
	isAuthenticated := err == nil && user != nil

	isJSON := r.Header.Get("Content-Type") == "application/json" || (len(r.URL.Path) >= 4 && r.URL.Path[:4] == "/api")

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
		r = r.WithContext(ctx)
	}

	nextHandler(w, r, db)

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
}

func GetUserIDFromContext(r *http.Request) (int64, bool) {
	userID, ok := r.Context().Value(userIDKey).(int64)
	return userID, ok
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
