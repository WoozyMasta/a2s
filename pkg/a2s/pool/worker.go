// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package pool

import (
	"context"
	"errors"
	"io"
	"net"
	"net/netip"
	"sync"
	"time"

	"github.com/woozymasta/a2s/pkg/a2s/transport"
)

// socketFamily identifies the address family of a worker-owned UDP socket.
type socketFamily uint8

const (
	// socketIPv4 selects an IPv4 UDP socket.
	socketIPv4 socketFamily = iota
	// socketIPv6 selects an IPv6 UDP socket.
	socketIPv6
)

// worker owns sockets and executes jobs serially.
type worker struct {
	sockets [2]*net.UDPConn // One lazy socket per address family.
	pool    *Pool           // Parent pool and shared configuration.
	readBuf []byte          // Reusable receive buffer for serialized jobs.
	mu      sync.Mutex      // Protects socket lifecycle during cleanup.
}

// datagramIO adapts an unconnected worker socket to one target.
type datagramIO struct {
	conn    *net.UDPConn
	target  netip.AddrPort
	started bool
}

// worker creates one socket owner and runs until the pool closes.
func (p *Pool) worker() {
	state := &worker{
		pool:    p,
		readBuf: make([]byte, p.bufferSize),
	}
	defer func() {
		state.closeSockets()
		p.workers.Done()
	}()

	for {
		select {
		case <-p.root.Done():
			return

		default:
		}

		select {
		case <-p.root.Done():
			return

		case job := <-p.jobs:
			p.releaseQueueSpace()
			state.execute(job)
		}
	}
}

// execute runs one complete logical query and never blocks on result delivery.
func (w *worker) execute(job workItem) {
	defer func() {
		if job.cleanup != nil {
			job.cleanup()
		}
	}()

	if err := job.ctx.Err(); err != nil {
		w.publish(job, queryResult{err: err})
		return
	}

	family := targetFamily(job.target)
	conn, err := w.socket(family)
	if err != nil {
		w.publish(job, queryResult{err: err})
		return
	}

	io := &datagramIO{
		conn:   conn,
		target: job.target,
	}
	stopDeadline := wakeOnCancel(job.ctx, conn)
	packet, meta, queryErr := transport.Execute(job.ctx, io, job.query, w.readBuf)
	stopDeadline()

	if io.started && shouldResetSocket(queryErr) {
		w.reset(family)
	}

	w.publish(job, queryResult{
		packet: packet,
		meta:   meta,
		err:    queryErr,
	})
}

// shouldResetSocket reports whether a failed transaction
// may have left unread or delayed datagrams on the worker socket.
func shouldResetSocket(err error) bool {
	if err == nil {
		return false
	}

	// An unsupported query is a complete, valid response from the server.
	// Keep the socket unless the error also contains cancellation
	// or deadline state from a later retry.
	if errors.Is(err, transport.ErrQueryUnsupported) &&
		!errors.Is(err, context.Canceled) &&
		!errors.Is(err, context.DeadlineExceeded) {
		return false
	}

	return true
}

// publish sends one result to the buffered job channel.
func (w *worker) publish(job workItem, result queryResult) {
	if job.stream != nil {
		job.stream.publish(streamCompletion{
			index:   job.index,
			request: job.request,
			result:  result,
		})
		return
	}

	select {
	case job.result <- result:
	default:
	}
}

// socket returns or lazily creates the socket for one address family.
func (w *worker) socket(family socketFamily) (*net.UDPConn, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if conn := w.sockets[family]; conn != nil {
		return conn, nil
	}

	network := "udp6"
	if family == socketIPv4 {
		network = "udp4"
	}
	conn, err := net.ListenUDP(network, nil)
	if err != nil {
		return nil, err
	}

	w.sockets[family] = conn
	return conn, nil
}

// reset closes only the socket family that had the failed transaction.
func (w *worker) reset(family socketFamily) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if conn := w.sockets[family]; conn != nil {
		_ = conn.Close()
		w.sockets[family] = nil
	}
}

// closeSockets releases all sockets owned by the worker.
func (w *worker) closeSockets() {
	w.mu.Lock()
	defer w.mu.Unlock()

	for index, conn := range w.sockets {
		if conn != nil {
			_ = conn.Close()
			w.sockets[index] = nil
		}
	}
}

// Write sends a datagram to the active target.
func (d *datagramIO) Write(ctx context.Context, data []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	d.started = true
	written, err := d.conn.WriteToUDP(data, udpAddr(d.target))
	if err == nil && written != len(data) {
		return io.ErrShortWrite
	}

	return err
}

// Read receives only datagrams from the active target.
func (d *datagramIO) Read(ctx context.Context, data []byte) (int, error) {
	for {
		if err := ctx.Err(); err != nil {
			return 0, err
		}
		if deadline, ok := ctx.Deadline(); ok {
			if err := d.conn.SetReadDeadline(deadline); err != nil {
				return 0, err
			}
		}

		n, remote, err := d.conn.ReadFromUDP(data)
		if err != nil {
			return 0, err
		}
		if sameTarget(remote, d.target) {
			return n, nil
		}
	}
}

// wakeOnCancel interrupts blocked UDP operations when ctx is canceled.
func wakeOnCancel(ctx context.Context, conn *net.UDPConn) func() {
	done := make(chan struct{})
	stop := context.AfterFunc(ctx, func() {
		_ = conn.SetDeadline(time.Now())
		close(done)
	})

	return func() {
		if !stop() {
			<-done
		}
		_ = conn.SetDeadline(time.Time{})
	}
}

// targetFamily selects the socket family after target normalization.
func targetFamily(target netip.AddrPort) socketFamily {
	if target.Addr().Is4() {
		return socketIPv4
	}

	return socketIPv6
}

// udpAddr converts a resolved target into the standard UDP address form.
func udpAddr(target netip.AddrPort) *net.UDPAddr {
	return &net.UDPAddr{
		IP:   net.IP(target.Addr().AsSlice()),
		Port: int(target.Port()),
		Zone: target.Addr().Zone(),
	}
}

// sameTarget rejects stale or unrelated datagrams from the worker socket.
func sameTarget(remote *net.UDPAddr, target netip.AddrPort) bool {
	if remote == nil || remote.Port != int(target.Port()) {
		return false
	}

	address, ok := netip.AddrFromSlice(remote.IP)
	if !ok {
		return false
	}

	if target.Addr().Zone() != "" && remote.Zone != target.Addr().Zone() {
		return false
	}

	return address.Unmap() == target.Addr().Unmap()
}
