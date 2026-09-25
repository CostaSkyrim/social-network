package database

import (
	"encoding/json"
	"testing"
	"time"
)

// TestGroupMemberCacheRoundTripPreservesUUIDs covers the group-member cache.
//
// Members are cached as []GroupMember, and GroupMember carries no internal
// numeric ids: identity is the nested user's UUID plus the inviter's UUID. A
// JSON round-trip is all the Redis cache does, so it must preserve those UUIDs.
// Every consumer — membership checks, presence, the websocket group broadcast,
// and notification fan-out — reads them, so losing any would silently break
// authorization or delivery.
func TestGroupMemberCacheRoundTripPreservesUUIDs(t *testing.T) {
	nickname := "carolw"
	invitedBy := "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
	joined := time.Date(2026, 9, 21, 16, 4, 46, 0, time.UTC)

	tests := []struct {
		name    string
		members []GroupMember
	}{
		{
			name: "accepted member with nested user",
			members: []GroupMember{
				{
					User: &User{
						UUID:      "c3d4e5f6-a7b8-9012-cdef-123456789012",
						FirstName: "Carol",
						LastName:  "Williams",
						Nickname:  &nickname,
					},
					Status:        "accepted",
					InvitedByUUID: &invitedBy,
					JoinedAt:      joined,
				},
			},
		},
		{
			name: "member without a nested user",
			members: []GroupMember{
				{Status: "pending", JoinedAt: joined},
			},
		},
		{
			name:    "empty group",
			members: []GroupMember{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Simulate the Redis JSON round-trip.
			raw, err := json.Marshal(tc.members)
			if err != nil {
				t.Fatalf("marshal members: %v", err)
			}

			var got []GroupMember
			if err := json.Unmarshal(raw, &got); err != nil {
				t.Fatalf("unmarshal members: %v", err)
			}

			if len(got) != len(tc.members) {
				t.Fatalf("member count: got %d, want %d", len(got), len(tc.members))
			}

			for i, want := range tc.members {
				if got[i].Status != want.Status {
					t.Errorf("member %d Status: got %q, want %q", i, got[i].Status, want.Status)
				}
				if !got[i].JoinedAt.Equal(want.JoinedAt) {
					t.Errorf("member %d JoinedAt: got %v, want %v", i, got[i].JoinedAt, want.JoinedAt)
				}

				if (got[i].InvitedByUUID == nil) != (want.InvitedByUUID == nil) {
					t.Fatalf("member %d InvitedByUUID nil-ness mismatch", i)
				}
				if want.InvitedByUUID != nil && *got[i].InvitedByUUID != *want.InvitedByUUID {
					t.Errorf("member %d InvitedByUUID: got %q, want %q", i, *got[i].InvitedByUUID, *want.InvitedByUUID)
				}

				if (got[i].User == nil) != (want.User == nil) {
					t.Fatalf("member %d nested User nil-ness mismatch", i)
				}
				if want.User == nil {
					continue
				}
				if got[i].User.UUID != want.User.UUID {
					t.Errorf("member %d User.UUID: got %q, want %q", i, got[i].User.UUID, want.User.UUID)
				}
				if got[i].User.FirstName != want.User.FirstName || got[i].User.LastName != want.User.LastName {
					t.Errorf("member %d User name: got %q %q, want %q %q",
						i, got[i].User.FirstName, got[i].User.LastName, want.User.FirstName, want.User.LastName)
				}
				if (got[i].User.Nickname == nil) != (want.User.Nickname == nil) {
					t.Fatalf("member %d User.Nickname nil-ness mismatch", i)
				}
				if want.User.Nickname != nil && *got[i].User.Nickname != *want.User.Nickname {
					t.Errorf("member %d User.Nickname: got %q, want %q", i, *got[i].User.Nickname, *want.User.Nickname)
				}
			}
		})
	}
}
