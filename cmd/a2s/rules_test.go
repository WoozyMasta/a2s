package main

import (
	"encoding/json"
	"io"
	"os"
	"testing"
)

func TestPrintRulesKeepsNativeA2SJSONShape(t *testing.T) {
	output := captureRulesStdout(t, func() {
		printRules(
			map[string]any{
				"hostname": "test server",
				"players":  2,
			},
			nil,
			NewFormatter("json"),
		)
	})

	var result map[string]any
	if err := json.Unmarshal([]byte(output), &result); err != nil {
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

func captureRulesStdout(t *testing.T, fn func()) string {
	t.Helper()

	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("create stdout pipe: %v", err)
	}

	original := os.Stdout
	os.Stdout = writer
	fn()
	if err := writer.Close(); err != nil {
		t.Fatalf("close stdout writer: %v", err)
	}
	os.Stdout = original

	output, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("read captured stdout: %v", err)
	}
	if err := reader.Close(); err != nil {
		t.Fatalf("close stdout reader: %v", err)
	}

	return string(output)
}
