// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package server

import "github.com/woozymasta/a2s/pkg/a2s"

// ChallengePolicy decides whether a query must include a valid challenge.
//
// The policy only describes challenge requirements.
// It does not issue or validate challenge tokens
// and does not enable deprecated ping or challenge request handling.
type ChallengePolicy interface {
	Required(a2s.QueryType) bool
}

// ChallengePolicyFunc adapts a function to ChallengePolicy.
type ChallengePolicyFunc func(a2s.QueryType) bool

// Required reports whether the adapted function requires a challenge.
func (f ChallengePolicyFunc) Required(query a2s.QueryType) bool {
	return f(query)
}

// SecureChallengePolicy returns the Internet-facing default policy.
//
// INFO, PLAYER, and RULES requests require a challenge.
// Deprecated PING and explicit challenge requests are not enabled by this policy.
func SecureChallengePolicy() ChallengePolicy {
	return ChallengePolicyFunc(func(query a2s.QueryType) bool {
		switch query {
		case a2s.InfoRequest, a2s.PlayerRequest, a2s.RulesRequest:
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
// Deprecated PING and explicit challenge requests are not enabled by this policy.
func LegacyChallengePolicy() ChallengePolicy {
	return ChallengePolicyFunc(func(query a2s.QueryType) bool {
		return query == a2s.PlayerRequest || query == a2s.RulesRequest
	})
}

// NoChallengePolicy returns a policy for controlled local or test use.
func NoChallengePolicy() ChallengePolicy {
	return ChallengePolicyFunc(func(a2s.QueryType) bool {
		return false
	})
}
