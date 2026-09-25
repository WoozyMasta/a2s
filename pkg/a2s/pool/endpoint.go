// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package pool

import (
	"context"
	"fmt"
	"net"
	"net/netip"
	"strconv"

	"github.com/woozymasta/a2s/pkg/a2s/transport"
)

// Endpoint is a pool target resolved once for repeated queries.
//
// Endpoint does not own a socket.
// Queries use the pool's reusable worker sockets,
// and no DNS lookup is performed after the endpoint is created.
type Endpoint struct {
	pool   *Pool
	target netip.AddrPort
}

// Bind resolves address once and returns an endpoint for repeated queries.
func (p *Pool) Bind(ctx context.Context, address string) (*Endpoint, error) {
	if p == nil {
		return nil, ErrClosed
	}
	if ctx == nil {
		return nil, ErrNilContext
	}
	if err := p.ensureOpen(); err != nil {
		return nil, err
	}
	if target, err := netip.ParseAddrPort(address); err == nil {
		return p.BindAddr(target)
	}

	host, port, err := splitAddress(address)
	if err != nil {
		return nil, err
	}

	addresses, err := p.resolver(ctx, host)
	if err != nil {
		return nil, fmt.Errorf("%w: resolve %q: %w", ErrInvalidTarget, host, err)
	}

	for _, resolved := range addresses {
		target := netip.AddrPortFrom(resolved, port)
		if err := validateTarget(target); err == nil {
			if err := p.ensureOpen(); err != nil {
				return nil, err
			}

			return &Endpoint{pool: p, target: normalizeTarget(target)}, nil
		}
	}

	return nil, fmt.Errorf("%w: no usable address for %q", ErrInvalidTarget, host)
}

// BindAddr creates an endpoint from an already resolved target.
func (p *Pool) BindAddr(target netip.AddrPort) (*Endpoint, error) {
	if p == nil {
		return nil, ErrClosed
	}
	if err := p.ensureOpen(); err != nil {
		return nil, err
	}
	if err := validateTarget(target); err != nil {
		return nil, err
	}

	return &Endpoint{pool: p, target: normalizeTarget(target)}, nil
}

// Query executes one query against the bound target without resolving it again.
func (e *Endpoint) Query(
	ctx context.Context,
	query transport.QueryType,
) (transport.Packet, transport.QueryMeta, error) {
	if e == nil || e.pool == nil {
		return transport.Packet{}, transport.QueryMeta{}, ErrClosed
	}

	return e.pool.Query(ctx, e.target, query)
}

// splitAddress separates a host:port string and validates its port.
func splitAddress(address string) (string, uint16, error) {
	host, portText, err := net.SplitHostPort(address)
	if err != nil || host == "" {
		return "", 0, fmt.Errorf("%w: %q", ErrInvalidTarget, address)
	}

	port, err := strconv.ParseUint(portText, 10, 16)
	if err != nil || port == 0 {
		return "", 0, fmt.Errorf("%w: invalid port in %q", ErrInvalidTarget, address)
	}

	return host, uint16(port), nil
}

// ensureOpen rejects endpoint creation after pool shutdown begins.
func (p *Pool) ensureOpen() error {
	p.stateMu.RLock()
	defer p.stateMu.RUnlock()

	if p.closed || p.root == nil || p.jobs == nil || p.resolver == nil {
		return ErrClosed
	}

	return nil
}

// lookupHost resolves one hostname using the caller's context.
func lookupHost(ctx context.Context, host string) ([]netip.Addr, error) {
	return net.DefaultResolver.LookupNetIP(ctx, "ip", host)
}
