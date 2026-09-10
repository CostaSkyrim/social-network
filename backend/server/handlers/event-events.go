package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"social-network/backend/config"
	database "social-network/backend/db/sql"
)

type CreateEventRequest struct {
	Title         string  `json:"title"`
	Description   string  `json:"description"`
	EventDatetime string  `json:"event_datetime"`
	ImagePath     *string `json:"image_path,omitempty"`
}

type EventRSVPRequest struct {
	Response string `json:"response"`
}

type eventWithMeta struct {
	*database.Event
	Going      int     `json:"going"`
	NotGoing   int     `json:"not_going"`
	Total      int     `json:"total"`
	MyResponse *string `json:"my_response,omitempty"`
}

// requireGroupMember checks that the requesting user is an accepted member or the creator of a group.
func requireGroupMember(w http.ResponseWriter, r *http.Request, db *database.DataBase, groupID, userID int64) bool {
	group, err := db.GetGroup(r.Context(), groupID)
	if err != nil {
		RespondError(w, http.StatusNotFound, "Group not found")
		return false
	}

	if group.CreatorID == userID {
		return true
	}

	members, _ := db.GetGroupMembers(r.Context(), groupID)
	for _, m := range members {
		if m.UserID == userID && m.Status == "accepted" {
			return true
		}
	}

	RespondError(w, http.StatusForbidden, "Must be a group member to access this content")
	return false
}

// resolveEvent looks up an event by its UUID (from the URL path). Responds with
// an error and returns false on failure.
func resolveEvent(w http.ResponseWriter, r *http.Request, db *database.DataBase) (*database.Event, bool) {
	eventUUID := r.PathValue("id")
	if eventUUID == "" {
		RespondError(w, http.StatusBadRequest, "Event ID required")
		return nil, false
	}

	event, err := db.GetEventByUUID(r.Context(), eventUUID)
	if err != nil {
		RespondError(w, http.StatusNotFound, "Event not found")
		return nil, false
	}

	return event, true
}

func validateEventRequest(req *CreateEventRequest) (time.Time, string) {
	req.Title = strings.TrimSpace(req.Title)

	cfg := config.GetConfig()
	var minTitle, maxTitle int
	if cfg != nil {
		minTitle = cfg.DatabaseConfiguration.Limits.MinTitle
		maxTitle = cfg.DatabaseConfiguration.Limits.MaxTitle
	}
	if minTitle == 0 {
		minTitle = 5
	}
	if maxTitle == 0 {
		maxTitle = 300
	}

	if req.Title == "" {
		return time.Time{}, "Title is required"
	}
	if len(req.Title) < minTitle {
		return time.Time{}, "Title must be at least 5 characters long"
	}
	if len(req.Title) > maxTitle {
		return time.Time{}, "Title must be at most 300 characters long"
	}

	if req.EventDatetime == "" {
		return time.Time{}, "Event datetime is required"
	}

	dt, err := time.Parse(time.RFC3339, req.EventDatetime)
	if err != nil {
		return time.Time{}, "Invalid event datetime format (use RFC3339)"
	}

	if !dt.After(time.Now()) {
		return time.Time{}, "Event datetime must be in the future"
	}

	return dt, ""
}

func buildEventWithMeta(db *database.DataBase, r *http.Request, event *database.Event, userID int64) (*eventWithMeta, error) {
	counts, err := db.GetEventResponseCounts(r.Context(), event.ID)
	if err != nil {
		return nil, err
	}

	myResp, err := db.GetEventResponseByUser(r.Context(), event.ID, userID)
	if err != nil {
		return nil, err
	}

	total := counts["going"] + counts["not_going"]

	meta := &eventWithMeta{
		Event:    event,
		Going:    counts["going"],
		NotGoing: counts["not_going"],
		Total:    total,
	}
	if myResp != nil {
		meta.MyResponse = &myResp.Response
	}

	return meta, nil
}

