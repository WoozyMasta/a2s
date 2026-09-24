// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package main

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestPrintRulesKeepsNativeA2SJSONShape(t *testing.T) {
	var output bytes.Buffer
	if err := printRules(
		NewApplication(&output, nil),
		map[string]any{
			"hostname": "test server",
			"players":  2,
		},
		nil,
		NewFormatter("json", &output),
	); err != nil {
		t.Fatalf("printRules() error = %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(output.Bytes(), &result); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if result["hostname"] != "test server" {
		t.Fatalf("hostname = %#v, want %q", result["hostname"], "test server")
	}
	if result["players"] != float64(2) {
		t.Fatalf("players = %#v, want 2", result["players"])
	}
	if _, ok := result["a2s"]; ok {
		t.Fatal("native A2S output unexpectedly contains an a2s wrapper")
	}
	if _, ok := result["extra_rules"]; ok {
		t.Fatal("native A2S output unexpectedly contains an extra_rules wrapper")
	}
}
