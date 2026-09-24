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
