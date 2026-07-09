package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net"
	"net/http"
	"os"
	database "social-network/backend/db/sql"
	"sync"
	"sync/atomic"
	"time"
)

type contextKey string

const userIDKey contextKey = "userID"

type writer = http.ResponseWriter
type request = *http.Request

func AuthMiddleware(
	requireAuth bool,
	w writer,
	r request,
	db *database.DataBase,
	nextHandler func(writer, request, *database.DataBase),
	rateLimitMaxRequests int,
	rateLimitIntervals float64,
	universalRateLimitRequests int,
	universalRateLimitInterval float64,
) {
	remoteAddr, _, _ := net.SplitHostPort(r.RemoteAddr)
	rateLimitTag := fmt.Sprint(remoteAddr, r.URL.Path)

	if BlockRequest(int64(universalRateLimitRequests), int64(universalRateLimitInterval*1000), remoteAddr) {
		fmt.Println("blocking general: ", remoteAddr)
		return
	}

	user, err := GetUserFromCookie(r, db)
	isAuthenticated := err == nil && user != nil

	isJSON := r.Header.Get("Content-Type") == "application/json" || (len(r.URL.Path) >= 4 && r.URL.Path[:4] == "/api")

	if requireAuth && !isAuthenticated {
		if isJSON {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Authentication required",
			})
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
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{
					"error": "Inernal server error",
				})
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
