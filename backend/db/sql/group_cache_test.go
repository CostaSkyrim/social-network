package database

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestGroupCacheRoundTripPreservesInternalIDs(t *testing.T) {
	avatar := "/images/groups/g001.png"
	lastMessage := time.Date(2026, 9, 22, 9, 30, 0, 0, time.UTC)
	created := time.Date(2026, 9, 21, 16, 4, 46, 0, time.UTC)
	updated := time.Date(2026, 9, 21, 16, 4, 46, 0, time.UTC)

	tests := []struct {
		name  string
		group *Group
	}{
		{
			name: "group with avatar and last message",
			group: &Group{
				ID:            1,
				UUID:          "g0010000-0000-4000-8000-000000000001",
				CreatorID:     3,
				CreatorUUID:   "c3d4e5f6-a7b8-9012-cdef-123456789012",
				Title:         "Golang Enthusiasts",
				Description:   "A group for Go programming lovers.",
				AvatarPath:    &avatar,
				LastMessageAt: &lastMessage,
				CreatedAt:     created,
				UpdatedAt:     updated,
			},
		},
		{
			name: "group without optional fields",
			group: &Group{
				ID:          2,
				UUID:        "g0020000-0000-4000-8000-000000000002",
				CreatorID:   1,
				CreatorUUID: "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
				Title:       "Photography Club",
				CreatedAt:   created,
				UpdatedAt:   updated,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			raw, err := json.Marshal(groupToCache(tc.group))
			if err != nil {
				t.Fatalf("marshal group cache entry: %v", err)
			}
			// The numeric group and creator ids must survive the round-trip;
			// without them authorization treats the creator as a stranger.
			if !strings.Contains(string(raw), `"group_id"`) || !strings.Contains(string(raw), `"creator_id"`) {
				t.Fatalf("cache payload drops the internal group ids: %s", raw)
			}

			var decoded groupCacheEntry
			if err := json.Unmarshal(raw, &decoded); err != nil {
				t.Fatalf("unmarshal group cache entry: %v", err)
			}

			got := groupFromCache(decoded)
			if got.ID != tc.group.ID {
				t.Errorf("ID: got %d, want %d", got.ID, tc.group.ID)
			}
			if got.CreatorID != tc.group.CreatorID {
				t.Errorf("CreatorID: got %d, want %d", got.CreatorID, tc.group.CreatorID)
			}
			if got.UUID != tc.group.UUID {
				t.Errorf("UUID: got %q, want %q", got.UUID, tc.group.UUID)
			}
			if got.CreatorUUID != tc.group.CreatorUUID {
				t.Errorf("CreatorUUID: got %q, want %q", got.CreatorUUID, tc.group.CreatorUUID)
			}
			if got.Title != tc.group.Title {
				t.Errorf("Title: got %q, want %q", got.Title, tc.group.Title)
			}
			if got.Description != tc.group.Description {
				t.Errorf("Description: got %q, want %q", got.Description, tc.group.Description)
			}
			if (got.AvatarPath == nil) != (tc.group.AvatarPath == nil) {
				t.Errorf("AvatarPath nil-ness: got %v, want %v", got.AvatarPath, tc.group.AvatarPath)
			} else if got.AvatarPath != nil && *got.AvatarPath != *tc.group.AvatarPath {
				t.Errorf("AvatarPath: got %q, want %q", *got.AvatarPath, *tc.group.AvatarPath)
			}
			if got.LastMessageAt == nil || !got.LastMessageAt.Equal(*tc.group.LastMessageAt) {
				if tc.group.LastMessageAt != nil {
					t.Errorf("LastMessageAt: got %v, want %v", got.LastMessageAt, *tc.group.LastMessageAt)
				}
			}
			if !got.CreatedAt.Equal(tc.group.CreatedAt) {
				t.Errorf("CreatedAt: got %v, want %v", got.CreatedAt, tc.group.CreatedAt)
			}
			if !got.UpdatedAt.Equal(tc.group.UpdatedAt) {
				t.Errorf("UpdatedAt: got %v, want %v", got.UpdatedAt, tc.group.UpdatedAt)
			}
		})
	}
}

