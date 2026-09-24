// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package server_test

import (
	"context"
	"net"
	"time"

	"github.com/woozymasta/a2s/pkg/a2s"
	"github.com/woozymasta/a2s/pkg/a2s/server"
	"github.com/woozymasta/a2s/pkg/a3sb"
)

func ExampleNew_server() {
	handler := server.HandlerFunc(func(_ context.Context, request *server.Request) (server.Response, error) {
		if request.Query.Type != a2s.InfoRequest {
			return nil, server.ErrDrop
		}

		return server.InfoResponse{Info: a2s.Info{
			Format: a2s.InfoFormat(a2s.ResponseInfo),
			Name:   "Example A2S server",
		}}, nil
	})

	serverInstance, err := server.New(handler)
	if err != nil {
		panic(err)
	}

	conn, err := net.ListenPacket("udp", ":0")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	serveErr := make(chan error, 1)
	go func() { serveErr <- serverInstance.Serve(conn) }()

	defer serverInstance.Shutdown(context.Background())
	_ = serveErr
}

func ExampleState() {
	state := server.NewState()
	if err := state.Store(server.Snapshot{
		Info: &a2s.Info{
			Format: a2s.InfoFormat(a2s.ResponseInfo),
			Name:   "Initial name",
		},
	}); err != nil {
		panic(err)
	}

	// Store publishes a complete replacement atomically.
	if err := state.Store(server.Snapshot{
		Info: &a2s.Info{
			Format: a2s.InfoFormat(a2s.ResponseInfo),
			Name:   "Updated name",
		},
	}); err != nil {
		panic(err)
	}

	serverInstance, err := server.New(state)
	if err != nil {
		panic(err)
	}
	_ = serverInstance
}

func ExampleHandlerFunc_custom() {
	handler := server.HandlerFunc(func(_ context.Context, request *server.Request) (server.Response, error) {
		switch request.Query.Type {
		case a2s.InfoRequest:
			return server.InfoResponse{Info: a2s.Info{
				Format: a2s.InfoFormat(a2s.ResponseInfo),
				Name:   "Custom handler",
			}}, nil

		case a2s.RulesRequest:
			return server.RulesResponse{Rules: a2s.Rules{
				{Name: "motd", Value: "Welcome"},
			}}, nil

		default:
			return nil, server.ErrDrop
		}
	})

	serverInstance, err := server.New(handler)
	if err != nil {
		panic(err)
	}
	_ = serverInstance
}

func ExampleRulesResponse_a3sb() {
	binaryData, err := a3sb.AppendBinary(nil, a3sb.Rules{Layout: a3sb.LayoutArma3})
	if err != nil {
		panic(err)
	}
	escaped := a3sb.AppendEscapeSequences(nil, binaryData)
	rules, err := a3sb.EncodePages(escaped, 0)
	if err != nil {
		panic(err)
	}

	handler := server.HandlerFunc(func(_ context.Context, request *server.Request) (server.Response, error) {
		if request.Query.Type != a2s.RulesRequest {
			return nil, server.ErrDrop
		}
		return server.RulesResponse{Rules: rules}, nil
	})

	serverInstance, err := server.New(handler)
	if err != nil {
		panic(err)
	}
	_ = serverInstance
}

func ExamplePacketResponse_proxy() {
	upstream, err := a2s.NewWithString("upstream.example:27015", a2s.WithTimeout(time.Second))
	if err != nil {
		panic(err)
	}
	defer upstream.Close()

	handler := server.HandlerFunc(func(ctx context.Context, request *server.Request) (server.Response, error) {
		packet, _, err := upstream.Query(ctx, request.Query.Type)
		if err != nil {
			return nil, err
		}

		return server.PacketResponse{Packet: packet}, nil
	})

	serverInstance, err := server.New(handler)
	if err != nil {
		panic(err)
	}
	_ = serverInstance
}

func ExampleInfoResponse_proxy() {
	upstream, err := a2s.NewWithString("upstream.example:27015", a2s.WithTimeout(time.Second))
	if err != nil {
		panic(err)
	}
	defer upstream.Close()

	handler := server.HandlerFunc(func(ctx context.Context, request *server.Request) (server.Response, error) {
		if request.Query.Type != a2s.InfoRequest {
			return nil, server.ErrDrop
		}

		info, err := upstream.GetInfo(ctx)
		if err != nil {
			return nil, err
		}
		info.Name = "Modified by proxy"

		return server.InfoResponse{Info: *info}, nil
	})

	serverInstance, err := server.New(handler)
	if err != nil {
		panic(err)
	}
	_ = serverInstance
}
