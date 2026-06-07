package a2s

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"
)

func TestGetContextDeadlineCoversChallengeExchange(t *testing.T) {
	server, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		t.Fatalf("listen UDP server: %v", err)
	}
	defer server.Close()

	serverReady := make(chan struct{})
	serverErr := make(chan error, 1)
	go func() {
		buffer := make([]byte, 64*1024)
		n, address, err := server.ReadFromUDP(buffer)
		if err != nil {
			serverErr <- err
			return
		}
		if n != 9 || QueryType(buffer[4]) != RulesRequest {
			serverErr <- ErrWrongRequest
			return
		}

		challenge := []byte{1, 2, 3, 4}
		if _, err := server.WriteToUDP(singlePacketFixture(ResponseChallenge, challenge), address); err != nil {
			serverErr <- err
			return
		}

		if _, _, err := server.ReadFromUDP(buffer); err != nil {
			serverErr <- err
			return
		}
		close(serverReady)
	}()

	client, err := NewWithAddr(server.LocalAddr().(*net.UDPAddr), WithTimeout(time.Second))
	if err != nil {
		t.Fatalf("create client: %v", err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	started := time.Now()
	_, _, _, err = client.Get(ctx, RulesRequest)
	elapsed := time.Since(started)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Get error = %v, want context deadline", err)
	}
	if elapsed >= 500*time.Millisecond {
		t.Fatalf("Get ignored total context deadline: elapsed %s", elapsed)
	}

	select {
	case err := <-serverErr:
		t.Fatalf("UDP server error: %v", err)
	case <-serverReady:
	case <-time.After(time.Second):
		t.Fatal("server did not observe challenged request")
	}
}

func TestGetContextDeadlineCoversQueryQueue(t *testing.T) {
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

	firstRequest := make(chan struct{})
	releaseFirst := make(chan struct{})
	serverErr := make(chan error, 1)
	go func() {
		buffer := make([]byte, 64*1024)
		n, address, err := server.ReadFromUDP(buffer)
		if err != nil {
			serverErr <- err
			return
		}

		if n < 5 || QueryType(buffer[4]) != PingRequest {
			serverErr <- ErrWrongRequest
			return
		}

		close(firstRequest)
		<-releaseFirst
		_, err = server.WriteToUDP(singlePacketFixture(ResponsePing, nil), address)
		serverErr <- err
	}()

	firstDone := make(chan error, 1)
	go func() {
		_, _, _, err := client.Get(context.Background(), PingRequest)
		firstDone <- err
	}()

	select {
	case err := <-serverErr:
		t.Fatalf("UDP server error: %v", err)
	case <-firstRequest:
	case <-time.After(time.Second):
		t.Fatal("server did not observe first request")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	started := time.Now()
	_, _, _, secondErr := client.Get(ctx, PingRequest)
	elapsed := time.Since(started)
	if !errors.Is(secondErr, context.DeadlineExceeded) {
		t.Fatalf("queued Get error = %v, want context deadline", secondErr)
	}
	if elapsed >= 500*time.Millisecond {
		t.Fatalf("queued Get ignored context deadline: elapsed %s", elapsed)
	}

	close(releaseFirst)
	if err := <-firstDone; err != nil {
		t.Fatalf("first Get returned error: %v", err)
	}
	if err := <-serverErr; err != nil {
		t.Fatalf("UDP server error: %v", err)
	}
}
