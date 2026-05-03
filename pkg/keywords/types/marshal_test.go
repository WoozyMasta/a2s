//go:build !a2s_no_marshal

package types

import (
	"encoding/json"
	"testing"
)

func TestUnknownValuesJSONPreserveRawValue(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  string
	}{
		{name: "server language", value: ServerLang(65599), want: `"Unknown(65599)"`},
		{name: "server state", value: ServerState(10), want: `"Unknown(10)"`},
		{name: "game type", value: GameType("custom_mode"), want: `"custom_mode"`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			encoded, err := json.Marshal(test.value)
			if err != nil {
				t.Fatalf("json.Marshal returned error: %v", err)
			}
			if string(encoded) != test.want {
				t.Fatalf("JSON = %s, want %s", encoded, test.want)
			}
		})
	}
}
