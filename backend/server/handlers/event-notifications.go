package handlers

import (
	"net/http"
	"strconv"

	database "social-network/backend/db/sql"
)

func GetNotificationsHandler(w http.ResponseWriter, r *http.Request, db *database.DataBase) {
	if r.Method != http.MethodGet {
		RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := GetUserIDFromContext(r)
	if !ok {
		RespondError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 20
	offset := 0

	if limitStr != "" {
		if v, err := strconv.Atoi(limitStr); err == nil && v > 0 && v <= maxPageLimit() {
			limit = v
		}
	}
	if offsetStr != "" {
		if v, err := strconv.Atoi(offsetStr); err == nil && v >= 0 {
			offset = v
		}
	}

	notifications, err := db.GetUserNotifications(r.Context(), userID, limit, offset)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to fetch notifications")
		return
	}

	if notifications == nil {
		notifications = []*database.Notification{}
	}

	RespondSuccess(w, http.StatusOK, "Notifications retrieved", notifications)
}

func GetUnreadNotificationCountHandler(w http.ResponseWriter, r *http.Request, db *database.DataBase) {
	if r.Method != http.MethodGet {
		RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := GetUserIDFromContext(r)
	if !ok {
		RespondError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	count, err := db.GetUnreadNotificationCount(r.Context(), userID)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to get unread count")
		return
	}

	RespondSuccess(w, http.StatusOK, "Unread count retrieved", map[string]int{
		"count": count,
	})
}

func MarkNotificationReadHandler(w http.ResponseWriter, r *http.Request, db *database.DataBase) {
	if r.Method != http.MethodPut {
		RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := GetUserIDFromContext(r)
	if !ok {
		RespondError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	notifIDStr := r.PathValue("id")
	notifID, err := strconv.ParseInt(notifIDStr, 10, 64)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "Invalid notification ID")
		return
	}

	if err := db.MarkNotificationAsRead(r.Context(), notifID, userID); err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to mark notification as read")
		return
	}

	RespondSuccess(w, http.StatusOK, "Notification marked as read", nil)
}

func MarkAllNotificationsReadHandler(w http.ResponseWriter, r *http.Request, db *database.DataBase) {
	if r.Method != http.MethodPut {
		RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := GetUserIDFromContext(r)
	if !ok {
		RespondError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	if err := db.MarkAllNotificationsAsRead(r.Context(), userID); err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to mark all notifications as read")
		return
	}

	RespondSuccess(w, http.StatusOK, "All notifications marked as read", nil)
}
