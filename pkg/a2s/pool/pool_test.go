// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package pool

import (
	"context"
	"errors"
	"net"
	"net/netip"
	"testing"
	"time"

	"github.com/woozymasta/a2s/pkg/a2s/transport"
)

func TestPoolReusesSocketAndFiltersOtherSources(t *testing.T) {
	response := transport.AppendPacket(nil, transport.Packet{Type: transport.ResponseInfo})
	first := newResponder(t, response, true)
	second := newResponder(t, response, false)

	p, err := New(WithWorkers(1), WithTimeout(time.Second))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = p.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if _, _, err := p.Query(ctx, first.target, transport.InfoRequest); err != nil {
		t.Fatalf("first Query() error = %v", err)
	}
	if _, _, err := p.Query(ctx, second.target, transport.InfoRequest); err != nil {
		t.Fatalf("second Query() error = %v", err)
	}

	firstPort := first.waitSourcePort(t)
	secondPort := second.waitSourcePort(t)
	if firstPort != secondPort {
		t.Fatalf("worker socket port changed: first=%d second=%d", firstPort, secondPort)
	}
	if first.noiseSent() == 0 {
		t.Fatal("responder did not send a packet from the unrelated source")
	}
}

func TestPoolQueueTimeoutDoesNotSendExpiredJob(t *testing.T) {
	first := newSilentResponder(t)
	second := newSilentResponder(t)

	p, err := New(
		WithWorkers(1),
		WithQueueSize(1),
		WithTimeout(time.Second),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = p.Close() }()

	firstDone := make(chan error, 1)
	go func() {
		_, _, err := p.Query(context.Background(), first.target, transport.InfoRequest)
		firstDone <- err
	}()
	first.waitRequest(t)

	secondCtx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	_, _, err = p.Query(secondCtx, second.target, transport.InfoRequest)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("queued Query() error = %v, want deadline exceeded", err)
	}

	select {
	case <-second.requests:
		t.Fatal("expired queued query was sent")
	case <-time.After(100 * time.Millisecond):
	}

	_ = p.Close()
	if err := <-firstDone; !errors.Is(err, ErrClosed) {
		t.Fatalf("in-flight Query() error = %v, want ErrClosed", err)
	}
}

func TestPoolReleasesQueueSlotWhenWorkerDequeuesJob(t *testing.T) {
	response := transport.AppendPacket(nil, transport.Packet{Type: transport.ResponseInfo})
	first := newDelayedResponder(t, response, 100*time.Millisecond)
	second := newSilentResponder(t)
	third := newSilentResponder(t)

	p, err := New(
		WithWorkers(1),
		WithQueueSize(1),
		WithTimeout(time.Second),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = p.Close() }()

	firstDone := make(chan error, 1)
	go func() {
		_, _, err := p.Query(context.Background(), first.target, transport.InfoRequest)
		firstDone <- err
	}()
	first.waitRequest(t)

	secondDone := make(chan error, 1)
	go func() {
		_, _, err := p.Query(context.Background(), second.target, transport.InfoRequest)
		secondDone <- err
	}()
	waitQueueLength(t, p, 1)

	if err := <-firstDone; err != nil {
		t.Fatalf("first Query() error = %v", err)
	}
	second.waitRequest(t)

	thirdDone := make(chan error, 1)
	go func() {
		_, _, err := p.Query(context.Background(), third.target, transport.InfoRequest)
		thirdDone <- err
	}()
	waitQueueLength(t, p, 1)

	_ = p.Close()
	if err := <-secondDone; !errors.Is(err, ErrClosed) {
		t.Fatalf("second Query() error = %v, want ErrClosed", err)
	}
	if err := <-thirdDone; !errors.Is(err, ErrClosed) {
		t.Fatalf("third Query() error = %v, want ErrClosed", err)
	}
}

type responder struct {
	target        netip.AddrPort
	conn          *net.UDPConn
	noise         *net.UDPConn
	response      []byte
	responseDelay time.Duration
	ports         chan int
	requests      chan struct{}
	noiseHits     chan struct{}
}

func newResponder(t *testing.T, response []byte, sendNoise bool) *responder {
	return newResponderWithDelay(t, response, sendNoise, 0)
}

func newDelayedResponder(t *testing.T, response []byte, delay time.Duration) *responder {
	return newResponderWithDelay(t, response, false, delay)
}

func newResponderWithDelay(
	t *testing.T,
	response []byte,
	sendNoise bool,
	responseDelay time.Duration,
) *responder {
	t.Helper()
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.ParseIP("127.0.0.1")})
	if err != nil {
		t.Fatal(err)
	}

	r := &responder{
		target:        netip.MustParseAddrPort(conn.LocalAddr().String()),
		conn:          conn,
		response:      response,
		responseDelay: responseDelay,
		ports:         make(chan int, 4),
		requests:      make(chan struct{}, 4),
		noiseHits:     make(chan struct{}, 4),
	}
	if sendNoise {
		r.noise, err = net.ListenUDP("udp4", &net.UDPAddr{IP: net.ParseIP("127.0.0.1")})
		if err != nil {
			_ = conn.Close()
			t.Fatal(err)
		}
	}

	t.Cleanup(func() {
		_ = r.conn.Close()
		if r.noise != nil {
			_ = r.noise.Close()
		}
	})

	go r.serve(sendNoise)
	return r
}

func newSilentResponder(t *testing.T) *responder {
	r := newResponder(t, nil, false)
	return r
}

func (r *responder) serve(sendNoise bool) {
	buffer := make([]byte, 2048)
	for {
		n, remote, err := r.conn.ReadFromUDP(buffer)
		if err != nil {
			return
		}
		if n == 0 {
			continue
		}

		select {
		case r.requests <- struct{}{}:
		default:
		}

		select {
		case r.ports <- remote.Port:
		default:
		}

		if sendNoise {
			if _, err := r.noise.WriteToUDP([]byte{0}, remote); err == nil {
				select {
				case r.noiseHits <- struct{}{}:
				default:
				}
			}
		}

		if r.response != nil {
			if r.responseDelay > 0 {
				time.Sleep(r.responseDelay)
			}
			_, _ = r.conn.WriteToUDP(r.response, remote)
		}
	}
}

func (r *responder) waitSourcePort(t *testing.T) int {
	t.Helper()
	select {
	case port := <-r.ports:
		return port
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for UDP request")
		return 0
	}
}

func (r *responder) waitRequest(t *testing.T) {
	t.Helper()
	select {
	case <-r.requests:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for UDP request")
	}
}

func (r *responder) noiseSent() int {
	count := 0
	for {
		select {
		case <-r.noiseHits:
			count++
		default:
			return count
		}
	}
}

func waitQueueLength(t *testing.T, p *Pool, want int) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if len(p.jobs) == want {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("queue length = %d, want %d", len(p.jobs), want)
}
