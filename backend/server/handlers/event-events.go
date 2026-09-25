package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
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
	userUUID, ok := GetUserUUIDFromContext(r)
	if !ok {
		RespondError(w, http.StatusUnauthorized, "Authentication required")
		return false
	}

	group, err := db.GetGroup(r.Context(), groupID)
	if err != nil {
		RespondError(w, http.StatusNotFound, "Group not found")
		return false
	}

	if group.CreatorUUID == userUUID {
		return true
	}

	members, _ := db.GetGroupMembers(r.Context(), groupID)
	for _, m := range members {
		if m.Status == "accepted" && memberUUID(m) == userUUID {
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

	if cfg != nil {
		if msg := validateLengthLimits(strings.TrimSpace(req.Description), "Description", 0, cfg.DatabaseConfiguration.Limits.MaxDescription); msg != "" {
			return time.Time{}, msg
		}
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
	userUUID, _ := GetUserUUIDFromContext(r)

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

	// Notify all accepted members except the creator. Notifications key off the
	// recipient's internal id, so the member's UUID is resolved to one here.
	group, _ := db.GetGroup(r.Context(), groupID)
	members, _ := db.GetGroupMembers(r.Context(), groupID)
	for _, m := range members {
		id := memberUUID(m)
		if id == "" || id == userUUID || m.Status != "accepted" {
			continue
		}
		member, err := db.GetUserByUUID(r.Context(), id)
		if err != nil {
			continue
		}
		sendNotification(db, member.ID, userID, NotifNewEvent, "New event '"+req.Title+"' in group: "+group.Title, &eventID, &group.UUID)
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

// userGroupEventWithMeta is an event from a user's groups, enriched with RSVP
// counts and (for the responder) the user's own response.
type userGroupEventWithMeta struct {
	*database.UserGroupEvent
	Going      int     `json:"going"`
	NotGoing   int     `json:"not_going"`
	Total      int     `json:"total"`
	MyResponse *string `json:"my_response,omitempty"`
}

func GetUserGroupEventsHandler(w http.ResponseWriter, r *http.Request, db *database.DataBase) {
	if r.Method != http.MethodGet {
		RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := GetUserIDFromContext(r)
	if !ok {
		RespondError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	limit := 10
	offset := 0
	if v, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && v > 0 && v <= maxPageLimit() {
		limit = v
	}
	if v, err := strconv.Atoi(r.URL.Query().Get("offset")); err == nil && v >= 0 {
		offset = v
	}

	events, err := db.GetUserGroupEvents(r.Context(), userID, limit, offset)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to fetch events")
		return
	}

	metaList := []*userGroupEventWithMeta{}
	for _, e := range events {
		counts, err := db.GetEventResponseCounts(r.Context(), e.ID)
		if err != nil {
			RespondError(w, http.StatusInternalServerError, "Failed to fetch event details")
			return
		}
		myResp, err := db.GetEventResponseByUser(r.Context(), e.ID, userID)
		if err != nil {
			RespondError(w, http.StatusInternalServerError, "Failed to fetch event details")
			return
		}

		meta := &userGroupEventWithMeta{
			UserGroupEvent: e,
			Going:          counts["going"],
			NotGoing:       counts["not_going"],
			Total:          counts["going"] + counts["not_going"],
		}
		if myResp != nil {
			meta.MyResponse = &myResp.Response
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

// canManageEvent reports whether the user may edit or cancel an event (its
// creator, or the group's creator).
func canManageEvent(r *http.Request, db *database.DataBase, event *database.Event) bool {
	userUUID, ok := GetUserUUIDFromContext(r)
	if !ok {
		return false
	}
	if event.CreatorUUID == userUUID {
		return true
	}
	group, err := db.GetGroup(r.Context(), event.GroupID)
	if err != nil {
		return false
	}
	return group.CreatorUUID == userUUID
}

func UpdateEventHandler(w http.ResponseWriter, r *http.Request, db *database.DataBase) {
	if r.Method != http.MethodPut && r.Method != http.MethodPost {
		RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	if _, ok := GetUserIDFromContext(r); !ok {
		RespondError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	event, ok := resolveEvent(w, r, db)
	if !ok {
		return
	}

	if !canManageEvent(r, db, event) {
		RespondError(w, http.StatusForbidden, "You cannot edit this event")
		return
	}

	var req CreateEventRequest
	imagePath := event.ImagePath // keep unless replaced/removed

	if isMultipart(r) {
		if err := parseMultipartForm(r); err != nil {
			RespondError(w, http.StatusBadRequest, "Invalid multipart form")
			return
		}
		req.Title = r.FormValue("title")
		req.Description = r.FormValue("description")
		req.EventDatetime = r.FormValue("event_datetime")
		if _, has := r.MultipartForm.File["image"]; has {
			path, ok := SaveMultipartImage(w, r)
			if !ok {
				return
			}
			imagePath = &path
		} else if removeImageRequested(r) {
			imagePath = nil
		}
	} else {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			RespondError(w, http.StatusBadRequest, "Invalid request body")
			return
		}
		imagePath = req.ImagePath
	}

	eventDatetime, validationErr := validateEventRequest(&req)
	if validationErr != "" {
		RespondError(w, http.StatusBadRequest, validationErr)
		return
	}

	if err := db.UpdateEvent(r.Context(), event.ID, req.Title, req.Description, imagePath, eventDatetime); err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to update event")
		return
	}

	RespondSuccess(w, http.StatusOK, "Event updated", map[string]interface{}{
		"id": event.UUID,
	})
}

func DeleteEventHandler(w http.ResponseWriter, r *http.Request, db *database.DataBase) {
	if r.Method != http.MethodDelete && r.Method != http.MethodPost {
		RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	if _, ok := GetUserIDFromContext(r); !ok {
		RespondError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	event, ok := resolveEvent(w, r, db)
	if !ok {
		return
	}

	if !canManageEvent(r, db, event) {
		RespondError(w, http.StatusForbidden, "You cannot cancel this event")
		return
	}

	if err := db.CancelEvent(r.Context(), event.ID); err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to cancel event")
		return
	}

	RespondSuccess(w, http.StatusOK, "Event cancelled", nil)
}
