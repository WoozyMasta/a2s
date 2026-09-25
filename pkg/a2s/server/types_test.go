// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package server_test

import (
	"context"
	"testing"

	"github.com/woozymasta/a2s/pkg/a2s"
	"github.com/woozymasta/a2s/pkg/a2s/server"
	"github.com/woozymasta/a2s/pkg/a2s/transport"
)

func ExampleHandlerFunc() {
	var handler server.Handler = server.HandlerFunc(func(
		context.Context,
		*server.Request,
	) (server.Response, error) {
		return server.InfoResponse{Info: a2s.Info{Name: "example"}}, nil
	})

	_, _ = handler.Handle(context.Background(), &server.Request{
		Query: transport.Request{Type: transport.InfoRequest},
	})
}

func ExamplePacketResponse() {
	var response server.Response = server.PacketResponse{
		Packet: transport.Packet{Type: transport.ResponseInfo},
	}

	_, _ = response.(server.PacketResponse)
}

func TestHandlerFunc(t *testing.T) {
	want := server.ErrDrop
	handler := server.HandlerFunc(func(context.Context, *server.Request) (server.Response, error) {
		return nil, want
	})

	_, err := handler.Handle(context.Background(), nil)
	if err != want {
		t.Fatalf("Handle() error = %v, want %v", err, want)
	}
}
