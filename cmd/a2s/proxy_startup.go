package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/woozymasta/a2s/pkg/a2s"
	proxycache "github.com/woozymasta/a2s/pkg/a2s/proxy"
	"github.com/woozymasta/a2s/pkg/a2s/server"
)

// proxyStartupRetryDelay is the base delay between startup probe attempts.
const proxyStartupRetryDelay = time.Second

// proxyPreparation contains resources and decisions required before serving.
type proxyPreparation struct {
	pollClient  *a2s.Client
	relayClient *a2s.Client
	cache       *proxycache.Cache
	packetizer  server.Packetizer
	policy      server.ChallengePolicy
	info        a2s.Packet
	infoMeta    a2s.QueryMeta
}

// executeProxy prepares upstream state before the runtime listener is added.
func executeProxy(_ *Application, command *ProxyCommand) error {
	preparation, err := prepareProxyStartup(context.Background(), command)
	if err != nil {
		return err
	}
	defer func() { _ = preparation.close() }()

	return fmt.Errorf("proxy command runtime is not implemented yet")
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
func prepareProxyStartup(ctx context.Context, command *ProxyCommand) (*proxyPreparation, error) {
	if ctx == nil {
		return nil, fmt.Errorf("proxy startup context must not be nil")
	}
	if err := validateProxyCommand(command); err != nil {
		return nil, err
	}

	address, err := normalizeEndpoint(command.Args.Host, command.Args.Port)
	if err != nil {
		return nil, fmt.Errorf("invalid upstream endpoint: %w", err)
	}

	preparation := &proxyPreparation{}
	cleanup := func(startupErr error) (*proxyPreparation, error) {
		if closeErr := preparation.close(); closeErr != nil {
			startupErr = errors.Join(startupErr, closeErr)
		}
		return nil, startupErr
	}

	preparation.pollClient, err = newProxyUpstreamClient(address, command)
	if err != nil {
		return cleanup(fmt.Errorf("create polling client: %w", err))
	}
	preparation.relayClient, err = newProxyUpstreamClient(address, command)
	if err != nil {
		return cleanup(fmt.Errorf("create relay client: %w", err))
	}

	info, infoMeta, err := proxyStartupQuery(
		ctx,
		preparation.pollClient,
		a2s.InfoRequest,
		command.Retries,
		command.Jitter,
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

	selectors, err := proxyCacheQueries(command.Cache)
	if err != nil {
		return cleanup(err)
	}

	packets := map[a2s.QueryType]a2s.Packet{a2s.InfoRequest: info}
	enabledQueries := make([]a2s.QueryType, 0, len(selectors))
	for _, query := range selectors {
		if query == a2s.InfoRequest {
			enabledQueries = append(enabledQueries, query)
			break
		}
	}

	for _, query := range selectors {
		if query == a2s.InfoRequest {
			continue
		}

		packet, _, probeErr := proxyStartupQuery(ctx, preparation.pollClient, query, 0, 0)
		if probeErr != nil {
			continue
		}

		packets[query] = packet
		enabledQueries = append(enabledQueries, query)
	}

	cache, err := proxycache.NewCache(enabledQueries)
	if err != nil {
		return cleanup(fmt.Errorf("create proxy cache: %w", err))
	}

	preparation.cache = cache
	for query, packet := range packets {
		if !cache.Enabled(query) {
			continue
		}
		if err := cache.Store(query, packet, time.Now()); err != nil {
			return cleanup(fmt.Errorf("seed %s cache: %w", proxyQueryName(query), err))
		}
	}

	return preparation, nil
}

// newProxyUpstreamClient creates one client for polling or relay traffic.
func newProxyUpstreamClient(address string, command *ProxyCommand) (*a2s.Client, error) {
	options := []a2s.Option{a2s.WithBufferSize(command.Buffer)}
	if command.Timeout > 0 {
		options = append(options, a2s.WithTimeout(command.Timeout))
	}

	return a2s.NewWithString(address, options...)
}

// proxyStartupQuery executes one startup probe with bounded retries.
func proxyStartupQuery(
	ctx context.Context,
	client *a2s.Client,
	query a2s.QueryType,
	retries int,
	jitter time.Duration,
) (a2s.Packet, a2s.QueryMeta, error) {
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
			return a2s.Packet{}, a2s.QueryMeta{}, err
		}
	}

	return a2s.Packet{}, a2s.QueryMeta{}, lastErr
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
func newProxyPacketizer(response a2s.ResponseType) (server.Packetizer, error) {
	switch response {
	case a2s.ResponseInfo:
		packetizer, err := server.NewSourcePacketizer()
		if err != nil {
			return nil, fmt.Errorf("create Source packetizer: %w", err)
		}
		return packetizer, nil

	case a2s.ResponseInfoGoldSource:
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
func proxyCacheQueries(selectors []string) ([]a2s.QueryType, error) {
	if len(selectors) == 0 {
		selectors = []string{proxyCacheAuto}
	}

	queries := make([]a2s.QueryType, 0, 3)
	for _, selector := range selectors {
		switch selector {
		case proxyCacheAuto:
			return []a2s.QueryType{
				a2s.InfoRequest,
				a2s.PlayerRequest,
				a2s.RulesRequest,
			}, nil

		case "info":
			queries = append(queries, a2s.InfoRequest)

		case "players":
			queries = append(queries, a2s.PlayerRequest)

		case "rules":
			queries = append(queries, a2s.RulesRequest)

		default:
			return nil, fmt.Errorf("unsupported cache selector %q", selector)
		}
	}

	return uniqueProxyQueries(queries), nil
}

// uniqueProxyQueries preserves selector order while removing duplicates.
func uniqueProxyQueries(queries []a2s.QueryType) []a2s.QueryType {
	seen := make(map[a2s.QueryType]struct{}, len(queries))
	unique := make([]a2s.QueryType, 0, len(queries))
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
func proxyQueryName(query a2s.QueryType) string {
	switch query {
	case a2s.InfoRequest:
		return "INFO"

	case a2s.PlayerRequest:
		return "PLAYERS"

	case a2s.RulesRequest:
		return "RULES"

	default:
		return fmt.Sprintf("0x%X", query)
	}
}
