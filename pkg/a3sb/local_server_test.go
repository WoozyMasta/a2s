package a3sb_test

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/woozymasta/a2s/pkg/a2s"
	"github.com/woozymasta/a2s/pkg/a2s/server"
	"github.com/woozymasta/a2s/pkg/a3sb"
)

func TestA3SBClientDecodesLocalEncodedServerResponse(t *testing.T) {
	binaryData, err := a3sb.AppendBinary(nil, a3sb.Rules{
		Layout:      a3sb.LayoutDayZ,
		Version:     2,
		Description: "local DayZ server",
		Mods: []a3sb.Mod{
			{Name: "test mod", ID: 123456, Hash: 0x12345678, IDLength: 4},
		},
	})
	if err != nil {
		t.Fatalf("AppendBinary() error = %v", err)
	}
	escaped := a3sb.AppendEscapeSequences(nil, binaryData)
	rules, err := a3sb.EncodePages(escaped, 0)
	if err != nil {
		t.Fatalf("EncodePages() error = %v", err)
	}

	serverInstance, err := server.New(
		server.HandlerFunc(func(_ context.Context, request *server.Request) (server.Response, error) {
			if request.Query.Type != a2s.RulesRequest {
				return nil, server.ErrDrop
			}
			return server.RulesResponse{Rules: rules}, nil
		}),
		server.WithWorkers(1),
	)
	if err != nil {
		t.Fatalf("server.New() error = %v", err)
	}

	conn, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("ListenPacket() error = %v", err)
	}
	serveErr := make(chan error, 1)
	go func() { serveErr <- serverInstance.Serve(conn) }()

	baseClient, err := a2s.NewWithAddr(conn.LocalAddr().(*net.UDPAddr), a2s.WithTimeout(time.Second))
	if err != nil {
		_ = conn.Close()
		<-serveErr
		t.Fatalf("NewWithAddr() error = %v", err)
	}
	defer baseClient.Close()
	client := &a3sb.Client{Client: baseClient}
	decoded, err := client.GetRulesDayZ(context.Background())
	if err != nil {
		t.Fatalf("GetRulesDayZ() error = %v", err)
	}
	if decoded.Version != 2 || decoded.Description != "local DayZ server" {
		t.Fatalf("decoded rules = %#v", decoded)
	}
	if len(decoded.Mods) != 1 || decoded.Mods[0].Name != "test mod" {
		t.Fatalf("decoded mods = %#v", decoded.Mods)
	}

	if err := serverInstance.Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown() error = %v", err)
	}
	if err := <-serveErr; !errors.Is(err, server.ErrServerClosed) {
		t.Fatalf("Serve() error = %v, want ErrServerClosed", err)
	}
	if err := conn.Close(); err != nil {
		t.Fatalf("PacketConn close error = %v", err)
	}
}
