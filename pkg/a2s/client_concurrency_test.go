package a2s

import (
	"context"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestClientSerializesConcurrentQueries(t *testing.T) {
	server, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		t.Fatalf("listen UDP server: %v", err)
	}
	defer server.Close()

	client, err := NewWithAddr(server.LocalAddr().(*net.UDPAddr), WithTimeout(time.Second))
	if err != nil {
		t.Fatalf("create client: %v", err)
	}
	defer client.Close()

	var active int32
	var maxActive int32
	serverErr := make(chan error, 1)
	serverDone := make(chan struct{})
	var handlers sync.WaitGroup

	go func() {
		defer close(serverDone)

		buffer := make([]byte, 64*1024)
		for i := 0; i < 2; i++ {
			n, address, err := server.ReadFromUDP(buffer)
			if err != nil {
				serverErr <- err
				return
			}

			if n < 5 || Flag(buffer[4]) != PingRequest {
				serverErr <- ErrWrongRequest
				return
			}

			current := atomic.AddInt32(&active, 1)
			for {
				observed := atomic.LoadInt32(&maxActive)
				if current <= observed || atomic.CompareAndSwapInt32(&maxActive, observed, current) {
					break
				}
			}

			handlers.Add(1)
			go func(address *net.UDPAddr) {
				defer handlers.Done()
				defer atomic.AddInt32(&active, -1)

				time.Sleep(25 * time.Millisecond)
				if _, err := server.WriteToUDP(singlePacketFixture(pingResponse, nil), address); err != nil {
					serverErr <- err
				}
			}(address)
		}
		handlers.Wait()
	}()

	var queries sync.WaitGroup
	queryErr := make(chan error, 2)
	for i := 0; i < 2; i++ {
		queries.Add(1)
		go func() {
			defer queries.Done()
			if _, _, _, err := client.Get(context.Background(), PingRequest); err != nil {
				queryErr <- err
			}
		}()
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

	if got, want := atomic.LoadInt32(&maxActive), int32(1); got != want {
		t.Fatalf("maximum concurrent server handlers = %d, want %d", got, want)
	}
}
