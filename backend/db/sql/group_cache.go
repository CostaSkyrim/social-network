package database

import "time"

// groupCacheEntry is the Redis representation of a Group.
//
// Group hides its internal numeric ids from API responses (json:"-"), so
// marshalling it directly into the cache drops ID and CreatorID. A cache hit
// would then hand back a group whose CreatorID is 0, silently breaking the
// creator short-circuit in the group authorization checks. This DTO keeps the
// internal ids so a cache round-trip is lossless.
//
// Keep this in sync with Group and bump cache.GroupKey when the shape changes,
// otherwise already-stored entries decode into zero-valued ids.
type groupCacheEntry struct {
	ID            int64      `json:"group_id"`
	UUID          string     `json:"id"`
	CreatorID     int64      `json:"creator_id"`
	CreatorUUID   string     `json:"creator_uuid"`
	Title         string     `json:"title"`
	Description   string     `json:"description"`
	AvatarPath    *string    `json:"avatar_path,omitempty"`
	LastMessageAt *time.Time `json:"last_message_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

func groupToCache(g *Group) groupCacheEntry {
	return groupCacheEntry{
		ID:            g.ID,
		UUID:          g.UUID,
		CreatorID:     g.CreatorID,
		CreatorUUID:   g.CreatorUUID,
		Title:         g.Title,
		Description:   g.Description,
		AvatarPath:    g.AvatarPath,
		LastMessageAt: g.LastMessageAt,
		CreatedAt:     g.CreatedAt,
		UpdatedAt:     g.UpdatedAt,
	}
}

func groupFromCache(e groupCacheEntry) *Group {
	return &Group{
		ID:            e.ID,
		UUID:          e.UUID,
		CreatorID:     e.CreatorID,
		CreatorUUID:   e.CreatorUUID,
		Title:         e.Title,
		Description:   e.Description,
		AvatarPath:    e.AvatarPath,
		LastMessageAt: e.LastMessageAt,
		CreatedAt:     e.CreatedAt,
		UpdatedAt:     e.UpdatedAt,
	}
}

// groupMemberCacheEntry is the Redis representation of a group membership.
//
// GroupMember deliberately hides its internal numeric ids from API responses
// (they are tagged `json:"-"`), which means marshalling a GroupMember straight
// into the cache silently drops them. On the next cache hit the members come
// back with UserID == 0, and every authorization check that compares
// `m.UserID == userID` (group posts, events, chat, invites) denies accepted
// members. This DTO keeps the internal ids so a cache round-trip is lossless.
//
// Keep this in sync with GroupMember and bump cache.GroupMembersKey when the
// shape changes, otherwise already-stored entries decode into zero-valued ids.
type groupMemberCacheEntry struct {
	ID            int64     `json:"member_id"`
	GroupID       int64     `json:"group_id"`
	UserID        int64     `json:"user_id"`
	Status        string    `json:"status"`
	InvitedBy     int64     `json:"invited_by"`
	InvitedByUUID *string   `json:"invited_by_uuid,omitempty"`
	JoinedAt      time.Time `json:"joined_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	User          *User     `json:"user,omitempty"`
}

// groupMembersToCache converts members read from the database into their
// cache representation.
func groupMembersToCache(members []GroupMember) []groupMemberCacheEntry {
	entries := make([]groupMemberCacheEntry, 0, len(members))
	for _, m := range members {
		entries = append(entries, groupMemberCacheEntry{
			ID:            m.ID,
			GroupID:       m.GroupID,
			UserID:        m.UserID,
			Status:        m.Status,
			InvitedBy:     m.InvitedBy,
			InvitedByUUID: m.InvitedByUUID,
			JoinedAt:      m.JoinedAt,
			UpdatedAt:     m.UpdatedAt,
			User:          m.User,
		})
	}
	return entries
}

// groupMembersFromCache restores members from their cache representation,
// including the internal ids that GroupMember hides from JSON.
func groupMembersFromCache(entries []groupMemberCacheEntry) []GroupMember {
	members := make([]GroupMember, 0, len(entries))
	for _, e := range entries {
		m := GroupMember{
			ID:            e.ID,
			GroupID:       e.GroupID,
			UserID:        e.UserID,
			Status:        e.Status,
			InvitedBy:     e.InvitedBy,
			InvitedByUUID: e.InvitedByUUID,
			JoinedAt:      e.JoinedAt,
			UpdatedAt:     e.UpdatedAt,
			User:          e.User,
		}
		// The nested User hides its numeric ID as well; the membership record
		// is the source of truth for it (see GetGroupMembers).
		if m.User != nil {
			m.User.ID = e.UserID
		}
		members = append(members, m)
	}
	return members
}
