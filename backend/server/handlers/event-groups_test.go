package handlers

import "testing"

func TestNormalizeNickname(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{name: "plain handle", raw: "alicej", want: "alicej"},
		{name: "leading at sign is dropped", raw: "@alicej", want: "alicej"},
		{name: "repeated at signs are dropped", raw: "@@alicej", want: "alicej"},
		{name: "surrounding whitespace is trimmed", raw: "  alicej  ", want: "alicej"},
		{name: "at sign surrounded by whitespace", raw: "  @alicej ", want: "alicej"},
		{name: "blank input stays blank", raw: "   ", want: ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := normalizeNickname(tc.raw); got != tc.want {
				t.Errorf("normalizeNickname(%q) = %q, want %q", tc.raw, got, tc.want)
			}
		})
	}
}

// TestCanDecideMembership pins down who may accept or decline a membership.
//
// Regression: invitations were decided by the group creator, so the invited
// user could never accept or decline their own invitation. A join request
// (pending) belongs to the creator; an invitation (invited) belongs to the
// invited user.
func TestCanDecideMembership(t *testing.T) {
	const (
		creator = "creator-uuid"
		target  = "target-uuid"
		other   = "other-uuid"
	)

	tests := []struct {
		name   string
		status string
		actor  string
		want   bool
	}{
		{name: "creator decides a join request", status: "pending", actor: creator, want: true},
		{name: "other member cannot decide a join request", status: "pending", actor: other, want: false},
		{name: "requester cannot decide their own join request", status: "pending", actor: target, want: false},
		{name: "invited user decides their invitation", status: "invited", actor: target, want: true},
		{name: "creator cannot decide an invitation", status: "invited", actor: creator, want: false},
		{name: "other member cannot decide an invitation", status: "invited", actor: other, want: false},
		{name: "accepted membership has no pending decision", status: "accepted", actor: creator, want: false},
		{name: "declined membership has no pending decision", status: "declined", actor: target, want: false},
		{name: "unauthenticated actor is never allowed", status: "pending", actor: "", want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := canDecideMembership(tc.status, creator, target, tc.actor); got != tc.want {
				t.Errorf("canDecideMembership(%q, creator, target, %q) = %v, want %v",
					tc.status, tc.actor, got, tc.want)
			}
		})
	}
}
