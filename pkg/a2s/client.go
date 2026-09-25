// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package a2s

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/woozymasta/a2s/pkg/a2s/transport"
)

// Client handles UDP connection and A2S protocol queries.
// Queries on one Client are serialized for the lifetime of each transaction.
type Client struct {
	conn       *net.UDPConn  // UDP connection to the server.
	address    *net.UDPAddr  // Server network address.
	querySem   chan struct{} // Context-aware query serialization.
	readBuf    []byte        // Reusable UDP read buffer.
	timeout    time.Duration // UDP read deadline.
	timeoutMu  sync.RWMutex  // Protects timeout changes and reads.
	queryMu    sync.Mutex    // Serializes queries and lifecycle changes.
	bufferSize uint16        // Maximum UDP datagram size to read.
}

// Option configures a Client before its UDP connection is opened.
type Option func(*Client) error

// New creates a client for a host and port and opens its UDP connection.
func New(host string, port int, opts ...Option) (*Client, error) {
	if host == "" || port < 1 || port > 65535 {
		return nil, ErrInvalidAddress
	}

	return NewWithString(net.JoinHostPort(host, strconv.Itoa(port)), opts...)
}

// NewWithString creates a client from a host:port address and opens its UDP connection.
func NewWithString(address string, opts ...Option) (*Client, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return nil, fmt.Errorf("%w %q: %v", ErrInvalidAddress, address, err)
	}

	return NewWithAddr(udpAddr, opts...)
}

// NewWithAddr creates a client for a resolved address and opens its UDP connection.
func NewWithAddr(addr *net.UDPAddr, opts ...Option) (*Client, error) {
	if err := validateAddress(addr); err != nil {
		return nil, err
	}

	client := &Client{
		address:    cloneAddress(addr),
		timeout:    DefaultDeadlineTimeout,
		bufferSize: transport.DefaultBufferSize,
		readBuf:    make([]byte, transport.DefaultBufferSize),
		querySem:   make(chan struct{}, 1),
	}

	for _, opt := range opts {
		if opt == nil {
			continue
		}
		if err := opt(client); err != nil {
			return nil, err
		}
	}

	conn, err := net.DialUDP("udp", nil, client.address)
	if err != nil {
		return nil, fmt.Errorf("dial %s: %w", client.address, err)
	}

	client.conn = conn
	return client, nil
}

// WithTimeout sets the UDP read deadline used by the client.
func WithTimeout(timeout time.Duration) Option {
	return func(client *Client) error {
		if timeout <= 0 {
			return ErrInvalidTimeout
		}

		client.timeout = timeout
		return nil
	}
}

// WithBufferSize sets the maximum UDP datagram size read by the client.
func WithBufferSize(size uint16) Option {
	return func(client *Client) error {
		return client.setBufferSize(size)
	}
}

// Addr returns a copy of the server network address.
func (c *Client) Addr() *net.UDPAddr {
	if c == nil {
		return nil
	}

	return cloneAddress(c.address)
}

// BufferSize returns the maximum UDP datagram size read by the client.
func (c *Client) BufferSize() uint16 {
	if c == nil {
		return 0
	}
	c.queryMu.Lock()
	defer c.queryMu.Unlock()

	return c.bufferSize
}

// Timeout returns the UDP read deadline.
func (c *Client) Timeout() time.Duration {
	if c == nil {
		return 0
	}
	c.timeoutMu.RLock()
	defer c.timeoutMu.RUnlock()

	return c.timeout
}

// SetBufferSize sets the maximum UDP datagram size read by the client.
func (c *Client) SetBufferSize(size uint16) error {
	if c == nil {
		return ErrClientClosed
	}
	c.queryMu.Lock()
	defer c.queryMu.Unlock()

	return c.setBufferSize(size)
}

// setBufferSize validates and applies a receive-buffer size while the client
// lifecycle lock is held by the caller.
func (c *Client) setBufferSize(size uint16) error {
	if size == 0 {
		return ErrInvalidBufferSize
	}

	c.bufferSize = size
	if cap(c.readBuf) < int(size) {
		c.readBuf = make([]byte, size)
	} else {
		c.readBuf = c.readBuf[:size]
	}

	return nil
}

// SetTimeout sets the UDP read deadline.
func (c *Client) SetTimeout(timeout time.Duration) error {
	if c == nil {
		return ErrClientClosed
	}
	c.queryMu.Lock()
	defer c.queryMu.Unlock()

	if timeout <= 0 {
		return ErrInvalidTimeout
	}

	c.timeoutMu.Lock()
	defer c.timeoutMu.Unlock()

	c.timeout = timeout
	return nil
}

