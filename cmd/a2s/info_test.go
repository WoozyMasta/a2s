package main

import (
	"encoding/json"
	"testing"

	"github.com/woozymasta/a2s/pkg/a2s"
	"github.com/woozymasta/a2s/pkg/appid"
)

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

func TestPrintInfoJSONPreservesGenericKeywords(t *testing.T) {
	output := captureRulesStdout(t, func() {
		printInfoJSON(&a2s.Info{
			ID:       1337,
			Keywords: []string{"foo", "bar"},
		}, NewFormatter("json"))
	})

	var result map[string]any
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}

	keywords, ok := result["keywords"].([]any)
	if !ok {
		t.Fatalf("keywords = %#v, want JSON array", result["keywords"])
	}
	if len(keywords) != 2 || keywords[0] != "foo" || keywords[1] != "bar" {
		t.Fatalf("keywords = %#v, want [foo bar]", keywords)
	}
}

func TestPrintInfoJSONUsesTypedKeywordsForSupportedGames(t *testing.T) {
	tests := []struct {
		name      string
		id        uint64
		keywords  []string
		wantField string
		wantValue any
	}{
		{
			name:      "arma3",
			id:        appid.Arma3,
			keywords:  []string{"bt", "r218"},
			wantField: "required_version",
			wantValue: float64(218),
		},
		{
			name:      "dayz",
			id:        appid.DayZ,
			keywords:  []string{"battleye", "shard001"},
			wantField: "shard",
			wantValue: "001",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureRulesStdout(t, func() {
				printInfoJSON(&a2s.Info{ID: tt.id, Keywords: tt.keywords}, NewFormatter("json"))
			})

			var result map[string]any
			if err := json.Unmarshal([]byte(output), &result); err != nil {
				t.Fatalf("unmarshal output: %v", err)
			}

			parsed, ok := result["keywords"].(map[string]any)
			if !ok {
				t.Fatalf("keywords = %#v, want typed object", result["keywords"])
			}
			if parsed[tt.wantField] != tt.wantValue {
				t.Fatalf("keywords[%q] = %#v, want %#v", tt.wantField, parsed[tt.wantField], tt.wantValue)
			}
		})
	}
}
