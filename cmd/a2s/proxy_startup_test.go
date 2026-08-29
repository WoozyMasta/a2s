// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package main

import (
	"context"
	"encoding/binary"
	"net"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/woozymasta/a2s/pkg/a2s"
	"github.com/woozymasta/a2s/pkg/a2s/server"
)

func TestPrepareProxyStartupSelectsPolicyPacketizerAndCapabilities(t *testing.T) {
	tests := []struct {
		name             string
		infoType         a2s.ResponseType
		infoChallenge    bool
		wantSecurePolicy bool
		wantPacketizer   any
	}{
		{
			name:             "source without info challenge",
			infoType:         a2s.ResponseInfo,
			wantPacketizer:   &server.SourcePacketizer{},
			wantSecurePolicy: false,
		},
		{
			name:             "goldsource with info challenge",
			infoType:         a2s.ResponseInfoGoldSource,
			infoChallenge:    true,
			wantPacketizer:   &server.GoldSourcePacketizer{},
			wantSecurePolicy: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixture := newProxyStartupFixture(t, tt.infoType, tt.infoChallenge, true, true)
			command := proxyStartupCommand(fixture, []string{"auto"})

			preparation, err := prepareProxyStartup(
				context.Background(),
				command,
				proxyStartupClientOptions(),
			)
			if err != nil {
				t.Fatalf("prepareProxyStartup() error = %v", err)
			}
			defer preparation.close()

			if preparation.infoMeta.UsedChallenge != tt.infoChallenge {
				t.Fatalf(
					"UsedChallenge = %v, want %v",
					preparation.infoMeta.UsedChallenge,
					tt.infoChallenge,
				)
			}
			if preparation.policy.Required(a2s.InfoRequest) != tt.wantSecurePolicy {
				t.Fatalf("INFO policy = %v, want %v", preparation.policy.Required(a2s.InfoRequest), tt.wantSecurePolicy)
			}
			if !preparation.policy.Required(a2s.PlayerRequest) || !preparation.policy.Required(a2s.RulesRequest) {
				t.Fatal("legacy and secure policies must require PLAYER/RULES challenges")
			}
			if _, ok := preparation.packetizer.(interface {
				Packetize([]byte) ([][]byte, error)
			}); !ok {
				t.Fatalf("packetizer does not implement Packetize: %T", preparation.packetizer)
			}
			switch tt.wantPacketizer.(type) {
			case *server.SourcePacketizer:
				if _, ok := preparation.packetizer.(*server.SourcePacketizer); !ok {
					t.Fatalf("packetizer = %T, want SourcePacketizer", preparation.packetizer)
				}
			case *server.GoldSourcePacketizer:
				if _, ok := preparation.packetizer.(*server.GoldSourcePacketizer); !ok {
					t.Fatalf("packetizer = %T, want GoldSourcePacketizer", preparation.packetizer)
				}
			}

			for _, query := range []a2s.QueryType{a2s.InfoRequest, a2s.PlayerRequest, a2s.RulesRequest} {
				if !preparation.cache.Enabled(query) {
					t.Fatalf("cache does not enable %s", proxyQueryName(query))
				}
				if _, ok := preparation.cache.Load(query); !ok {
					t.Fatalf("cache has no startup packet for %s", proxyQueryName(query))
				}
			}
		})
	}
}

func TestPrepareProxyStartupSkipsUnsupportedAutoQueries(t *testing.T) {
	fixture := newProxyStartupFixture(t, a2s.ResponseInfo, false, true, false)
	command := proxyStartupCommand(fixture, []string{"auto"})
	clientOptions := proxyStartupClientOptions()
	clientOptions.Timeout = Duration(10 * time.Millisecond)

	preparation, err := prepareProxyStartup(context.Background(), command, clientOptions)
	if err != nil {
		t.Fatalf("prepareProxyStartup() error = %v", err)
	}
	defer preparation.close()

	if !preparation.cache.Enabled(a2s.InfoRequest) {
		t.Fatal("auto cache did not enable INFO")
	}
	if preparation.cache.Enabled(a2s.PlayerRequest) {
		t.Fatal("auto cache enabled unsupported PLAYER query")
	}
	if preparation.cache.Enabled(a2s.RulesRequest) {
		t.Fatal("auto cache enabled unsupported RULES query")
	}
}

func TestPrepareProxyStartupFailsBeforeServingOnInfoError(t *testing.T) {
	fixture := newProxyStartupFixture(t, a2s.ResponseInfo, false, false, false)
	command := proxyStartupCommand(fixture, []string{"info"})
	clientOptions := proxyStartupClientOptions()
	clientOptions.Timeout = Duration(10 * time.Millisecond)

	if _, err := prepareProxyStartup(context.Background(), command, clientOptions); err == nil {
		t.Fatal("prepareProxyStartup() returned nil error for unavailable INFO")
	}
}

