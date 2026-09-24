// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

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
		{name: "unknown platform", got: Platform("x").String(), want: "x"},
		{name: "known platform", got: OSWLinux.String(), want: "Linux"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.got != test.want {
				t.Fatalf("String() = %q, want %q", test.got, test.want)
			}
		})
	}
}
