// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/woozymasta/a2s/pkg/a2s"
	"github.com/woozymasta/a2s/pkg/a2s/server"
	"github.com/woozymasta/a2s/pkg/a3sb"
)

func TestExecuteAllJSONEmitsOneStructuredDocument(t *testing.T) {
	fixture := newProxyStartupFixture(t, a2s.ResponseInfo, false, true, true)
	output := &bytes.Buffer{}
	command := &AllCommand{
		Args: ServerArgs{
			Host: fixture.addr.IP.String(),
			Port: strconv.Itoa(fixture.addr.Port),
		},
		OutputOptions: OutputOptions{Format: "json"},
	}

	if err := executeAll(NewApplication(output, output), command, proxyStartupClientOptions()); err != nil {
		t.Fatalf("executeAll() error = %v", err)
	}

	var document map[string]json.RawMessage
	if err := json.Unmarshal(output.Bytes(), &document); err != nil {
		t.Fatalf("json.Unmarshal() error = %v; output = %q", err, output.String())
	}

	if len(document) != 3 {
		t.Fatalf("top-level keys = %v, want exactly info, rules, players", document)
	}
	for _, key := range []string{"info", "rules", "players"} {
		if _, ok := document[key]; !ok {
			t.Fatalf("top-level JSON has no %q field", key)
		}
	}

	var info map[string]any
	if err := json.Unmarshal(document["info"], &info); err != nil {
		t.Fatalf("info JSON error = %v", err)
	}
	if info["name"] != "proxy test" {
		t.Fatalf("info.name = %#v, want proxy test", info["name"])
	}

	var rules map[string]any
	if err := json.Unmarshal(document["rules"], &rules); err != nil {
		t.Fatalf("rules JSON error = %v", err)
	}
	if rules["hostname"] != "proxy test" {
		t.Fatalf("rules.hostname = %#v, want proxy test", rules["hostname"])
	}

	var players []any
	if err := json.Unmarshal(document["players"], &players); err != nil {
		t.Fatalf("players JSON error = %v", err)
	}
	if players == nil {
		t.Fatal("players JSON is null, want an array")
	}
}

func TestExecuteAllJSONPreservesA3SBRulesShape(t *testing.T) {
	binaryData, err := a3sb.AppendBinary(nil, a3sb.Rules{
		Layout:      a3sb.LayoutDayZ,
		Version:     2,
		Description: "A3SB test",
	})
	if err != nil {
		t.Fatalf("AppendBinary() error = %v", err)
	}
	rules, err := a3sb.EncodePages(a3sb.AppendEscapeSequences(nil, binaryData), 0)
	if err != nil {
		t.Fatalf("EncodePages() error = %v", err)
	}

	serverInstance, err := server.New(server.HandlerFunc(func(_ context.Context, request *server.Request) (server.Response, error) {
		switch request.Query.Type {
		case a2s.InfoRequest:
			return server.InfoResponse{Info: a2s.Info{
				Format: a2s.InfoFormat(a2s.ResponseInfo),
				Name:   "A3SB test",
			}}, nil

		case a2s.RulesRequest:
			return server.RulesResponse{Rules: rules}, nil

		case a2s.PlayerRequest:
			return server.PlayersResponse{Players: []a2s.Player{}}, nil

		default:
			return nil, server.ErrDrop
		}
	}))
	if err != nil {
		t.Fatalf("server.New() error = %v", err)
	}

	conn, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("ListenPacket() error = %v", err)
	}
	serveErr := make(chan error, 1)
	go func() { serveErr <- serverInstance.Serve(conn) }()
	t.Cleanup(func() {
		if err := serverInstance.Shutdown(context.Background()); err != nil {
			t.Errorf("Shutdown() error = %v", err)
		}
		if err := <-serveErr; !errors.Is(err, server.ErrServerClosed) {
			t.Errorf("Serve() error = %v, want ErrServerClosed", err)
		}
		if err := conn.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	})

	address := conn.LocalAddr().(*net.UDPAddr)
	output := &bytes.Buffer{}
	command := &AllCommand{
		Args: ServerArgs{
			Host: address.IP.String(),
			Port: strconv.Itoa(address.Port),
		},
		OutputOptions: OutputOptions{Format: "json"},
		RulesOptions:  RulesOptions{Game: "dayz"},
	}
	if err := executeAll(NewApplication(output, output), command, ClientOptions{
		Timeout: Duration(time.Second),
		Buffer:  a2s.DefaultBufferSize,
	}); err != nil {
		t.Fatalf("executeAll() error = %v", err)
	}

	var document map[string]json.RawMessage
	if err := json.Unmarshal(output.Bytes(), &document); err != nil {
		t.Fatalf("json.Unmarshal() error = %v; output = %q", err, output.String())
	}
	var rulesJSON map[string]any
	if err := json.Unmarshal(document["rules"], &rulesJSON); err != nil {
		t.Fatalf("rules JSON error = %v", err)
	}
	if rulesJSON["version"] != float64(2) || rulesJSON["description"] != "A3SB test" {
		t.Fatalf("rules JSON = %#v, want A3SB fields", rulesJSON)
	}
}
