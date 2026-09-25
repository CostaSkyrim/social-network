package handlers

import (
	"net/http"
	"testing"
)

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

// TestCanRejoinOrReinvite covers re-inviting someone who previously declined or
// left the group (both stored as "declined"), while refusing to clobber an
// accepted, pending, or invited membership.
func TestCanRejoinOrReinvite(t *testing.T) {
	tests := []struct {
		status string
		want   bool
	}{
		{status: "declined", want: true},
		{status: "accepted", want: false},
		{status: "pending", want: false},
		{status: "invited", want: false},
		{status: "", want: false},
	}

	for _, tc := range tests {
		t.Run(tc.status, func(t *testing.T) {
			if got := canRejoinOrReinvite(tc.status); got != tc.want {
				t.Errorf("canRejoinOrReinvite(%q) = %v, want %v", tc.status, got, tc.want)
			}
		})
	}
}

// TestRemovableMemberError covers who may kick a member. Only the creator may,
// the creator is not removable, and only accepted members can be kicked.
func TestRemovableMemberError(t *testing.T) {
	const (
		creator = "creator-uuid"
		member  = "member-uuid"
		other   = "other-uuid"
	)

	tests := []struct {
		name   string
		actor  string
		target string
		status string
		want   int
	}{
		{name: "creator removes an accepted member", actor: creator, target: member, status: "accepted", want: 0},
		{name: "non-creator cannot remove", actor: other, target: member, status: "accepted", want: http.StatusForbidden},
		{name: "unauthenticated cannot remove", actor: "", target: member, status: "accepted", want: http.StatusForbidden},
		{name: "creator cannot be removed", actor: creator, target: creator, status: "accepted", want: http.StatusBadRequest},
		{name: "non-member cannot be removed", actor: creator, target: member, status: "", want: http.StatusNotFound},
		{name: "pending member cannot be kicked", actor: creator, target: member, status: "pending", want: http.StatusConflict},
		{name: "invited member cannot be kicked", actor: creator, target: member, status: "invited", want: http.StatusConflict},
		{name: "declined member cannot be kicked", actor: creator, target: member, status: "declined", want: http.StatusConflict},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, msg := removableMemberError(tc.actor, creator, tc.target, tc.status)
			if got != tc.want {
				t.Errorf("removableMemberError(%q, creator, %q, %q) status = %d, want %d",
					tc.actor, tc.target, tc.status, got, tc.want)
			}
			if (msg == "") != (got == 0) {
				t.Errorf("message presence = %q does not match status %d", msg, got)
			}
		})
	}
}
