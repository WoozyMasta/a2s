package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	proxycache "github.com/woozymasta/a2s/pkg/a2s/proxy"
	"github.com/woozymasta/a2s/pkg/a2s/server"
)

// executeProxy owns process signal handling for the proxy command.
func executeProxy(app *Application, command *ProxyCommand) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	return executeProxyContext(ctx, app, command)
}

// executeProxyContext prepares and runs proxy
// until cancellation or a fatal local server error.
func executeProxyContext(ctx context.Context, app *Application, command *ProxyCommand) error {
	if ctx == nil {
		return fmt.Errorf("proxy runtime context must not be nil")
	}

	preparation, err := prepareProxyStartup(ctx, command)
	if err != nil {
		return err
	}
	defer func() { _ = preparation.close() }()

	provider, err := server.NewStatelessChallengeProvider()
	if err != nil {
		return fmt.Errorf("create downstream challenge provider: %w", err)
	}

	handler, err := proxycache.NewHandler(
		preparation.cache,
		preparation.relayClient,
		proxycache.HandlerConfig{
			ChallengeProvider: provider,
			LocalPing:         !command.UpstreamPing,
		},
	)
	if err != nil {
		return fmt.Errorf("create proxy handler: %w", err)
	}

	poller, err := proxycache.NewPoller(
		preparation.cache,
		preparation.pollClient,
		proxycache.PollerConfig{
			TTL:         command.TTL,
			InactiveTTL: command.InactiveTTL,
			Jitter:      command.Jitter,
			Retries:     command.Retries,
			OnStateChange: func(change proxycache.StateChange) {
				writeProxyStateChange(app, change)
			},
		},
	)
	if err != nil {
		return fmt.Errorf("create proxy poller: %w", err)
	}

	proxyServer, err := server.New(
		handler,
		server.WithChallengeProvider(provider),
		server.WithChallengePolicy(preparation.policy),
		server.WithPacketizer(preparation.packetizer),
	)
	if err != nil {
		return fmt.Errorf("create proxy server: %w", err)
	}

	conn, err := net.ListenPacket("udp", command.Listen)
	if err != nil {
		return fmt.Errorf("listen on %q: %w", command.Listen, err)
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
		return fmt.Errorf("proxy server stopped: %w", err)

	case err := <-pollFailure:
		cancel()
		<-serveErr
		<-pollDone
		return fmt.Errorf("proxy poller stopped: %w", err)
	}
}

// isProxyContextStop identifies normal context-driven server termination.
func isProxyContextStop(err error) bool {
	return err == nil || errors.Is(err, context.Canceled) || errors.Is(err, server.ErrServerClosed)
}

// writeProxyStartup reports the endpoint mapping after the listener is ready.
func writeProxyStartup(app *Application, listen, upstream string) {
	if app == nil || app.Out == nil {
		return
	}
	_, _ = fmt.Fprintf(app.Out, "proxy listening on %s -> %s\n", listen, upstream)
}

// writeProxyStateChange reports meaningful cache availability transitions.
func writeProxyStateChange(app *Application, change proxycache.StateChange) {
	if app == nil || app.Err == nil {
		return
	}

	if change.Active {
		_, _ = fmt.Fprintf(app.Err, "%s recovered\n", proxyQueryName(change.Query))
		return
	}

	_, _ = fmt.Fprintf(
		app.Err,
		"%s unavailable: %v\n",
		proxyQueryName(change.Query),
		change.Err,
	)
}
