package proxy

import (
	"context"
	"fmt"

	"github.com/woozymasta/a2s/pkg/a2s"
	"github.com/woozymasta/a2s/pkg/a2s/server"
)

// HandlerConfig configures downstream request handling.
type HandlerConfig struct {
	// LocalPing answers deprecated A2A_PING requests without contacting upstream.
	LocalPing bool
	// ChallengeProvider issues tokens for local GETCHALLENGE responses.
	ChallengeProvider server.ChallengeProvider
}

// Handler routes downstream requests through the cache, relay, or local
// protocol responses.
type Handler struct {
	cache     *Cache                   // Cache for enabled INFO/PLAYER/RULES queries.
	relay     Upstream                 // Upstream used for live passthrough.
	provider  server.ChallengeProvider // Provider shared with the downstream server.
	localPing bool                     // Whether deprecated PING is answered locally.
}

// Ensure Handler implements server.Handler.
var _ server.Handler = (*Handler)(nil)

// NewHandler creates a proxy handler over a cache and live relay upstream.
func NewHandler(cache *Cache, relay Upstream, config HandlerConfig) (*Handler, error) {
	if cache == nil {
		return nil, fmt.Errorf("%w: cache is nil", ErrHandler)
	}
	if relay == nil {
		return nil, fmt.Errorf("%w: relay is nil", ErrHandler)
	}
	if config.ChallengeProvider == nil {
		return nil, fmt.Errorf("%w: challenge provider is nil", ErrHandler)
	}

	return &Handler{
		cache:     cache,
		relay:     relay,
		provider:  config.ChallengeProvider,
		localPing: config.LocalPing,
	}, nil
}

// Handle routes one decoded downstream request.
//
// Cache misses and expected relay failures return server.ErrDrop
// so the transport silently produces no response for unavailable data.
func (h *Handler) Handle(ctx context.Context, request *server.Request) (server.Response, error) {
	if h == nil || h.cache == nil || h.relay == nil || h.provider == nil {
		return nil, fmt.Errorf("%w: handler is not initialized", ErrHandler)
	}
	if request == nil {
		return nil, fmt.Errorf("%w: request is nil", ErrHandler)
	}

	switch request.Query.Type {
	case a2s.InfoRequest, a2s.PlayerRequest, a2s.RulesRequest:
		if h.cache.Enabled(request.Query.Type) {
			return h.cachedResponse(request.Query.Type)
		}
		return h.relayResponse(ctx, request.Query.Type)

	case a2s.ChallengeRequest:
		return challengeResponse(h.provider.Issue(request.Remote)), nil

	case a2s.PingRequest:
		if h.localPing {
			return server.PacketResponse{Packet: a2s.Packet{Type: a2s.ResponsePing}}, nil
		}
		return h.relayResponse(ctx, a2s.PingRequest)

	default:
		return nil, server.ErrDrop
	}
}

// cachedResponse returns one cloned packet from the selected cache entry.
func (h *Handler) cachedResponse(query a2s.QueryType) (server.Response, error) {
	packet, ok := h.cache.Load(query)
	if !ok {
		return nil, server.ErrDrop
	}

	return server.PacketResponse{Packet: packet}, nil
}

// relayResponse queries upstream and converts expected failures into drops.
func (h *Handler) relayResponse(ctx context.Context, query a2s.QueryType) (server.Response, error) {
	packet, _, err := h.relay.Query(ctx, query)
	if err != nil {
		return nil, server.ErrDrop
	}

	return server.PacketResponse{Packet: packet}, nil
}

// challengeResponse creates a local logical GETCHALLENGE response.
func challengeResponse(challenge a2s.Challenge) server.Response {
	return server.PacketResponse{
		Packet: a2s.Packet{
			Type:    a2s.ResponseChallenge,
			Payload: append([]byte(nil), challenge[:]...),
		},
	}
}