// TestGroupMemberCacheRoundTripPreservesInternalIDs is a regression test for the
// bug where GetGroupMembers cached GroupMember values directly. GroupMember
// hides its internal numeric ids (json:"-"), so a cache hit returned members
// with UserID == 0 and every membership check (group posts, events, chat,
// invites) rejected accepted members with a 403.
func TestGroupMemberCacheRoundTripPreservesInternalIDs(t *testing.T) {
	nickname := "carolw"
	invitedByUUID := "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
	joined := time.Date(2026, 9, 21, 16, 4, 46, 0, time.UTC)

	tests := []struct {
		name    string
		members []GroupMember
	}{
		{
			name: "accepted member with nested user",
			members: []GroupMember{
				{
					ID:            7,
					GroupID:       1,
					UserID:        3,
					Status:        "accepted",
					InvitedBy:     3,
					InvitedByUUID: &invitedByUUID,
					JoinedAt:      joined,
					User: &User{
						ID:        3,
						UUID:      "c3d4e5f6-a7b8-9012-cdef-123456789012",
						Email:     "carol@example.com",
						FirstName: "Carol",
						LastName:  "Williams",
						Nickname:  &nickname,
					},
				},
			},
		},
		{
			name: "pending member without inviter",
			members: []GroupMember{
				{
					ID:       9,
					GroupID:  2,
					UserID:   6,
					Status:   "pending",
					JoinedAt: joined,
					User: &User{
						ID:        6,
						UUID:      "f6a7b8c9-d0e1-2345-fabc-456789012345",
						FirstName: "Frank",
						LastName:  "Miller",
					},
				},
			},
		},
		{
			name:    "empty group",
			members: []GroupMember{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			entries := groupMembersToCache(tc.members)

			// Simulate the Redis JSON round-trip.
			raw, err := json.Marshal(entries)
			if err != nil {
				t.Fatalf("marshal cache entries: %v", err)
			}

			// Guard against reintroducing the lossy payload: the internal
			// numeric ids must be present in the stored JSON.
			if len(tc.members) > 0 && !strings.Contains(string(raw), `"user_id"`) {
				t.Fatalf("cache payload drops the numeric user id: %s", raw)
			}

			var decoded []groupMemberCacheEntry
			if err := json.Unmarshal(raw, &decoded); err != nil {
				t.Fatalf("unmarshal cache entries: %v", err)
			}

			if len(decoded) != len(tc.members) {
				t.Fatalf("member count: got %d, want %d", len(decoded), len(tc.members))
			}

			got := groupMembersFromCache(decoded)
			for i, want := range tc.members {
				if got[i].ID != want.ID {
					t.Errorf("member %d ID: got %d, want %d", i, got[i].ID, want.ID)
				}
				if got[i].GroupID != want.GroupID {
					t.Errorf("member %d GroupID: got %d, want %d", i, got[i].GroupID, want.GroupID)
				}
				if got[i].UserID != want.UserID {
					t.Errorf("member %d UserID: got %d, want %d", i, got[i].UserID, want.UserID)
				}
				if got[i].InvitedBy != want.InvitedBy {
					t.Errorf("member %d InvitedBy: got %d, want %d", i, got[i].InvitedBy, want.InvitedBy)
				}
				if got[i].Status != want.Status {
					t.Errorf("member %d Status: got %q, want %q", i, got[i].Status, want.Status)
				}
				if !got[i].JoinedAt.Equal(want.JoinedAt) {
					t.Errorf("member %d JoinedAt: got %v, want %v", i, got[i].JoinedAt, want.JoinedAt)
				}
				if want.InvitedByUUID != nil {
					if got[i].InvitedByUUID == nil || *got[i].InvitedByUUID != *want.InvitedByUUID {
						t.Errorf("member %d InvitedByUUID: got %v, want %v", i, got[i].InvitedByUUID, *want.InvitedByUUID)
					}
				}
				if want.User == nil {
					continue
				}
				if got[i].User == nil {
					t.Fatalf("member %d nested User is nil", i)
				}
				// The nested user's numeric ID is what presence lookups use.
				if got[i].User.ID != want.User.ID {
					t.Errorf("member %d User.ID: got %d, want %d", i, got[i].User.ID, want.User.ID)
				}
				if got[i].User.UUID != want.User.UUID {
					t.Errorf("member %d User.UUID: got %q, want %q", i, got[i].User.UUID, want.User.UUID)
				}
			}
		})
	}
}
