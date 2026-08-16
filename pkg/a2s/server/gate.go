// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package server

import (
	"context"
	"fmt"

	"github.com/woozymasta/a2s/pkg/a2s"
)

// ChallengeGate validates challenge-protected requests
// before forwarding them to the wrapped handler.
type ChallengeGate struct {
	next     Handler
	policy   ChallengePolicy
	provider ChallengeProvider
}

// Ensure ChallengeGate remains a Handler implementation.
var _ Handler = (*ChallengeGate)(nil)

// NewChallengeGate wraps next with the supplied challenge policy and provider.
func NewChallengeGate(
	next Handler,
	policy ChallengePolicy,
	provider ChallengeProvider,
) (*ChallengeGate, error) {
	if next == nil {
		return nil, fmt.Errorf("%w: handler is nil", ErrChallengeGate)
	}
	if policy == nil {
		return nil, fmt.Errorf("%w: policy is nil", ErrChallengeGate)
	}
	if provider == nil {
		return nil, fmt.Errorf("%w: provider is nil", ErrChallengeGate)
	}

	return &ChallengeGate{
		next:     next,
		policy:   policy,
		provider: provider,
	}, nil
}

// Handle forwards a request when its challenge is valid
// or returns a new challenge packet without invoking the wrapped handler.
func (g *ChallengeGate) Handle(ctx context.Context, req *Request) (Response, error) {
	if req == nil {
		return nil, fmt.Errorf("%w: request is nil", ErrChallengeGate)
	}
	if g == nil || g.next == nil || g.policy == nil || g.provider == nil {
		return nil, fmt.Errorf("%w: gate is not initialized", ErrChallengeGate)
	}

	if !g.policy.Required(req.Query.Type) {
		return g.next.Handle(ctx, req)
	}
	if req.Query.HasChallenge && g.provider.Validate(req.Remote, req.Query.Challenge) {
		return g.next.Handle(ctx, req)
	}

	return challengeResponse(g.provider.Issue(req.Remote)), nil
}

// challengeResponse creates the logical A2S challenge response packet.
func challengeResponse(challenge a2s.Challenge) Response {
	return PacketResponse{
		Packet: a2s.Packet{
			Type:    a2s.ResponseChallenge,
			Payload: append([]byte(nil), challenge[:]...),
		},
	}
}
