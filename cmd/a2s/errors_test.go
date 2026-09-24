// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package main

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestCommandErrorsAreReturned(t *testing.T) {
	app := NewApplication(nil, nil)
	tests := []struct {
		name string
		call func() error
	}{
		{name: "info", call: func() error { return executeInfo(app, &InfoCommand{}, ClientOptions{}) }},
		{name: "players", call: func() error { return executePlayers(app, &PlayersCommand{}, ClientOptions{}) }},
		{name: "rules", call: func() error { return executeRules(app, &RulesCommand{}, ClientOptions{}) }},
		{name: "all", call: func() error { return executeAll(app, &AllCommand{}, ClientOptions{}) }},
		{name: "ping", call: func() error { return executePing(app, &PingCommand{}, ClientOptions{}) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.call(); err == nil {
				t.Fatal("command returned nil error, want missing-host error")
			}
		})
	}
}

func TestCreateClientReturnsError(t *testing.T) {
	client, err := createClient("", "", Duration(3*time.Second), 8192)
	if err == nil {
		t.Fatal("createClient() error = nil, want invalid-address error")
	}
	if client != nil {
		t.Fatalf("createClient() client = %#v, want nil", client)
	}
}

func TestFormatterPrintJSONReturnsMarshalError(t *testing.T) {
	err := NewFormatter("json", nil).PrintJSON(func() {})
	if err == nil {
		t.Fatal("PrintJSON() error = nil, want marshal error")
	}
}

func TestFriendlyQueryErrorExplainsTimeout(t *testing.T) {
	err := friendlyQueryError(
		NewApplication(nil, nil),
		"error.server_info",
		"failed to get server info",
		context.DeadlineExceeded,
		3*time.Second,
	)
	message := err.Error()

	if !strings.Contains(message, "server did not respond within 3 seconds") {
		t.Fatalf("error = %q, want user-facing timeout explanation", message)
	}
	if strings.Contains(message, "context deadline exceeded") {
		t.Fatalf("error = %q, must not expose low-level context message", message)
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal("friendly timeout error does not preserve its cause")
	}
}
