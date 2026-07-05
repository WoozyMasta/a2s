package server

import (
	"bytes"
	"context"
	"errors"
	"net"
	"sync/atomic"
	"testing"
	"time"

	"github.com/woozymasta/a2s/pkg/a2s"
)

func TestServerServesTypedQueriesThroughA2SClient(t *testing.T) {
	server, err := New(HandlerFunc(func(_ context.Context, request *Request) (Response, error) {
		switch request.Query.Type {
		case a2s.InfoRequest:
			return InfoResponse{Info: a2s.Info{
				Format:   a2s.InfoFormat(a2s.ResponseInfo),
				Protocol: 17,
				Name:     "test server",
				Map:      "test_map",
				Folder:   "test",
				Game:     "Test Game",
				Version:  "1.0",
			}}, nil
		case a2s.PlayerRequest:
			return PlayersResponse{Players: []a2s.Player{{Name: "player", Score: 3, Duration: time.Second}}}, nil
		case a2s.RulesRequest:
			return RulesResponse{Rules: a2s.Rules{{Name: "hostname", Value: "test server"}}}, nil
		default:
			return nil, ErrDrop
		}
	}), WithWorkers(2))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	conn, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("ListenPacket() error = %v", err)
	}
	serveErr := make(chan error, 1)
	go func() {
		serveErr <- server.Serve(conn)
	}()

	client, err := a2s.NewWithAddr(conn.LocalAddr().(*net.UDPAddr), a2s.WithTimeout(time.Second))
	if err != nil {
		_ = conn.Close()
		<-serveErr
		t.Fatalf("NewWithAddr() error = %v", err)
	}
	defer client.Close()

	info, err := client.GetInfo(context.Background())
	if err != nil {
		t.Fatalf("GetInfo() error = %v", err)
	}
	if info.Name != "test server" || info.Map != "test_map" {
		t.Fatalf("GetInfo() = %#v", info)
	}

	players, err := client.GetPlayers(context.Background())
	if err != nil {
		t.Fatalf("GetPlayers() error = %v", err)
	}
	if len(players) != 1 || players[0].Name != "player" {
		t.Fatalf("GetPlayers() = %#v", players)
	}

	rules, err := client.GetRules(context.Background())
	if err != nil {
		t.Fatalf("GetRules() error = %v", err)
	}
	if len(rules) != 1 || rules[0].Value != "test server" {
		t.Fatalf("GetRules() = %#v", rules)
	}

	if err := conn.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if err := <-serveErr; err == nil {
		t.Fatal("Serve() returned nil after connection close")
	}
}

func TestServerDropsOversizedDatagrams(t *testing.T) {
	var calls atomic.Int32
	server, err := New(
		HandlerFunc(func(context.Context, *Request) (Response, error) {
			calls.Add(1)
			return PacketResponse{Packet: a2s.Packet{Type: a2s.ResponsePing}}, nil
		}),
		WithChallengePolicy(NoChallengePolicy()),
		WithMaxRequestSize(8),
		WithWorkers(1),
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	conn, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("ListenPacket() error = %v", err)
	}
	serveErr := make(chan error, 1)
	go func() { serveErr <- server.Serve(conn) }()

	client, err := net.DialUDP("udp", nil, conn.LocalAddr().(*net.UDPAddr))
	if err != nil {
		_ = conn.Close()
		<-serveErr
		t.Fatalf("DialUDP() error = %v", err)
	}
	defer client.Close()

	if _, err := client.Write(bytes.Repeat([]byte{0xFF}, 9)); err != nil {
		t.Fatalf("Write(oversized) error = %v", err)
	}
	request, err := a2s.AppendRequest(nil, a2s.Request{Type: a2s.PingRequest})
	if err != nil {
		t.Fatalf("AppendRequest() error = %v", err)
	}
	if _, err := client.Write(request); err != nil {
		t.Fatalf("Write(valid) error = %v", err)
	}

	buffer := make([]byte, 128)
	if err := client.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
		t.Fatalf("SetReadDeadline() error = %v", err)
	}
	if _, err := client.Read(buffer); err != nil {
		t.Fatalf("Read(response) error = %v", err)
	}
	if got := calls.Load(); got != 1 {
		t.Fatalf("handler calls = %d, want 1", got)
	}

	_ = conn.Close()
	<-serveErr
}

func TestServerInvokesHandlerConcurrently(t *testing.T) {
	entered := make(chan struct{}, 2)
	release := make(chan struct{})
	var active atomic.Int32
	var maximum atomic.Int32
	server, err := New(
		HandlerFunc(func(context.Context, *Request) (Response, error) {
			current := active.Add(1)
			for {
				previous := maximum.Load()
				if current <= previous || maximum.CompareAndSwap(previous, current) {
					break
				}
			}
			entered <- struct{}{}
			<-release
			active.Add(-1)
			return PacketResponse{Packet: a2s.Packet{Type: a2s.ResponsePing}}, nil
		}),
		WithChallengePolicy(NoChallengePolicy()),
		WithWorkers(2),
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	conn, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("ListenPacket() error = %v", err)
	}
	serveErr := make(chan error, 1)
	go func() { serveErr <- server.Serve(conn) }()

	clients := make([]*net.UDPConn, 2)
	request, err := a2s.AppendRequest(nil, a2s.Request{Type: a2s.PingRequest})
	if err != nil {
		t.Fatalf("AppendRequest() error = %v", err)
	}
	for index := range clients {
		clients[index], err = net.DialUDP("udp", nil, conn.LocalAddr().(*net.UDPAddr))
		if err != nil {
			for _, client := range clients[:index] {
				_ = client.Close()
			}
			_ = conn.Close()
			<-serveErr
			t.Fatalf("DialUDP() error = %v", err)
		}
		defer clients[index].Close()
		if _, err := clients[index].Write(request); err != nil {
			t.Fatalf("Write() error = %v", err)
		}
	}

	for range clients {
		select {
		case <-entered:
		case <-time.After(time.Second):
			t.Fatal("handler did not receive concurrent requests")
		}
	}
	if got := maximum.Load(); got < 2 {
		t.Fatalf("maximum active handlers = %d, want at least 2", got)
	}
	close(release)

	_ = conn.Close()
	if err := <-serveErr; err == nil {
		t.Fatal("Serve() returned nil after connection close")
	}
}

func TestNewServerRejectsInvalidOptions(t *testing.T) {
	handler := HandlerFunc(func(context.Context, *Request) (Response, error) {
		return nil, nil
	})

	for name, options := range map[string][]Option{
		"workers":    {WithWorkers(0)},
		"request":    {WithMaxRequestSize(0)},
		"policy":     {WithChallengePolicy(nil)},
		"provider":   {WithChallengeProvider(nil)},
		"packetizer": {WithSourcePacketizer(nil)},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := New(handler, options...)
			if !errors.Is(err, ErrServer) {
				t.Fatalf("New() error = %v, want ErrServer", err)
			}
		})
	}
}
