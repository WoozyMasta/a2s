// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	proxycache "github.com/woozymasta/a2s/pkg/a2s/proxy"
	"github.com/woozymasta/a2s/pkg/a2s/server"
)

// executeProxy owns process signal handling for the proxy command.
func executeProxy(app *Application, command *ProxyCommand, clientOptions ClientOptions) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	return executeProxyContext(ctx, app, command, clientOptions)
}

// executeProxyContext prepares and runs proxy
// until cancellation or a fatal local server error.
func executeProxyContext(
	ctx context.Context,
	app *Application,
	command *ProxyCommand,
	clientOptions ClientOptions,
) error {
	if ctx == nil {
		return errors.New("proxy runtime context must not be nil")
	}

	preparation, err := prepareProxyStartup(ctx, command, clientOptions)
	if err != nil {
		return app.wrapError("error.proxy_startup", "proxy startup failed", err)
	}
	defer func() { _ = preparation.close() }()

	provider, err := server.NewStatelessChallengeProvider()
	if err != nil {
		return app.wrapError("error.proxy_challenge_provider", "create downstream challenge provider", err)
	}

	handler, err := proxycache.NewHandler(
		preparation.cache,
		preparation.relayClient,
		proxycache.HandlerConfig{
			ChallengeProvider: provider,
			// Keep the default ping fast and local; --upstream-ping opts into relay.
			LocalPing: !command.UpstreamPing,
		},
	)
	if err != nil {
		return app.wrapError("error.proxy_handler", "create proxy handler", err)
	}

	poller, err := proxycache.NewPoller(
		preparation.cache,
		preparation.pollClient,
		proxycache.PollerConfig{
			TTL:         time.Duration(command.CacheOptions.TTL),
			InactiveTTL: time.Duration(command.CacheOptions.InactiveTTL),
			Jitter:      time.Duration(command.CacheOptions.Jitter),
			Retries:     command.CacheOptions.Retries,
			OnStateChange: func(change proxycache.StateChange) {
				writeProxyStateChange(app, change)
			},
		},
	)
	if err != nil {
		return app.wrapError("error.proxy_poller", "create proxy poller", err)
	}

	serverOptions := []server.Option{
		server.WithChallengeProvider(provider),
		server.WithChallengePolicy(preparation.policy),
		server.WithPacketizer(preparation.packetizer),
	}
	if command.RateLimitOptions.RateLimit > 0 || command.RateLimitOptions.RateClientLimit > 0 {
		limiter, limiterErr := proxycache.NewRateLimiter(proxycache.RateLimitConfig{
			Global: proxycache.Rate{
				Requests: command.RateLimitOptions.RateLimit,
				Window:   time.Duration(command.RateLimitOptions.RateWindow),
			},
			Client: proxycache.Rate{
				Requests: command.RateLimitOptions.RateClientLimit,
				Window:   time.Duration(command.RateLimitOptions.RateWindow),
			},
		})
		if limiterErr != nil {
			return app.wrapError("error.proxy_rate_limit", "create proxy rate limiter", limiterErr)
		}
		serverOptions = append(serverOptions, server.WithMiddleware(limiter.Middleware()))
	}

	proxyServer, err := server.New(handler, serverOptions...)
	if err != nil {
		return app.wrapError("error.proxy_server", "create proxy server", err)
	}

	conn, err := net.ListenPacket("udp", command.Listen)
	if err != nil {
		return fmt.Errorf(
			"%s: %w",
			app.localize("error.proxy_listen", "listen on %s", command.Listen),
			err,
		)
	}
	defer func() { _ = conn.Close() }()

	writeProxyStartup(app, conn.LocalAddr().String(), preparation.upstream)

	runtimeCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	serveErr := make(chan error, 1)
	go func() { serveErr <- proxyServer.ServeContext(runtimeCtx, conn) }()

	pollDone := make(chan struct{})
	pollFailure := make(chan error, 1)
	go func() {
		defer close(pollDone)
		if err := poller.Run(runtimeCtx); err != nil {
			pollFailure <- err
		}
	}()

	select {
	case <-ctx.Done():
		cancel()
		<-serveErr
		<-pollDone
		return nil

	case err := <-serveErr:
		cancel()
		<-pollDone
		if isProxyContextStop(err) {
			return nil
		}
		return app.wrapError("error.proxy_server_stopped", "proxy server stopped", err)

	case err := <-pollFailure:
		cancel()
		<-serveErr
		<-pollDone
		return app.wrapError("error.proxy_poller_stopped", "proxy poller stopped", err)
	}
}

// isProxyContextStop identifies normal context-driven server termination.
func isProxyContextStop(err error) bool {
	return err == nil ||
		errors.Is(err, context.Canceled) ||
		errors.Is(err, server.ErrServerClosed)
}

// writeProxyStartup reports the endpoint mapping after the listener is ready.
func writeProxyStartup(app *Application, listen, upstream string) {
	if app == nil || app.Out == nil {
		return
	}
	_, _ = fmt.Fprintln(app.Out, app.localize(
		"proxy.listening",
		"proxy listening on %s -> %s",
		listen,
		upstream,
	))
}

// writeProxyStateChange reports meaningful cache availability transitions.
func writeProxyStateChange(app *Application, change proxycache.StateChange) {
	if app == nil || app.Err == nil {
		return
	}

	if change.Active {
		_, _ = fmt.Fprintln(app.Err, app.localize(
			"proxy.recovered",
			"%s recovered",
			proxyQueryName(change.Query),
		))
		return
	}

	_, _ = fmt.Fprintln(app.Err, app.localize(
		"proxy.unavailable",
		"%s unavailable: %v",
		proxyQueryName(change.Query),
		change.Err,
	))
}