// Close closes the UDP connection. It is safe to call multiple times.
func (c *Client) Close() error {
	if c == nil {
		return nil
	}
	c.queryMu.Lock()
	defer c.queryMu.Unlock()

	if c.conn == nil {
		return nil
	}

	err := c.conn.Close()
	c.conn = nil
	return err
}

// validateAddress rejects unusable wildcard and zero-port destinations.
func validateAddress(addr *net.UDPAddr) error {
	if addr == nil || addr.IP == nil || addr.IP.IsUnspecified() || addr.Port == 0 {
		return ErrInvalidAddress
	}

	return nil
}

// cloneAddress returns an address copy that cannot be mutated through the caller's IP slice.
func cloneAddress(addr *net.UDPAddr) *net.UDPAddr {
	if addr == nil {
		return nil
	}

	clone := *addr
	clone.IP = append(net.IP(nil), addr.IP...)
	return &clone
}

// Query sends one A2S request and returns the final logical response packet.
//
// Challenge-response, split-packet assembly,
// and compressed split responses are handled internally.
// QueryMeta.Duration covers the complete logical transaction,
// including retries and challenge exchange.
func (c *Client) Query(ctx context.Context, requestType transport.QueryType) (transport.Packet, transport.QueryMeta, error) {
	if c == nil {
		return transport.Packet{}, transport.QueryMeta{}, ErrClientClosed
	}
	if ctx == nil {
		return transport.Packet{}, transport.QueryMeta{}, ErrNilContext
	}

	effectiveCtx, cancel := c.effectiveContext(ctx)
	defer cancel()
	if err := c.acquireQuery(effectiveCtx); err != nil {
		return transport.Packet{}, transport.QueryMeta{}, err
	}
	defer c.releaseQuery()

	c.queryMu.Lock()
	defer c.queryMu.Unlock()

	if c.conn == nil {
		return transport.Packet{}, transport.QueryMeta{}, ErrClientClosed
	}

	readDeadlineDone := make(chan struct{})
	stopReadDeadline := context.AfterFunc(effectiveCtx, func() {
		defer close(readDeadlineDone)
		_ = c.conn.SetReadDeadline(time.Now())
	})
	defer func() {
		if !stopReadDeadline() {
			<-readDeadlineDone
		}
	}()

	packet, meta, err := transport.Execute(
		effectiveCtx,
		clientDatagramIO{conn: c.conn},
		requestType,
		c.readBuf,
	)
	return packet, meta, err
}

// Get sends one A2S request and returns its payload, response type, and query duration.
//
// Deprecated: use Query to retain the complete logical response as a Packet.
func (c *Client) Get(ctx context.Context, requestType transport.QueryType) ([]byte, transport.ResponseType, time.Duration, error) {
	packet, meta, err := c.Query(ctx, requestType)
	return packet.Payload, packet.Type, meta.Duration, err
}

// effectiveContext applies the client timeout
// only when the caller did not provide a deadline of its own.
func (c *Client) effectiveContext(ctx context.Context) (context.Context, context.CancelFunc) {
	c.timeoutMu.RLock()
	timeout := c.timeout
	c.timeoutMu.RUnlock()

	if _, ok := ctx.Deadline(); ok {
		return ctx, func() {}
	}

	return context.WithTimeout(ctx, timeout)
}

// acquireQuery reserves the client for one complete query transaction.
func (c *Client) acquireQuery(ctx context.Context) error {
	if c.querySem == nil {
		return ErrClientClosed
	}

	select {
	case c.querySem <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// releaseQuery releases the client query reservation.
func (c *Client) releaseQuery() {
	<-c.querySem
}

// clientDatagramIO adapts the connected client socket to the shared engine.
type clientDatagramIO struct {
	conn *net.UDPConn // Connected UDP socket used by the client.
}

// Write sends one datagram through the connected client socket.
func (io clientDatagramIO) Write(ctx context.Context, data []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	_, err := io.conn.Write(data)
	return err
}

// Read receives one datagram and applies the context deadline to the socket.
func (io clientDatagramIO) Read(ctx context.Context, data []byte) (int, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	if deadline, ok := ctx.Deadline(); ok {
		if err := io.conn.SetReadDeadline(deadline); err != nil {
			return 0, err
		}
	}

	return io.conn.Read(data)
}
