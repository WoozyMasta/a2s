// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package a2s

import (
	"context"
	"errors"
	"io"
	"net"
	"testing"
)

func TestParsePing(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{name: "empty acknowledgement"},
		{name: "textual acknowledgement", data: []byte("pong\x00")},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := parsePing(test.data); err != nil {
				t.Fatalf("parsePing() error = %v, want nil", err)
			}
		})
	}
}

func TestParsePingMalformedPayload(t *testing.T) {
	err := parsePing([]byte("unterminated"))
	if !errors.Is(err, ErrPingRead) {
		t.Fatalf("parsePing() error = %v, want ErrPingRead", err)
	}
	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("parsePing() error = %v, want io.ErrUnexpectedEOF", err)
	}
}

func TestGetPingAcceptsEmptyAcknowledgement(t *testing.T) {
	server, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		t.Fatalf("listen UDP server: %v", err)
	}
	defer server.Close()

	client, err := NewWithAddr(server.LocalAddr().(*net.UDPAddr))
	if err != nil {
		t.Fatalf("create client: %v", err)
	}
	defer client.Close()

	serverErr := make(chan error, 1)
	go func() {
		buffer := make([]byte, 64)
		n, address, err := server.ReadFromUDP(buffer)
		if err != nil {
			serverErr <- err
			return
		}
		if n < requestHeaderSize || QueryType(buffer[4]) != PingRequest {
			serverErr <- ErrWrongRequest
			return
		}

		_, err = server.WriteToUDP(singlePacketFixture(ResponsePing, nil), address)
		serverErr <- err
	}()

	if _, err := client.GetPing(context.Background()); err != nil {
		t.Fatalf("GetPing() error = %v, want nil", err)
	}

	if err := <-serverErr; err != nil {
		t.Fatalf("UDP server error: %v", err)
	}
}
