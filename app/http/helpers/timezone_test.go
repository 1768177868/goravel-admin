package helpers

import "testing"

func TestNormalizeTimeQueryValue(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "url encoded space between date and time",
			input: "2026-07-12+17:00:00",
			want:  "2026-07-12 17:00:00",
		},
		{
			name:  "already normalized datetime",
			input: "2026-07-12 17:00:00",
			want:  "2026-07-12 17:00:00",
		},
		{
			name:  "iso offset should stay unchanged",
			input: "2026-07-12T17:00:00+08:00",
			want:  "2026-07-12T17:00:00+08:00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeTimeQueryValue(tt.input); got != tt.want {
				t.Fatalf("normalizeTimeQueryValue(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestHasExplicitTimezoneOffset(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{name: "zulu", input: "2026-07-12T09:00:00Z", want: true},
		{name: "offset", input: "2026-07-12T17:00:00+08:00", want: true},
		{name: "local datetime", input: "2026-07-12 17:00:00", want: false},
		{name: "url plus datetime", input: "2026-07-12+17:00:00", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasExplicitTimezoneOffset(tt.input); got != tt.want {
				t.Fatalf("hasExplicitTimezoneOffset(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
