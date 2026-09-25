// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package server

import "github.com/woozymasta/a2s/pkg/a2s/transport"

// ChallengePolicy decides whether a query must include a valid challenge.
//
// The policy only describes challenge requirements.
// It does not issue or validate challenge tokens
// and does not enable obsolete ping or challenge wire-request handling.
type ChallengePolicy interface {
	Required(transport.QueryType) bool
}

// ChallengePolicyFunc adapts a function to ChallengePolicy.
type ChallengePolicyFunc func(transport.QueryType) bool

// Required reports whether the adapted function requires a challenge.
func (f ChallengePolicyFunc) Required(query transport.QueryType) bool {
	return f(query)
}

// SecureChallengePolicy returns the Internet-facing default policy.
//
// INFO, PLAYER, and RULES requests require a challenge.
// Obsolete PING and explicit challenge wire requests are not enabled by this policy.
func SecureChallengePolicy() ChallengePolicy {
	return ChallengePolicyFunc(func(query transport.QueryType) bool {
		switch query {
		case transport.InfoRequest, transport.PlayerRequest, transport.RulesRequest:
			return true
		default:
			return false
		}
	})
}

// LegacyChallengePolicy returns the compatibility policy
// used by older A2S servers, where INFO is answered without a challenge.
//
// PLAYER and RULES requests require a challenge.
// Obsolete PING and explicit challenge wire requests are not enabled by this policy.
func LegacyChallengePolicy() ChallengePolicy {
	return ChallengePolicyFunc(func(query transport.QueryType) bool {
		return query == transport.PlayerRequest || query == transport.RulesRequest
	})
}

// NoChallengePolicy returns a policy for controlled local or test use.
func NoChallengePolicy() ChallengePolicy {
	return ChallengePolicyFunc(func(transport.QueryType) bool {
		return false
	})
}
