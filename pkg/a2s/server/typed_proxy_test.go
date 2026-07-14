package server

import (
	"context"
	"errors"
	"net"
	"reflect"
	"testing"
	"time"

	"github.com/woozymasta/a2s/pkg/a2s"
)

func TestTypedModifyingProxyPreservesInfoFields(t *testing.T) {
	gameID := uint64(123456789)
	original := a2s.Info{
		Format:       a2s.InfoFormat(a2s.ResponseInfo),
		Protocol:     17,
		Name:         "upstream name",
		Map:          "upstream_map",
		Folder:       "game",
		Game:         "Test Game",
		Version:      "1.2.3",
		SourceTVName: "spectator",
		Keywords:     []string{"one", "two"},
		SteamID:      76561198000000001,
		AppID:        730,
		GameID:       &gameID,
		Port:         27015,
		SourceTVPort: 27020,
		Players:      12,
		MaxPlayers:   32,
		Bots:         2,
		ServerType:   a2s.ServerType('d'),
		Environment:  a2s.Environment('l'),
		Visibility:   true,
		VAC:          true,
	}

	upstreamProvider := &proxyChallengeProvider{token: a2s.Challenge{1, 2, 3, 4}}
	upstreamHandler := HandlerFunc(func(_ context.Context, request *Request) (Response, error) {
		if request.Query.Type != a2s.InfoRequest {
			return nil, ErrDrop
		}
		return InfoResponse{Info: original}, nil
	})
	upstream, upstreamConn, upstreamErr := newProxyServer(
		t,
		upstreamHandler,
		upstreamProvider,
		1248,
	)
	defer upstreamConn.Close()
	defer upstream.Shutdown(context.Background())

	upstreamClient, err := a2s.NewWithAddr(
		upstreamConn.LocalAddr().(*net.UDPAddr),
		a2s.WithTimeout(time.Second),
	)
	if err != nil {
		t.Fatalf("NewWithAddr(upstream) error = %v", err)
	}
	defer upstreamClient.Close()

	upstreamInfo, err := upstreamClient.GetInfo(context.Background())
	if err != nil {
		t.Fatalf("upstream GetInfo() error = %v", err)
	}

	downstreamProvider := &proxyChallengeProvider{token: a2s.Challenge{5, 6, 7, 8}}
	downstreamHandler := HandlerFunc(func(ctx context.Context, request *Request) (Response, error) {
		if request.Query.Type != a2s.InfoRequest {
			return nil, ErrDrop
		}
		info, err := upstreamClient.GetInfo(ctx)
		if err != nil {
			return nil, err
		}
		info.Name = "modified downstream name"
		return InfoResponse{Info: *info}, nil
	})
	downstream, downstreamConn, downstreamErr := newProxyServer(
		t,
		downstreamHandler,
		downstreamProvider,
		1248,
	)
	defer downstreamConn.Close()
	defer downstream.Shutdown(context.Background())

	downstreamClient, err := a2s.NewWithAddr(
		downstreamConn.LocalAddr().(*net.UDPAddr),
		a2s.WithTimeout(time.Second),
	)
	if err != nil {
		t.Fatalf("NewWithAddr(downstream) error = %v", err)
	}
	defer downstreamClient.Close()

	got, err := downstreamClient.GetInfo(context.Background())
	if err != nil {
		t.Fatalf("downstream GetInfo() error = %v", err)
	}
	want := *upstreamInfo
	want.Name = "modified downstream name"
	if !reflect.DeepEqual(*got, want) {
		t.Fatalf("modified info = %#v, want %#v", *got, want)
	}

	unchanged, err := upstreamClient.GetInfo(context.Background())
	if err != nil {
		t.Fatalf("upstream follow-up GetInfo() error = %v", err)
	}
	if !reflect.DeepEqual(*unchanged, *upstreamInfo) {
		t.Fatalf("upstream info changed: got %#v, want %#v", *unchanged, *upstreamInfo)
	}

	if err := downstream.Shutdown(context.Background()); err != nil {
		t.Fatalf("downstream Shutdown() error = %v", err)
	}
	if err := <-downstreamErr; !errors.Is(err, ErrServerClosed) {
		t.Fatalf("downstream Serve() error = %v, want ErrServerClosed", err)
	}
	if err := upstream.Shutdown(context.Background()); err != nil {
		t.Fatalf("upstream Shutdown() error = %v", err)
	}
	if err := <-upstreamErr; !errors.Is(err, ErrServerClosed) {
		t.Fatalf("upstream Serve() error = %v, want ErrServerClosed", err)
	}
}
