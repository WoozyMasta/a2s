// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

/*
Package pool provides bounded concurrent A2S queries over reusable UDP sockets.

Use the regular [a2s.Client] for simple work against one server.
A Client serializes complete transactions, while [Pool] is intended
for bounded concurrency across many already-known targets.
The pool reuses one UDP socket per address family for each worker;
it never multiplexes two A2S transactions over the same socket.
Challenge exchange, split responses, and protocol validation
remain part of each logical transaction.

[WithTimeout] sets a total query budget.
The budget starts when work is submitted,
so queue wait is included together with socket scheduling,
network I/O, split assembly, and validation.
The caller's context can impose an earlier deadline.
Calling [WithTimeout] with zero disables the pool-specific budget.

[Pool.Bind] resolves a hostname once using the supplied context
and stores one usable resolved address in the returned [Endpoint].
Literal IP addresses and [Pool.BindAddr] do not perform DNS lookups.
Endpoint queries use the pool's worker sockets and never resolve the target again;
callers that need DNS refresh must bind again.

Workers validate the source address of every datagram
before passing it to the transaction engine.
This rejects late packets from other servers,
but A2S has no general transaction identifier.
A late compatible packet from the same server
and port cannot always be distinguished from the current response.
Socket reuse and reset-after-failure reduce this ambiguity
but cannot remove the protocol limitation.

[Pool.Stream] accepts requests through a caller-owned channel
and closes only its result channel.
The caller must close the input channel during normal completion.
Results are delivered in completion order;
[Result.Index] preserves submission order.
Accepted work and internal completions are bounded by the pool,
so an unread output applies backpressure instead of growing memory without limit.
Canceling the context stops input acceptance, cancels accepted work,
and closes output without waiting for the input channel to be closed.

An [Endpoint] implements the minimal upstream contract used by [pkg/a2s/proxy],
so a bound pool endpoint can be passed directly to a proxy poller or handler.

[a2s.Client]: https://pkg.go.dev/github.com/woozymasta/a2s/pkg/a2s#Client
[pkg/a2s/proxy]: https://pkg.go.dev/github.com/woozymasta/a2s/pkg/a2s/proxy
*/
package pool