func TestExecuteProxyContextServesCachedInfoAndStops(t *testing.T) {
	fixture := newProxyStartupFixture(t, a2s.ResponseInfo, false, true, true)
	command := proxyStartupCommand(fixture, []string{"auto"})
	command.Listen = freeProxyListenAddress(t)
	command.TTL = Duration(time.Hour)
	clientOptions := proxyStartupClientOptions()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	output := &proxyRuntimeTestWriter{}
	app := NewApplication(output, output)
	runtimeErr := make(chan error, 1)
	go func() { runtimeErr <- executeProxyContext(ctx, app, command, clientOptions) }()

	waitForProxyRuntimeOutput(t, output, "proxy listening")

	client, err := a2s.NewWithString(command.Listen, a2s.WithTimeout(time.Second))
	if err != nil {
		t.Fatalf("NewWithString() error = %v", err)
	}
	defer client.Close()

	info, err := client.GetInfo(context.Background())
	if err != nil {
		t.Fatalf("GetInfo() error = %v", err)
	}
	if info.Name != "proxy test" {
		t.Fatalf("info name = %q, want %q", info.Name, "proxy test")
	}
	if _, err := client.GetPing(context.Background()); err != nil {
		t.Fatalf("GetPing() error = %v", err)
	}

	cancel()
	select {
	case err := <-runtimeErr:
		if err != nil {
			t.Fatalf("executeProxyContext() error = %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("executeProxyContext() did not stop after cancellation")
	}
}

func TestExecuteProxyContextStaysLiveWithoutCacheEntries(t *testing.T) {
	fixture := newProxyStartupFixture(t, a2s.ResponseInfo, false, true, false)
	command := proxyStartupCommand(fixture, []string{"players"})
	command.Listen = freeProxyListenAddress(t)
	command.TTL = Duration(time.Hour)
	clientOptions := proxyStartupClientOptions()
	clientOptions.Timeout = Duration(10 * time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	output := &proxyRuntimeTestWriter{}
	runtimeErr := make(chan error, 1)
	go func() {
		runtimeErr <- executeProxyContext(ctx, NewApplication(output, output), command, clientOptions)
	}()

	waitForProxyRuntimeOutput(t, output, "proxy listening")
	select {
	case err := <-runtimeErr:
		t.Fatalf("executeProxyContext() stopped without cache entries: %v", err)
	case <-time.After(100 * time.Millisecond):
	}

	cancel()
	select {
	case err := <-runtimeErr:
		if err != nil {
			t.Fatalf("executeProxyContext() error = %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("executeProxyContext() did not stop after cancellation")
	}
}

func TestExecuteProxyContextRecoversCachedQueryAfterUpstreamOutage(t *testing.T) {
	fixture := newProxyStartupFixture(t, a2s.ResponseInfo, false, true, false)
	command := proxyStartupCommand(fixture, []string{"info"})
	command.Listen = freeProxyListenAddress(t)
	command.TTL = Duration(20 * time.Millisecond)
	command.InactiveTTL = Duration(20 * time.Millisecond)
	command.Retries = 0
	clientOptions := proxyStartupClientOptions()
	clientOptions.Timeout = Duration(10 * time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	output := &proxyRuntimeTestWriter{}
	runtimeErr := make(chan error, 1)
	go func() {
		runtimeErr <- executeProxyContext(ctx, NewApplication(output, output), command, clientOptions)
	}()

	waitForProxyRuntimeOutput(t, output, "proxy listening")
	fixture.setInfoResponse(true)
	client, err := a2s.NewWithString(command.Listen, a2s.WithTimeout(100*time.Millisecond))
	if err != nil {
		t.Fatalf("NewWithString() error = %v", err)
	}
	defer client.Close()
	if _, err := client.GetInfo(context.Background()); err != nil {
		t.Fatalf("initial GetInfo() error = %v", err)
	}

	fixture.setInfoResponse(false)
	waitForProxyRuntimeOutput(t, output, "INFO unavailable")
	if _, err := client.GetInfo(context.Background()); err == nil {
		t.Fatal("GetInfo() succeeded while the cached entry was unavailable")
	}

	fixture.setInfoResponse(true)
	waitForProxyRuntimeOutput(t, output, "INFO recovered")
	if _, err := client.GetInfo(context.Background()); err != nil {
		t.Fatalf("recovered GetInfo() error = %v", err)
	}

	cancel()
	select {
	case err := <-runtimeErr:
		if err != nil {
			t.Fatalf("executeProxyContext() error = %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("executeProxyContext() did not stop after cancellation")
	}
}

func proxyStartupCommand(fixture *proxyStartupFixture, cache []string) *ProxyCommand {
	return &ProxyCommand{
		Args: ServerArgs{
			Host: fixture.addr.IP.String(),
			Port: strconv.Itoa(fixture.addr.Port),
		},
		Listen: ":27016",
		Cache:  cache,
	}
}

func proxyStartupClientOptions() ClientOptions {
	return ClientOptions{
		Timeout: Duration(250 * time.Millisecond),
		Buffer:  8192,
	}
}

func freeProxyListenAddress(t *testing.T) string {
	t.Helper()

	conn, err := net.ListenPacket("udp4", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("ListenPacket() error = %v", err)
	}
	address := conn.LocalAddr().String()
	if err := conn.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	return address
}

func waitForProxyRuntimeOutput(t *testing.T, output *proxyRuntimeTestWriter, text string) {
	t.Helper()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if strings.Contains(output.String(), text) {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}

	t.Fatalf("runtime output does not contain %q: %q", text, output.String())
}

type proxyRuntimeTestWriter struct {
	mu   sync.RWMutex
	text string
}

func (w *proxyRuntimeTestWriter) Write(data []byte) (int, error) {
	w.mu.Lock()
	w.text += string(data)
	w.mu.Unlock()

	return len(data), nil
}

func (w *proxyRuntimeTestWriter) String() string {
	w.mu.RLock()
	defer w.mu.RUnlock()

	return w.text
}

type proxyStartupFixture struct {
	conn          *net.UDPConn
	addr          *net.UDPAddr
	info          []byte
	infoChallenge bool
	respondInfo   atomic.Bool
	respondExtra  atomic.Bool
	done          chan struct{}
}

func newProxyStartupFixture(
	t *testing.T,
	infoType a2s.ResponseType,
	infoChallenge bool,
	respondInfo bool,
	respondExtra bool,
) *proxyStartupFixture {
	t.Helper()

	conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		t.Fatalf("ListenUDP() error = %v", err)
	}

	info, err := proxyStartupInfoPacket(infoType)
	if err != nil {
		_ = conn.Close()
		t.Fatalf("proxyStartupInfoPacket() error = %v", err)
	}

	fixture := &proxyStartupFixture{
		conn:          conn,
		addr:          conn.LocalAddr().(*net.UDPAddr),
		info:          info,
		infoChallenge: infoChallenge,
		done:          make(chan struct{}),
	}
	fixture.respondInfo.Store(respondInfo)
	fixture.respondExtra.Store(respondExtra)
	go fixture.serve()

	t.Cleanup(func() {
		_ = conn.Close()
		<-fixture.done
	})

	return fixture
}

func (f *proxyStartupFixture) serve() {
	defer close(f.done)

	buffer := make([]byte, 64*1024)
	infoChallenged := false
	for {
		n, remote, err := f.conn.ReadFromUDP(buffer)
		if err != nil {
			return
		}
		if n < 5 {
			continue
		}

		query := a2s.QueryType(buffer[4])
		if query == a2s.InfoRequest && f.infoChallenge && !infoChallenged {
			infoChallenged = true
			_, _ = f.conn.WriteToUDP(proxyStartupChallengePacket(), remote)
			continue
		}

		var response []byte
		switch query {
		case a2s.InfoRequest:
			if f.respondInfo.Load() {
				response = f.info
			}

		case a2s.PlayerRequest:
			if f.respondExtra.Load() {
				response, _ = a2s.AppendPlayers(nil, nil)
			}

		case a2s.RulesRequest:
			if f.respondExtra.Load() {
				response, _ = a2s.AppendRules(nil, nil)
			}
		}

		if len(response) > 0 {
			_, _ = f.conn.WriteToUDP(response, remote)
		}
	}
}

func (f *proxyStartupFixture) setInfoResponse(enabled bool) {
	f.respondInfo.Store(enabled)
}

func proxyStartupInfoPacket(responseType a2s.ResponseType) ([]byte, error) {
	info := a2s.Info{
		Format:      a2s.InfoFormat(responseType),
		Name:        "proxy test",
		Map:         "test_map",
		Folder:      "test_game",
		Game:        "Test Game",
		Version:     "1.0",
		ServerType:  a2s.ServerType('d'),
		Environment: a2s.Environment('l'),
	}
	if responseType == a2s.ResponseInfoGoldSource {
		info.Address = "127.0.0.1:27015"
	}

	return a2s.AppendInfo(nil, info)
}

func proxyStartupChallengePacket() []byte {
	packet := make([]byte, 0, 9)
	packet = binary.LittleEndian.AppendUint32(packet, ^uint32(0))
	packet = append(packet, byte(a2s.ResponseChallenge))
	packet = append(packet, 1, 2, 3, 4)

	return packet
}
