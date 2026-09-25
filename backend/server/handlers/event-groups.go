package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"social-network/backend/config"
	database "social-network/backend/db/sql"
)

type CreateGroupRequest struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	AvatarPath  *string `json:"avatar_path,omitempty"`
}

type GroupMemberRequest struct {
	UserID string `json:"user_id"`
}

type InviteMemberRequest struct {
	UserID   string `json:"user_id"`
	Nickname string `json:"nickname"`
}

type GroupResponse struct {
	ID          string  `json:"id"`
	CreatorID   int64   `json:"creator_id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	AvatarPath  *string `json:"avatar_path,omitempty"`
	CreatedAt   string  `json:"created_at"`
}

// resolveGroupID looks up a group by its UUID (from the URL path) and returns
// the numeric internal ID. Responds with an error and returns false on failure.
func resolveGroupID(w http.ResponseWriter, r *http.Request, db *database.DataBase) (int64, bool) {
	groupUUID := r.PathValue("id")
	if groupUUID == "" {
		RespondError(w, http.StatusBadRequest, "Group ID required")
		return 0, false
	}

	group, err := db.GetGroupByUUID(r.Context(), groupUUID)
	if err != nil {
		RespondError(w, http.StatusNotFound, "Group not found")
		return 0, false
	}

	return group.ID, true
}

func CreateGroupHandler(w http.ResponseWriter, r *http.Request, db *database.DataBase) {
	if r.Method != http.MethodPost {
		RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := GetUserIDFromContext(r)
	if !ok {
		RespondError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	var req CreateGroupRequest
	if isMultipart(r) {
		if err := parseMultipartForm(r); err != nil {
			RespondError(w, http.StatusBadRequest, "Invalid multipart form")
			return
		}
		req.Title = r.FormValue("title")
		req.Description = r.FormValue("description")
		imgPath, ok := multipartImage(w, r)
		if !ok {
			return
		}
		req.AvatarPath = imgPath
	} else {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			RespondError(w, http.StatusBadRequest, "Invalid request body")
			return
		}
	}

	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" {
		RespondError(w, http.StatusBadRequest, "Title is required")
		return
	}

	if cfg := config.GetConfig(); cfg != nil {
		limits := cfg.DatabaseConfiguration.Limits
		if msg := validateLengthLimits(req.Title, "Title", 0, limits.MaxGroupTitle); msg != "" {
			RespondError(w, http.StatusBadRequest, msg)
			return
		}
		if msg := validateLengthLimits(strings.TrimSpace(req.Description), "Description", 0, limits.MaxDescription); msg != "" {
			RespondError(w, http.StatusBadRequest, msg)
			return
		}
	}

	userUUID, _ := GetUserUUIDFromContext(r)

	group := &database.Group{
		UUID:        generateUUID(),
		CreatorUUID: userUUID,
		Title:       req.Title,
		Description: req.Description,
		AvatarPath:  req.AvatarPath,
	}

	groupID, err := db.CreateGroup(r.Context(), group, userID)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to create group")
		return
	}

	if err := db.AddGroupMember(r.Context(), groupID, userID, &userID, "accepted"); err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to add creator as member")
		return
	}

	RespondSuccess(w, http.StatusCreated, "Group created successfuly", map[string]interface{}{
		"id":    group.UUID,
		"db_id": groupID,
	})
}

func GetGroupHandler(w http.ResponseWriter, r *http.Request, db *database.DataBase) {
	if r.Method != http.MethodGet {
		RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	groupID, ok := resolveGroupID(w, r, db)
	if !ok {
		return
	}

	group, err := db.GetGroup(r.Context(), groupID)
	if err != nil {
		RespondError(w, http.StatusNotFound, "Group not found")
		return
	}

	members, _ := db.GetGroupMembers(r.Context(), groupID)
	enrichMembersPresence(members)

	RespondSuccess(w, http.StatusOK, "Group retrieved", map[string]interface{}{
		"group":   group,
		"members": members,
	})
}

// enrichMembersPresence fills in each member's live online status from the hub.
func enrichMembersPresence(members []database.GroupMember) {
	if GlobalHub == nil {
		return
	}
	for i := range members {
		if members[i].User != nil {
			members[i].User.IsOnline = GlobalHub.IsUserConnectedUUID(members[i].User.UUID)
		}
	}
}

func UpdateGroupHandler(w http.ResponseWriter, r *http.Request, db *database.DataBase) {
	if r.Method != http.MethodPut {
		RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	if _, ok := GetUserIDFromContext(r); !ok {
		RespondError(w, http.StatusUnauthorized, "Authentication required")
		return
	}
	userUUID, _ := GetUserUUIDFromContext(r)

	groupID, ok := resolveGroupID(w, r, db)
	if !ok {
		return
	}

	group, err := db.GetGroup(r.Context(), groupID)
	if err != nil {
		RespondError(w, http.StatusNotFound, "Group not found")
		return
	}

	if group.CreatorUUID != userUUID {
		RespondError(w, http.StatusForbidden, "Only the creator can update the group")
		return
	}

	var req CreateGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" {
		req.Title = group.Title
	}

	if cfg := config.GetConfig(); cfg != nil {
		limits := cfg.DatabaseConfiguration.Limits
		if msg := validateLengthLimits(req.Title, "Title", 0, limits.MaxGroupTitle); msg != "" {
			RespondError(w, http.StatusBadRequest, msg)
			return
		}
		if msg := validateLengthLimits(strings.TrimSpace(req.Description), "Description", 0, limits.MaxDescription); msg != "" {
			RespondError(w, http.StatusBadRequest, msg)
			return
		}
	}

	if err := db.UpdateGroup(r.Context(), groupID, req.Title, req.Description, req.AvatarPath); err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to update group")
		return
	}

	RespondSuccess(w, http.StatusOK, "Group updated successfuly", nil)
}

func DeleteGroupHandler(w http.ResponseWriter, r *http.Request, db *database.DataBase) {
	if r.Method != http.MethodDelete {
		RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := GetUserIDFromContext(r)
	if !ok {
		RespondError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	groupID, ok := resolveGroupID(w, r, db)
	if !ok {
		return
	}

	if err := db.DeleteGroup(r.Context(), groupID, userID); err != nil {
		RespondError(w, http.StatusNotFound, "Group not found or not authorized")
		return
	}

	RespondSuccess(w, http.StatusOK, "Group deleted successfuly", nil)
}

func BrowseGroupsHandler(w http.ResponseWriter, r *http.Request, db *database.DataBase) {
	if r.Method != http.MethodGet {
		RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 20
	offset := 0

	if limitStr != "" {
		if v, err := strconv.Atoi(limitStr); err == nil && v > 0 && v <= 100 {
			limit = v
		}
	}
	if offsetStr != "" {
		if v, err := strconv.Atoi(offsetStr); err == nil && v >= 0 {
			offset = v
		}
	}

	groups, err := db.GetAllGroups(r.Context(), limit, offset)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to retrieve groups")
		return
	}

	if groups == nil {
		groups = []database.Group{}
	}

	RespondSuccess(w, http.StatusOK, "Groups retrieved", groups)
}

func GetUserGroupsHandler(w http.ResponseWriter, r *http.Request, db *database.DataBase) {
	if r.Method != http.MethodGet {
		RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		RespondError(w, http.StatusBadRequest, "user_id query parameter required")
		return
	}

	userID, err := resolveUserID(r, db, userIDStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "Invalid user_id")
		return
	}

	groups, err := db.GetUserGroups(r.Context(), userID)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to get user groups")
		return
	}

	if groups == nil {
		groups = []*database.Group{}
	}

	RespondSuccess(w, http.StatusOK, "User groups retrieved", groups)
}

// normalizeNickname cleans a hand-typed handle: surrounding whitespace and a
// leading "@" are dropped, because nicknames are displayed as "@handle"
// elsewhere in the UI and users copy them verbatim.
func normalizeNickname(raw string) string {
	return strings.TrimLeft(strings.TrimSpace(raw), "@")
}

func InviteToGroupHandler(w http.ResponseWriter, r *http.Request, db *database.DataBase) {
	if r.Method != http.MethodPost {
		RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	currentUserID, ok := GetUserIDFromContext(r)
	if !ok {
		RespondError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	groupID, ok := resolveGroupID(w, r, db)
	if !ok {
		return
	}

	group, err := db.GetGroup(r.Context(), groupID)
	if err != nil {
		RespondError(w, http.StatusNotFound, "Group not found")
		return
	}

	currentUserUUID, _ := GetUserUUIDFromContext(r)

	members, _ := db.GetGroupMembers(r.Context(), groupID)
	isMember := false
	for _, m := range members {
		if m.Status == "accepted" && memberUUID(m) == currentUserUUID {
			isMember = true
			break
		}
	}
	if !isMember && group.CreatorUUID != currentUserUUID {
		RespondError(w, http.StatusForbidden, "Must be a group member to invite")
		return
	}

	var req InviteMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	var targetUser *database.User
	if req.Nickname != "" {
		nickname := normalizeNickname(req.Nickname)
		if nickname == "" {
			RespondError(w, http.StatusBadRequest, "Provide a nickname or user ID to invite")
			return
		}
		targetUser, err = db.GetUserByNickname(r.Context(), nickname)
		if err != nil {
			// Stored nicknames are normally lowercase, but accept any casing
			// the inviter typed.
			targetUser, err = db.GetUserByNicknameFold(r.Context(), nickname)
		}
		if err != nil {
			RespondError(w, http.StatusNotFound, "No user found with nickname "+strconv.Quote(nickname))
			return
		}
	} else if req.UserID != "" {
		targetUser, err = db.GetUserByUUID(r.Context(), req.UserID)
		if err != nil {
			RespondError(w, http.StatusNotFound, "No user found with that id")
			return
		}
	} else {
		RespondError(w, http.StatusBadRequest, "Provide a nickname or user ID to invite")
		return
	}

	for _, m := range members {
		if memberUUID(m) == targetUser.UUID {
			RespondError(w, http.StatusConflict, "User is already a member or has a pending invitation")
			return
		}
	}

	if err := db.AddGroupMember(r.Context(), groupID, targetUser.ID, &currentUserID, "invited"); err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to invite user")
		return
	}

	inviter, _ := db.GetUserByID(r.Context(), currentUserID)
	inviterName := "A user"
	if inviter != nil {
		inviterName = getDisplayName(inviter)
	}

	sendNotification(db, targetUser.ID, currentUserID, NotifGroupInvitation, inviterName+" invited you to join group: "+group.Title, &groupID, &group.UUID)

	RespondSuccess(w, http.StatusOK, "Invitation sent", nil)
}

func RequestJoinGroupHandler(w http.ResponseWriter, r *http.Request, db *database.DataBase) {
	if r.Method != http.MethodPost {
		RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	currentUserID, ok := GetUserIDFromContext(r)
	if !ok {
		RespondError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	groupID, ok := resolveGroupID(w, r, db)
	if !ok {
		return
	}

	group, err := db.GetGroup(r.Context(), groupID)
	if err != nil {
		RespondError(w, http.StatusNotFound, "Group not found")
		return
	}

	currentUserUUID, _ := GetUserUUIDFromContext(r)

	members, _ := db.GetGroupMembers(r.Context(), groupID)
	for _, m := range members {
		if memberUUID(m) == currentUserUUID && m.Status != "declined" {
			if m.Status == "accepted" {
				RespondError(w, http.StatusConflict, "Already a member")
			} else {
				RespondError(w, http.StatusConflict, "Already has a pending request or invitation")
			}
			return
		}
	}

	if err := db.AddGroupMember(r.Context(), groupID, currentUserID, nil, "pending"); err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to send join request")
		return
	}

	requester, _ := db.GetUserByID(r.Context(), currentUserID)
	requesterName := "A user"
	if requester != nil {
		requesterName = getDisplayName(requester)
	}

	// Notify the group creator. Only the creator's public UUID is carried by
	// the group, so resolve the numeric id the notification insert needs.
	if creator, err := db.GetUserByUUID(r.Context(), group.CreatorUUID); err == nil {
		sendNotification(db, creator.ID, currentUserID, NotifGroupJoinRequest, requesterName+" wants to join your group: "+group.Title, &groupID, &group.UUID)
	}

	RespondSuccess(w, http.StatusOK, "Join request sent", nil)
}

// canDecideMembership reports whether the caller identified by actorUUID may
// accept or decline the membership in the given status.
//
// A join request (pending) is the group creator's decision. An invitation
// (invited) is the invited user's own decision — the creator must not accept or
// decline on their behalf. Anything else has no pending decision.
func canDecideMembership(status, creatorUUID, targetUUID, actorUUID string) bool {
	if actorUUID == "" {
		return false
	}
	switch status {
	case "pending":
		return actorUUID == creatorUUID
	case "invited":
		return actorUUID == targetUUID
	default:
		return false
	}
}

// authorizeMembershipDecision checks that the caller may respond to a pending
// membership and returns its current status.
func authorizeMembershipDecision(w http.ResponseWriter, r *http.Request, db *database.DataBase, groupID int64, group *database.Group, targetUser *database.User) (string, bool) {
	currentUserUUID, _ := GetUserUUIDFromContext(r)

	status, err := db.GetGroupMemberStatus(r.Context(), groupID, targetUser.ID)
	if err != nil {
		RespondError(w, http.StatusNotFound, "No pending request or invitation for this user")
		return "", false
	}

	if !canDecideMembership(status, group.CreatorUUID, targetUser.UUID, currentUserUUID) {
		switch status {
		case "pending":
			RespondError(w, http.StatusForbidden, "Only the group creator can respond to join requests")
		case "invited":
			RespondError(w, http.StatusForbidden, "Only the invited user can respond to this invitation")
		default:
			RespondError(w, http.StatusConflict, "No pending request or invitation for this user")
		}
		return "", false
	}

	return status, true
}

func AcceptGroupMemberHandler(w http.ResponseWriter, r *http.Request, db *database.DataBase) {
	if r.Method != http.MethodPost {
		RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	currentUserID, ok := GetUserIDFromContext(r)
	if !ok {
		RespondError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	groupID, ok := resolveGroupID(w, r, db)
	if !ok {
		return
	}

	group, err := db.GetGroup(r.Context(), groupID)
	if err != nil {
		RespondError(w, http.StatusNotFound, "Group not found")
		return
	}

	var req GroupMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	targetUser, err := db.GetUserByUUID(r.Context(), req.UserID)
	if err != nil {
		RespondError(w, http.StatusNotFound, "User not found")
		return
	}

	status, ok := authorizeMembershipDecision(w, r, db, groupID, group, targetUser)
	if !ok {
		return
	}

	if err := db.UpdateMemberStatus(r.Context(), groupID, targetUser.ID, "accepted"); err != nil {
		RespondError(w, http.StatusNotFound, "Member not found")
		return
	}

	if status == "pending" {
		// The creator accepted a join request: update their own notification and
		// tell the requester.
		db.UpdateGroupJoinNotification(r.Context(), group.UUID, targetUser.ID,
			"You have accepted "+getDisplayName(targetUser)+"'s request to join "+group.Title)
		sendNotification(db, targetUser.ID, currentUserID, NotifGroupAccepted,
			"accepted your request to join group: "+group.Title, &groupID, &group.UUID)
	} else {
		// The invitee accepted: update their invitation and let the creator know.
		db.UpdateGroupInvitationNotification(r.Context(), group.UUID, targetUser.ID,
			"You joined group: "+group.Title)
		if creator, err := db.GetUserByUUID(r.Context(), group.CreatorUUID); err == nil {
			sendNotification(db, creator.ID, targetUser.ID, NotifGroupAccepted,
				getDisplayName(targetUser)+" joined your group: "+group.Title, &groupID, &group.UUID)
		}
	}

	RespondSuccess(w, http.StatusOK, "Member accepted", nil)
}

func RejectGroupMemberHandler(w http.ResponseWriter, r *http.Request, db *database.DataBase) {
	if r.Method != http.MethodPost {
		RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	if _, ok := GetUserIDFromContext(r); !ok {
		RespondError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	groupID, ok := resolveGroupID(w, r, db)
	if !ok {
		return
	}

	group, err := db.GetGroup(r.Context(), groupID)
	if err != nil {
		RespondError(w, http.StatusNotFound, "Group not found")
		return
	}

	var req GroupMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	targetUser, err := db.GetUserByUUID(r.Context(), req.UserID)
	if err != nil {
		RespondError(w, http.StatusNotFound, "User not found")
		return
	}

	status, ok := authorizeMembershipDecision(w, r, db, groupID, group, targetUser)
	if !ok {
		return
	}

	if err := db.UpdateMemberStatus(r.Context(), groupID, targetUser.ID, "declined"); err != nil {
		RespondError(w, http.StatusNotFound, "Member not found")
		return
	}

	if status == "pending" {
		db.UpdateGroupJoinNotification(r.Context(), group.UUID, targetUser.ID,
			"You have declined "+getDisplayName(targetUser)+"'s request to join "+group.Title)
	} else {
		db.UpdateGroupInvitationNotification(r.Context(), group.UUID, targetUser.ID,
			"You declined the invitation to join "+group.Title)
	}

	RespondSuccess(w, http.StatusOK, "Member rejected", nil)
}

func LeaveGroupHandler(w http.ResponseWriter, r *http.Request, db *database.DataBase) {
	if r.Method != http.MethodPost {
		RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	currentUserID, ok := GetUserIDFromContext(r)
	if !ok {
		RespondError(w, http.StatusUnauthorized, "Authentication required")
		return
	}
	currentUserUUID, _ := GetUserUUIDFromContext(r)

	groupID, ok := resolveGroupID(w, r, db)
	if !ok {
		return
	}

	group, err := db.GetGroup(r.Context(), groupID)
	if err != nil {
		RespondError(w, http.StatusNotFound, "Group not found")
		return
	}

	if group.CreatorUUID == currentUserUUID {
		RespondError(w, http.StatusBadRequest, "Creator cannot leave the group. Delete it instead.")
		return
	}

	if err := db.UpdateMemberStatus(r.Context(), groupID, currentUserID, "declined"); err != nil {
		RespondError(w, http.StatusNotFound, "Not a member of this group")
		return
	}

	RespondSuccess(w, http.StatusOK, "Left the group", nil)
}

func UpdateGroupAvatarHandler(w http.ResponseWriter, r *http.Request, db *database.DataBase) {
	if r.Method != http.MethodPost {
		RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	if _, ok := GetUserIDFromContext(r); !ok {
		RespondError(w, http.StatusUnauthorized, "Authentication required")
		return
	}
	currentUserUUID, _ := GetUserUUIDFromContext(r)

	groupID, ok := resolveGroupID(w, r, db)
	if !ok {
		return
	}

	group, err := db.GetGroup(r.Context(), groupID)
	if err != nil {
		RespondError(w, http.StatusNotFound, "Group not found")
		return
	}

	if group.CreatorUUID != currentUserUUID {
		RespondError(w, http.StatusForbidden, "Only the group creator can change the avatar")
		return
	}

	avatarPath, ok := SaveMultipartImage(w, r)
	if !ok {
		return
	}

	if err := db.UpdateGroupAvatar(r.Context(), groupID, avatarPath); err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to update group avatar")
		return
	}

	// Remove the previous avatar file now that it's no longer referenced.
	DeleteImageFile(group.AvatarPath)

	RespondSuccess(w, http.StatusOK, "Group avatar updated", map[string]interface{}{
		"avatar_path": avatarPath,
	})
}

func GetGroupMembersHandler(w http.ResponseWriter, r *http.Request, db *database.DataBase) {
	if r.Method != http.MethodGet {
		RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	groupID, ok := resolveGroupID(w, r, db)
	if !ok {
		return
	}

	members, err := db.GetGroupMembers(r.Context(), groupID)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to get members")
		return
	}

	if members == nil {
		members = []database.GroupMember{}
	}

	// Enrich with live presence so the member list can show online status.
	enrichMembersPresence(members)

	RespondSuccess(w, http.StatusOK, "Members retrieved", members)
}
