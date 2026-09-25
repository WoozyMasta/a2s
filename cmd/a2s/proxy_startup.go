// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/woozymasta/a2s/pkg/a2s"
	"github.com/woozymasta/a2s/pkg/a2s/proxy"
	"github.com/woozymasta/a2s/pkg/a2s/server"
	"github.com/woozymasta/a2s/pkg/a2s/transport"
)

// proxyStartupRetryDelay is the base delay between startup probe attempts.
const proxyStartupRetryDelay = time.Second

// proxyPreparation contains resources and decisions required before serving.
type proxyPreparation struct {
	upstream    string                 // Normalized upstream endpoint for diagnostics.
	pollClient  *a2s.Client            // Client dedicated to cache refreshes.
	relayClient *a2s.Client            // Client dedicated to live passthrough.
	cache       *proxy.Cache           // Cache seeded by successful startup probes.
	packetizer  server.Packetizer      // Downstream framing selected from INFO.
	policy      server.ChallengePolicy // Downstream challenge policy selected from INFO.
	info        transport.Packet       // Successful startup INFO packet.
	infoMeta    transport.QueryMeta    // Metadata from the startup INFO exchange.
}

// close releases clients owned by a startup preparation.
func (p *proxyPreparation) close() error {
	if p == nil {
		return nil
	}

	var errs []error
	if p.pollClient != nil {
		if err := p.pollClient.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if p.relayClient != nil {
		if err := p.relayClient.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

// prepareProxyStartup validates and probes all state needed before binding.
func prepareProxyStartup(
	ctx context.Context,
	command *ProxyCommand,
	clientOptions ClientOptions,
) (*proxyPreparation, error) {
	if ctx == nil {
		return nil, errors.New("proxy startup context must not be nil")
	}
	if err := validateProxyCommand(command); err != nil {
		return nil, err
	}

	address, err := normalizeEndpoint(command.Args.Host, command.Args.Port)
	if err != nil {
		return nil, fmt.Errorf("invalid upstream endpoint: %w", err)
	}

	preparation := &proxyPreparation{upstream: address}
	cleanup := func(startupErr error) (*proxyPreparation, error) {
		if closeErr := preparation.close(); closeErr != nil {
			startupErr = errors.Join(startupErr, closeErr)
		}
		return nil, startupErr
	}

	preparation.pollClient, err = newProxyUpstreamClient(address, clientOptions)
	if err != nil {
		return cleanup(fmt.Errorf("create polling client: %w", err))
	}
	preparation.relayClient, err = newProxyUpstreamClient(address, clientOptions)
	if err != nil {
		return cleanup(fmt.Errorf("create relay client: %w", err))
	}

	info, infoMeta, err := proxyStartupQuery(
		ctx,
		preparation.pollClient,
		transport.InfoRequest,
		command.CacheOptions.Retries,
		time.Duration(command.CacheOptions.Jitter),
	)
	if err != nil {
		return cleanup(fmt.Errorf("startup INFO probe: %w", err))
	}
	if _, err := a2s.DecodeInfo(info); err != nil {
		return cleanup(fmt.Errorf("startup INFO decode: %w", err))
	}

	preparation.info = info
	preparation.infoMeta = infoMeta
	preparation.packetizer, err = newProxyPacketizer(info.Type)
	if err != nil {
		return cleanup(err)
	}

	if infoMeta.UsedChallenge {
		preparation.policy = server.SecureChallengePolicy()
	} else {
		preparation.policy = server.LegacyChallengePolicy()
	}

	selectors, err := proxyCacheQueries(command.CacheOptions.Cache)
	if err != nil {
		return cleanup(err)
	}

	packets := map[transport.QueryType]transport.Packet{transport.InfoRequest: info}
	enabledQueries := make([]transport.QueryType, 0, len(selectors))
	autoCache := len(command.CacheOptions.Cache) == 1 &&
		command.CacheOptions.Cache[0] == proxyCacheAuto
	for _, query := range selectors {
		if query == transport.InfoRequest {
			enabledQueries = append(enabledQueries, query)
			break
		}
	}

	for _, query := range selectors {
		if query == transport.InfoRequest {
			continue
		}
		if !autoCache {
			enabledQueries = append(enabledQueries, query)
		}

		packet, _, probeErr := proxyStartupQuery(ctx, preparation.pollClient, query, 0, 0)
		if probeErr != nil {
			continue
		}

		packets[query] = packet
		if autoCache {
			enabledQueries = append(enabledQueries, query)
		}
	}

	cache, err := proxy.NewCache(enabledQueries)
	if err != nil {
		return cleanup(fmt.Errorf("create proxy cache: %w", err))
	}

	preparation.cache = cache
	for query, packet := range packets {
		if !cache.Enabled(query) {
			continue
		}
		if err := cache.Store(query, packet); err != nil {
			return cleanup(fmt.Errorf("seed %s cache: %w", proxyQueryName(query), err))
		}
	}

	return preparation, nil
}

// newProxyUpstreamClient creates one client for polling or relay traffic.
func newProxyUpstreamClient(address string, clientOptions ClientOptions) (*a2s.Client, error) {
	options := []a2s.Option{a2s.WithBufferSize(clientOptions.Buffer)}
	if clientOptions.Timeout > 0 {
		options = append(options, a2s.WithTimeout(time.Duration(clientOptions.Timeout)))
	}

	return a2s.NewWithString(address, options...)
}

// proxyStartupQuery executes one startup probe with bounded retries.
func proxyStartupQuery(
	ctx context.Context,
	client *a2s.Client,
	query transport.QueryType,
	retries int,
	jitter time.Duration,
) (transport.Packet, transport.QueryMeta, error) {
	var lastErr error
	for attempt := 0; attempt <= retries; attempt++ {
		packet, meta, err := client.Query(ctx, query)
		if err == nil {
			return packet, meta, nil
		}

		lastErr = err
		if ctx.Err() != nil || attempt == retries {
			break
		}

		if err := waitProxyStartupRetry(ctx, jitter); err != nil {
			return transport.Packet{}, transport.QueryMeta{}, err
		}
	}

	return transport.Packet{}, transport.QueryMeta{}, lastErr
}

// waitProxyStartupRetry waits for the retry base plus bounded jitter.
func waitProxyStartupRetry(ctx context.Context, jitter time.Duration) error {
	delay := proxyStartupRetryDelay + proxyStartupJitter(jitter)
	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

// proxyStartupJitter returns a non-negative startup retry jitter sample.
func proxyStartupJitter(upperBound time.Duration) time.Duration {
	if upperBound <= 0 {
		return 0
	}
	if upperBound == time.Duration(1<<63-1) {
		// #nosec G404 -- startup jitter is timing noise, not a security value.
		return time.Duration(rand.Int63())
	}

	// #nosec G404 -- startup jitter is timing noise, not a security value.
	return time.Duration(rand.Int63n(int64(upperBound) + 1))
}

// newProxyPacketizer creates framing compatible with the upstream INFO format.
func newProxyPacketizer(response transport.ResponseType) (server.Packetizer, error) {
	switch response {
	case transport.ResponseInfo:
		packetizer, err := server.NewSourcePacketizer()
		if err != nil {
			return nil, fmt.Errorf("create Source packetizer: %w", err)
		}
		return packetizer, nil

	case transport.ResponseInfoGoldSource:
		packetizer, err := server.NewGoldSourcePacketizer()
		if err != nil {
			return nil, fmt.Errorf("create GoldSource packetizer: %w", err)
		}
		return packetizer, nil

	default:
		return nil, fmt.Errorf("unsupported INFO response type 0x%X", response)
	}
}

// proxyCacheQueries expands normalized CLI cache selectors into query types.
// The none selector intentionally expands to an empty cache.
func proxyCacheQueries(selectors []string) ([]transport.QueryType, error) {
	if len(selectors) == 0 {
		selectors = []string{proxyCacheAuto}
	}

	queries := make([]transport.QueryType, 0, 3)
	for _, selector := range selectors {
		switch selector {
		case proxyCacheAuto:
			return []transport.QueryType{
				transport.InfoRequest,
				transport.PlayerRequest,
				transport.RulesRequest,
			}, nil

		case proxyCacheNone:
			return nil, nil

		case "info":
			queries = append(queries, transport.InfoRequest)

		case "players":
			queries = append(queries, transport.PlayerRequest)

		case "rules":
			queries = append(queries, transport.RulesRequest)

		default:
			return nil, fmt.Errorf("unsupported cache selector %q", selector)
		}
	}

	return uniqueProxyQueries(queries), nil
}

// uniqueProxyQueries preserves selector order while removing duplicates.
func uniqueProxyQueries(queries []transport.QueryType) []transport.QueryType {
	seen := make(map[transport.QueryType]struct{}, len(queries))
	unique := make([]transport.QueryType, 0, len(queries))
	for _, query := range queries {
		if _, ok := seen[query]; ok {
			continue
		}
		seen[query] = struct{}{}
		unique = append(unique, query)
	}

	return unique
}

// proxyQueryName returns the concise name used in startup diagnostics.
func proxyQueryName(query transport.QueryType) string {
	switch query {
	case transport.InfoRequest:
		return "INFO"

	case transport.PlayerRequest:
		return "PLAYERS"

	case transport.RulesRequest:
		return "RULES"

	default:
		return fmt.Sprintf("0x%X", query)
	}
}
