package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"social-network/backend/cache"
	database "social-network/backend/db/sql"
	ws "social-network/backend/server/websocket"

	"github.com/google/uuid"
)

var (
	ErrSessionNotFound = errors.New("session not found")
	ErrNoCookie        = errors.New("no session cookie")

	globalRedis *cache.RedisClient
)

func setGlobalRedis(rc *cache.RedisClient) {
	globalRedis = rc
}

func getRedis() *cache.RedisClient {
	return globalRedis
}

func resolveUserID(r *http.Request, db *database.DataBase, idStr string) (int64, error) {
	userID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		user, err := db.GetUserByUUID(r.Context(), idStr)
		if err != nil {
			return 0, fmt.Errorf("invalid user identifier")
		}
		userID = user.ID
	}
	return userID, nil
}

func sendNotification(db *database.DataBase, userID int64, fromUserID int64, notifType string, content string, relatedID *int64, relatedUUID *string) {
	notif, err := db.CreateNotification(context.Background(), &database.Notification{
		UserID:      userID,
		FromUserID:  &fromUserID,
		Type:        notifType,
		Content:     content,
		RelatedID:   relatedID,
		RelatedUUID: relatedUUID,
	})
	if err != nil {
		fmt.Printf("Error creating notification: %v\n", err)
		return
	}

	if GlobalHub != nil {
		var fromUserUUID *string
		if fromUser, err := db.GetUserByID(context.Background(), fromUserID); err == nil {
			uuid := fromUser.UUID
			fromUserUUID = &uuid
		}

		payload, _ := json.Marshal(ws.NotificationPayload{
			ID:           notif,
			Type:         notifType,
			Content:      content,
			RelatedID:    relatedID,
			RelatedUUID:  relatedUUID,
			FromUserID:   &fromUserID,
			FromUserUUID: fromUserUUID,
			TargetID:     userID,
			IsRead:       false,
			CreatedAt:    time.Now().UTC().Format(time.RFC3339),
		})
		GlobalHub.SendToUser(userID, &ws.WSMessage{
			Type:      ws.TypeNotification,
			Payload:   payload,
			SenderID:  fromUserID,
			Timestamp: time.Now(),
		})
	}
}

type JSONResponse struct {
	Message string      `json:"message,omitempty"`
	Error   string      `json:"error,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func RespondJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if payload != nil {
		json.NewEncoder(w).Encode(payload)
	}
}

func RespondError(w http.ResponseWriter, statusCode int, errorMessage string) {
	RespondJSON(w, statusCode, JSONResponse{
		Error: errorMessage,
	})
}

func RespondSuccess(w http.ResponseWriter, statusCode int, message string, data interface{}) {
	RespondJSON(w, statusCode, JSONResponse{
		Message: message,
		Data:    data,
	})
}

func GenerateSessionID() string {
	return uuid.New().String()
}

func SetSessionCookie(w http.ResponseWriter, r *http.Request, sessionID string, expiration time.Time) {
	isTLS := r.TLS != nil
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   isTLS,
		SameSite: sameSiteMode(isTLS),
		Expires:  expiration,
	})
}

func sameSiteMode(isTLS bool) http.SameSite {
	if isTLS {
		return http.SameSiteNoneMode
	}
	return http.SameSiteLaxMode
}

func ClearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
	})
}

func CreateUserSession(w http.ResponseWriter, r *http.Request, db *database.DataBase, redisClient *cache.RedisClient, userID int64) error {
	sessionID := GenerateSessionID()

	session := &database.Session{
		SessionID: sessionID,
		UserID:    userID,
		IPAddress: r.RemoteAddr,
		UserAgent: r.UserAgent(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	_, err := db.CreateSession(r.Context(), session)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	if redisClient != nil {
		if cacheErr := redisClient.CacheSession(r.Context(), cache.SessionKey(sessionID), userID, cache.SessionTTL); cacheErr != nil {
			fmt.Printf("Warning: failed to cache session in Redis: %v\n", cacheErr)
		}
	}

	SetSessionCookie(w, r, sessionID, session.ExpiresAt)
	return nil
}

func GetUserFromCookie(r *http.Request, db *database.DataBase, redisClient *cache.RedisClient) (*database.User, error) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		if err == http.ErrNoCookie {
			return nil, ErrNoCookie
		}
		return nil, fmt.Errorf("error reading cookie: %w", err)
	}

	if _, err := uuid.Parse(cookie.Value); err != nil {
		return nil, ErrSessionNotFound
	}

	sessionKey := cache.SessionKey(cookie.Value)
	var userID int64
	cacheHit := false

	if redisClient != nil {
		if uid, cacheErr := redisClient.GetSession(r.Context(), sessionKey); cacheErr == nil {
			userID = uid
			cacheHit = true
		}
	}

	if !cacheHit {
		session, err := db.GetSession(r.Context(), cookie.Value)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, ErrSessionNotFound
			}
			return nil, fmt.Errorf("error getting session: %w", err)
		}

		if session == nil || !session.IsActive {
			return nil, ErrSessionNotFound
		}

		userID = session.UserID

		if redisClient != nil {
			if cacheErr := redisClient.CacheSession(r.Context(), sessionKey, userID, cache.SessionTTL); cacheErr != nil {
				fmt.Printf("Warning: failed to cache session in Redis: %v\n", cacheErr)
			}
		}
	}

	user, err := db.GetUserByID(r.Context(), userID)
	if err != nil {
		return nil, fmt.Errorf("error getting user: %w", err)
	}

	return user, nil
}
