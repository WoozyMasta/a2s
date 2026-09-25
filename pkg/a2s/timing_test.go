// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package a2s

import (
	"context"
	"encoding/binary"
	"net"
	"testing"
	"time"

	"github.com/woozymasta/a2s/pkg/a2s/transport"
)

func TestGetDurationIncludesChallengeExchange(t *testing.T) {
	server, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		t.Fatalf("listen UDP server: %v", err)
	}
	defer server.Close()

	serverErr := make(chan error, 1)
	go func() {
		buffer := make([]byte, 64*1024)
		n, address, err := server.ReadFromUDP(buffer)
		if err != nil {
			serverErr <- err
			return
		}
		if n != 9 || transport.QueryType(buffer[4]) != transport.RulesRequest {
			serverErr <- transport.ErrWrongRequest
			return
		}

		time.Sleep(30 * time.Millisecond)
		challenge := make([]byte, 4)
		binary.LittleEndian.PutUint32(challenge, 0x12345678)
		if _, err := server.WriteToUDP(singlePacketFixture(transport.ResponseChallenge, challenge), address); err != nil {
			serverErr <- err
			return
		}

		n, address, err = server.ReadFromUDP(buffer)
		if err != nil {
			serverErr <- err
			return
		}
		if n != 9 || transport.QueryType(buffer[4]) != transport.RulesRequest {
			serverErr <- transport.ErrWrongRequest
			return
		}

		time.Sleep(30 * time.Millisecond)
		_, err = server.WriteToUDP(singlePacketFixture(transport.ResponseRules, []byte("rules")), address)
		serverErr <- err
	}()

	client, err := NewWithAddr(server.LocalAddr().(*net.UDPAddr), WithTimeout(time.Second))
	if err != nil {
		t.Fatalf("create client: %v", err)
	}
	defer client.Close()

	_, flag, duration, err := client.Get(context.Background(), transport.RulesRequest)
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if flag != transport.ResponseRules {
		t.Fatalf("response flag = 0x%X, want 0x%X", flag, transport.ResponseRules)
	}
	if duration < 50*time.Millisecond {
		t.Fatalf("query duration = %s, want challenge exchange included", duration)
	}
	if err := <-serverErr; err != nil {
		t.Fatalf("UDP server error: %v", err)
	}
}

func TestGetDurationIncludesSplitAssembly(t *testing.T) {
	server, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		t.Fatalf("listen UDP server: %v", err)
	}
	defer server.Close()

	assembled := singlePacketFixture(transport.ResponseRules, []byte("split timing"))
	packets := sourceSplitPacketSequence(0x12345678, assembled, len(assembled)-1)
	if len(packets) != 2 {
		t.Fatalf("split fixture packet count = %d, want 2", len(packets))
	}

	serverErr := make(chan error, 1)
	go func() {
		buffer := make([]byte, 64*1024)
		_, address, err := server.ReadFromUDP(buffer)
		if err != nil {
			serverErr <- err
			return
		}

		if _, err := server.WriteToUDP(packets[0], address); err != nil {
			serverErr <- err
			return
		}
		time.Sleep(30 * time.Millisecond)
		_, err = server.WriteToUDP(packets[1], address)
		serverErr <- err
	}()

	client, err := NewWithAddr(server.LocalAddr().(*net.UDPAddr), WithTimeout(time.Second))
	if err != nil {
		t.Fatalf("create client: %v", err)
	}
	defer client.Close()

	data, flag, duration, err := client.Get(context.Background(), transport.RulesRequest)
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if flag != transport.ResponseRules || string(data) != "split timing" {
		t.Fatalf("response = (%q, 0x%X), want (split timing, 0x%X)", data, flag, transport.ResponseRules)
	}
	if duration < 25*time.Millisecond {
		t.Fatalf("query duration = %s, want split assembly included", duration)
	}
	if err := <-serverErr; err != nil {
		t.Fatalf("UDP server error: %v", err)
	}
}
