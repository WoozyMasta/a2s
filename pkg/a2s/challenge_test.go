// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package a2s

import (
	"bytes"
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/woozymasta/a2s/pkg/a2s/transport"
)

func TestGetChallengeReturnsChallengeResponse(t *testing.T) {
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
		if n != 5 || transport.QueryType(buffer[4]) != transport.ChallengeRequest {
			serverErr <- errors.New("unexpected challenge request")
			return
		}

		payload := []byte{0x78, 0x56, 0x34, 0x12}
		_, err = server.WriteToUDP(singlePacketFixture(transport.ResponseChallenge, payload), address)
		serverErr <- err
	}()

	client, err := NewWithAddr(server.LocalAddr().(*net.UDPAddr), WithTimeout(time.Second))
	if err != nil {
		t.Fatalf("create client: %v", err)
	}
	defer client.Close()

	got, err := client.GetChallenge(context.Background())
	if err != nil {
		t.Fatalf("GetChallenge returned error: %v", err)
	}
	want := transport.Challenge{0x78, 0x56, 0x34, 0x12}
	if got != want {
		t.Fatalf("challenge = %X, want %X", got, want)
	}

	if err := <-serverErr; err != nil {
		t.Fatalf("UDP server error: %v", err)
	}
}

func TestGetPreservesChallengeBytes(t *testing.T) {
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
			serverErr <- errors.New("unexpected initial rules request")
			return
		}

		challenge := transport.Challenge{0x78, 0x56, 0x34, 0x12}
		payload := append([]byte(nil), challenge[:]...)
		if _, err := server.WriteToUDP(singlePacketFixture(transport.ResponseChallenge, payload), address); err != nil {
			serverErr <- err
			return
		}

		n, address, err = server.ReadFromUDP(buffer)
		if err != nil {
			serverErr <- err
			return
		}
		if n != 9 || transport.QueryType(buffer[4]) != transport.RulesRequest {
			serverErr <- errors.New("unexpected challenged rules request")
			return
		}
		if !bytes.Equal(buffer[5:9], challenge[:]) {
			serverErr <- errors.New("challenge bytes were not preserved")
			return
		}

		_, err = server.WriteToUDP(singlePacketFixture(transport.ResponseRules, []byte("rules")), address)
		serverErr <- err
	}()

	client, err := NewWithAddr(server.LocalAddr().(*net.UDPAddr), WithTimeout(time.Second))
	if err != nil {
		t.Fatalf("create client: %v", err)
	}
	defer client.Close()

	data, flag, _, err := client.Get(context.Background(), transport.RulesRequest)
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if flag != transport.ResponseRules || string(data) != "rules" {
		t.Fatalf("response = (%q, 0x%X), want (rules, 0x%X)", data, flag, transport.ResponseRules)
	}

	if err := <-serverErr; err != nil {
		t.Fatalf("UDP server error: %v", err)
	}
}

func TestGetStopsAfterBoundedChallengeResponses(t *testing.T) {
	server, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		t.Fatalf("listen UDP server: %v", err)
	}
	defer server.Close()

	serverErr := make(chan error, 1)
	go func() {
		buffer := make([]byte, 64*1024)
		payload := []byte{1, 0, 0, 0}
		for i := 0; i < 4; i++ {
			n, address, err := server.ReadFromUDP(buffer)
			if err != nil {
				serverErr <- err
				return
			}
			if n != 9 || transport.QueryType(buffer[4]) != transport.RulesRequest {
				serverErr <- errors.New("unexpected rules request")
				return
			}
			if _, err := server.WriteToUDP(singlePacketFixture(transport.ResponseChallenge, payload), address); err != nil {
				serverErr <- err
				return
			}
		}
		serverErr <- nil
	}()

	client, err := NewWithAddr(server.LocalAddr().(*net.UDPAddr), WithTimeout(time.Second))
	if err != nil {
		t.Fatalf("create client: %v", err)
	}
	defer client.Close()

	_, flag, _, err := client.Get(context.Background(), transport.RulesRequest)
	if !errors.Is(err, transport.ErrChallengeLoop) {
		t.Fatalf("Get error = %v, want transport.ErrChallengeLoop", err)
	}
	if flag != transport.ResponseChallenge {
		t.Fatalf("response flag = 0x%X, want 0x%X", flag, transport.ResponseChallenge)
	}

	if err := <-serverErr; err != nil {
		t.Fatalf("UDP server error: %v", err)
	}
}
