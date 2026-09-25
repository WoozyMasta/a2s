// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package a2s

import (
	"context"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/woozymasta/a2s/pkg/a2s/transport"
)

func TestClientSerializesConcurrentQueries(t *testing.T) {
	server, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		t.Fatalf("listen UDP server: %v", err)
	}
	defer func() { _ = server.Close() }()

	client, err := NewWithAddr(server.LocalAddr().(*net.UDPAddr), WithTimeout(time.Second))
	if err != nil {
		t.Fatalf("create client: %v", err)
	}
	defer func() { _ = client.Close() }()

	var active atomic.Int32
	var maxActive atomic.Int32
	serverErr := make(chan error, 1)
	serverDone := make(chan struct{})
	var handlers sync.WaitGroup

	go func() {
		defer close(serverDone)

		buffer := make([]byte, 64*1024)
		for range 2 {
			n, address, err := server.ReadFromUDP(buffer)
			if err != nil {
				serverErr <- err
				return
			}

			if n < 5 || transport.QueryType(buffer[4]) != transport.PingRequest {
				serverErr <- transport.ErrWrongRequest
				return
			}

			current := active.Add(1)
			for {
				observed := maxActive.Load()
				if current <= observed || maxActive.CompareAndSwap(observed, current) {
					break
				}
			}

			handlers.Add(1)
			go func(address *net.UDPAddr) {
				defer handlers.Done()
				defer active.Add(-1)

				time.Sleep(25 * time.Millisecond)
				if _, err := server.WriteToUDP(singlePacketFixture(transport.ResponsePing, nil), address); err != nil {
					serverErr <- err
				}
			}(address)
		}
		handlers.Wait()
	}()

	var queries sync.WaitGroup
	queryErr := make(chan error, 2)
	for range 2 {
		queries.Go(func() {
			if _, _, _, err := client.Get(context.Background(), transport.PingRequest); err != nil {
				queryErr <- err
			}
		})
	}
	queries.Wait()

	select {
	case <-serverDone:
	case <-time.After(time.Second):
		t.Fatal("UDP server did not finish")
	}

	select {
	case err := <-serverErr:
		t.Fatalf("UDP server error: %v", err)
	default:
	}

	select {
	case err := <-queryErr:
		t.Fatalf("concurrent query error: %v", err)
	default:
	}

	if got, want := maxActive.Load(), int32(1); got != want {
		t.Fatalf("maximum concurrent server handlers = %d, want %d", got, want)
	}
}
