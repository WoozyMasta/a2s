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

func TestServerShutdownCancelsHandlersAndKeepsConnectionOpen(t *testing.T) {
	started := make(chan struct{})
	server, err := New(
		HandlerFunc(func(ctx context.Context, _ *Request) (Response, error) {
			close(started)
			<-ctx.Done()
			return nil, ctx.Err()
		}),
		WithChallengePolicy(NoChallengePolicy()),
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
	request, err := a2s.AppendRequest(nil, a2s.Request{Type: a2s.PingRequest})
	if err != nil {
		t.Fatalf("AppendRequest() error = %v", err)
	}
	if _, err := client.Write(request); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("handler did not start")
	}

	if err := server.Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown() error = %v", err)
	}
	if err := server.Shutdown(context.Background()); err != nil {
		t.Fatalf("second Shutdown() error = %v", err)
	}
	if err := <-serveErr; !errors.Is(err, ErrServerClosed) {
		t.Fatalf("Serve() error = %v, want ErrServerClosed", err)
	}
	if err := conn.Close(); err != nil {
		t.Fatalf("PacketConn was closed by Serve: %v", err)
	}
}

func TestServerContextCancellationStopsServe(t *testing.T) {
	server, err := New(
		HandlerFunc(func(context.Context, *Request) (Response, error) {
			return PacketResponse{Packet: a2s.Packet{Type: a2s.ResponsePing}}, nil
		}),
		WithChallengePolicy(NoChallengePolicy()),
		WithWorkers(1),
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	conn, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("ListenPacket() error = %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	serveErr := make(chan error, 1)
	go func() { serveErr <- server.ServeContext(ctx, conn) }()

	cancel()
	if err := <-serveErr; !errors.Is(err, context.Canceled) {
		t.Fatalf("ServeContext() error = %v, want context.Canceled", err)
	}
	if err := conn.Close(); err != nil {
		t.Fatalf("PacketConn was closed by ServeContext: %v", err)
	}
}

func TestServerRejectsConcurrentServe(t *testing.T) {
	started := make(chan struct{})
	server, err := New(
		HandlerFunc(func(context.Context, *Request) (Response, error) {
			close(started)
			return PacketResponse{Packet: a2s.Packet{Type: a2s.ResponsePing}}, nil
		}),
		WithChallengePolicy(NoChallengePolicy()),
		WithWorkers(1),
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	first, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("ListenPacket(first) error = %v", err)
	}
	firstErr := make(chan error, 1)
	go func() { firstErr <- server.Serve(first) }()
	client, err := net.DialUDP("udp", nil, first.LocalAddr().(*net.UDPAddr))
	if err != nil {
		_ = first.Close()
		<-firstErr
		t.Fatalf("DialUDP() error = %v", err)
	}
	defer client.Close()
	request, err := a2s.AppendRequest(nil, a2s.Request{Type: a2s.PingRequest})
	if err != nil {
		_ = first.Close()
		<-firstErr
		t.Fatalf("AppendRequest() error = %v", err)
	}
	if _, err := client.Write(request); err != nil {
		_ = first.Close()
		<-firstErr
		t.Fatalf("Write() error = %v", err)
	}
	select {
	case <-started:
	case <-time.After(time.Second):
		_ = first.Close()
		<-firstErr
		t.Fatal("first server did not start")
	}

	second, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		_ = first.Close()
		<-firstErr
		t.Fatalf("ListenPacket(second) error = %v", err)
	}
	defer second.Close()
	if err := server.Serve(second); !errors.Is(err, ErrServerRunning) {
		t.Fatalf("second Serve() error = %v, want ErrServerRunning", err)
	}

	if err := server.Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown() error = %v", err)
	}
	if err := <-firstErr; !errors.Is(err, ErrServerClosed) {
		t.Fatalf("first Serve() error = %v, want ErrServerClosed", err)
	}
	if err := first.Close(); err != nil {
		t.Fatalf("first PacketConn close error = %v", err)
	}
}

func TestServerRecoversHandlerPanicAndContinues(t *testing.T) {
	var calls atomic.Int32
	reports := make(chan PanicReport, 1)
	server, err := New(
		HandlerFunc(func(context.Context, *Request) (Response, error) {
			if calls.Add(1) == 1 {
				panic("handler failure")
			}
			return PacketResponse{Packet: a2s.Packet{Type: a2s.ResponsePing}}, nil
		}),
		WithChallengePolicy(NoChallengePolicy()),
		WithPanicReporter(func(_ context.Context, report PanicReport) {
			reports <- report
		}),
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
	request, err := a2s.AppendRequest(nil, a2s.Request{Type: a2s.PingRequest})
	if err != nil {
		t.Fatalf("AppendRequest() error = %v", err)
	}
	if _, err := client.Write(request); err != nil {
		t.Fatalf("Write(first) error = %v", err)
	}

	select {
	case report := <-reports:
		if report.Value != "handler failure" {
			t.Fatalf("panic value = %#v", report.Value)
		}
		if len(report.Stack) == 0 {
			t.Fatal("panic stack is empty")
		}
	case <-time.After(time.Second):
		t.Fatal("panic was not reported")
	}

	if _, err := client.Write(request); err != nil {
		t.Fatalf("Write(second) error = %v", err)
	}
	if err := client.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
		t.Fatalf("SetReadDeadline() error = %v", err)
	}
	buffer := make([]byte, 128)
	if _, err := client.Read(buffer); err != nil {
		t.Fatalf("Read(second response) error = %v", err)
	}

	if err := server.Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown() error = %v", err)
	}
	if err := <-serveErr; !errors.Is(err, ErrServerClosed) {
		t.Fatalf("Serve() error = %v, want ErrServerClosed", err)
	}
	if err := conn.Close(); err != nil {
		t.Fatalf("PacketConn close error = %v", err)
	}
}

func TestServerDefaultPanicReporterIsSafe(t *testing.T) {
	server, err := New(
		HandlerFunc(func(context.Context, *Request) (Response, error) {
			panic("unreported failure")
		}),
		WithChallengePolicy(NoChallengePolicy()),
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
	request, err := a2s.AppendRequest(nil, a2s.Request{Type: a2s.PingRequest})
	if err != nil {
		t.Fatalf("AppendRequest() error = %v", err)
	}
	if _, err := client.Write(request); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if err := client.SetReadDeadline(time.Now().Add(500 * time.Millisecond)); err != nil {
		t.Fatalf("SetReadDeadline() error = %v", err)
	}
	if _, err := client.Read(make([]byte, 128)); err == nil {
		t.Fatal("Read() unexpectedly received a response")
	}

	if err := server.Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown() error = %v", err)
	}
	if err := <-serveErr; !errors.Is(err, ErrServerClosed) {
		t.Fatalf("Serve() error = %v, want ErrServerClosed", err)
	}
	if err := conn.Close(); err != nil {
		t.Fatalf("PacketConn close error = %v", err)
	}
}
