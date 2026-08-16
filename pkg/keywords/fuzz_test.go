// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package keywords

import (
	"strings"
	"testing"

	"github.com/woozymasta/a2s/pkg/appid"
)

func FuzzParseKeywords(f *testing.F) {
	f.Add([]byte("battleye,port2303,c-21--52"))
	f.Fuzz(func(t *testing.T, data []byte) {
		values := strings.Split(string(data), ",")
		_, _ = Parse(appid.Arma3, values)
		_, _ = Parse(appid.DayZ, values)
	})
}

func FuzzParseCoordinates(f *testing.F) {
	f.Add([]byte("-21--52"))
	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = parseCoordinates(string(data))
	})
}