func GroupEventsHandler(w http.ResponseWriter, r *http.Request, db *database.DataBase) {
	switch r.Method {
	case http.MethodPost:
		createEventHandler(w, r, db)
	case http.MethodGet:
		getGroupEventsHandler(w, r, db)
	default:
		RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func createEventHandler(w http.ResponseWriter, r *http.Request, db *database.DataBase) {
	userID, ok := GetUserIDFromContext(r)
	if !ok {
		RespondError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	groupID, ok := resolveGroupID(w, r, db)
	if !ok {
		return
	}

	if !requireGroupMember(w, r, db, groupID, userID) {
		return
	}

	var req CreateEventRequest
	if isMultipart(r) {
		if err := parseMultipartForm(r); err != nil {
			RespondError(w, http.StatusBadRequest, "Invalid multipart form")
			return
		}
		req.Title = r.FormValue("title")
		req.Description = r.FormValue("description")
		req.EventDatetime = r.FormValue("event_datetime")
		imgPath, ok := multipartImage(w, r)
		if !ok {
			return
		}
		req.ImagePath = imgPath
	} else {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			RespondError(w, http.StatusBadRequest, "Invalid request body")
			return
		}
	}

	eventDatetime, validationErr := validateEventRequest(&req)
	if validationErr != "" {
		RespondError(w, http.StatusBadRequest, validationErr)
		return
	}

	event := &database.Event{
		UUID:          generateUUID(),
		GroupID:       groupID,
		CreatorID:     userID,
		Title:         req.Title,
		Description:   req.Description,
		ImagePath:     req.ImagePath,
		EventDateTime: eventDatetime,
	}

	eventID, err := db.CreateEvent(r.Context(), event)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to create event")
		return
	}

	// Notify all accepted members except the creator
	group, _ := db.GetGroup(r.Context(), groupID)
	members, _ := db.GetGroupMembers(r.Context(), groupID)
	for _, m := range members {
		if m.UserID == userID || m.Status != "accepted" {
			continue
		}
		sendNotification(db, m.UserID, userID, NotifNewEvent, "New event '"+req.Title+"' in group: "+group.Title, &eventID, &group.UUID)
	}

	RespondSuccess(w, http.StatusCreated, "Event created", map[string]interface{}{
		"id": event.UUID,
	})
}

func getGroupEventsHandler(w http.ResponseWriter, r *http.Request, db *database.DataBase) {
	userID, ok := GetUserIDFromContext(r)
	if !ok {
		RespondError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	groupID, ok := resolveGroupID(w, r, db)
	if !ok {
		return
	}

	if !requireGroupMember(w, r, db, groupID, userID) {
		return
	}

	events, err := db.GetGroupEvents(r.Context(), groupID)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to fetch events")
		return
	}

	metaList := []*eventWithMeta{}
	for _, e := range events {
		meta, err := buildEventWithMeta(db, r, e, userID)
		if err != nil {
			RespondError(w, http.StatusInternalServerError, "Failed to fetch event details")
			return
		}
		metaList = append(metaList, meta)
	}

	RespondSuccess(w, http.StatusOK, "Events retrieved", metaList)
}

func GetEventHandler(w http.ResponseWriter, r *http.Request, db *database.DataBase) {
	if r.Method != http.MethodGet {
		RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := GetUserIDFromContext(r)
	if !ok {
		RespondError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	event, ok := resolveEvent(w, r, db)
	if !ok {
		return
	}

	if !requireGroupMember(w, r, db, event.GroupID, userID) {
		return
	}

	meta, err := buildEventWithMeta(db, r, event, userID)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to fetch event details")
		return
	}

	RespondSuccess(w, http.StatusOK, "Event retrieved", meta)
}

func EventRSVPHandler(w http.ResponseWriter, r *http.Request, db *database.DataBase) {
	if r.Method != http.MethodPost {
		RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := GetUserIDFromContext(r)
	if !ok {
		RespondError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	event, ok := resolveEvent(w, r, db)
	if !ok {
		return
	}

	if !requireGroupMember(w, r, db, event.GroupID, userID) {
		return
	}

	var req EventRSVPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Response != "going" && req.Response != "not_going" {
		RespondError(w, http.StatusBadRequest, "Response must be 'going' or 'not_going'")
		return
	}

	if err := db.CreateEventRSVP(r.Context(), event.ID, userID, req.Response); err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to save response")
		return
	}

	meta, err := buildEventWithMeta(db, r, event, userID)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to fetch event details")
		return
	}

	RespondSuccess(w, http.StatusOK, "RSVP updated", meta)
}
