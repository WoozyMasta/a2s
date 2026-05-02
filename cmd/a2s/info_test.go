package main

import "testing"

func TestFormatAppID(t *testing.T) {
	tests := []struct {
		name string
		id   uint64
		want string
	}{
		{name: "known", id: 107410, want: "Arma 3 (107410)"},
		{name: "unknown", id: 123456, want: "123456"},
		{name: "game ID wider than AppID", id: 1<<32 + 123, want: "4294967419"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatAppID(tt.id); got != tt.want {
				t.Fatalf("formatAppID(%d) = %q, want %q", tt.id, got, tt.want)
			}
		})
	}
}
