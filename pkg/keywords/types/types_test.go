package types

import (
	"testing"
)

func TestUnknownValuesPreserveRawValue(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		{name: "server language", got: ServerLang(65599).String(), want: "Unknown(65599)"},
		{name: "server state", got: ServerState(10).String(), want: "Unknown(10)"},
		{name: "game type", got: GameType("custom_mode").String(), want: "custom_mode"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.got != test.want {
				t.Fatalf("String() = %q, want %q", test.got, test.want)
			}
		})
	}
}
