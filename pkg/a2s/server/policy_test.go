// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package server_test

import (
	"testing"

	"github.com/woozymasta/a2s/pkg/a2s/server"
	"github.com/woozymasta/a2s/pkg/a2s/transport"
)

func TestChallengePolicyPresets(t *testing.T) {
	tests := []struct {
		name     string
		policy   server.ChallengePolicy
		required map[transport.QueryType]bool
	}{
		{
			name:   "secure",
			policy: server.SecureChallengePolicy(),
			required: map[transport.QueryType]bool{
				transport.InfoRequest:      true,
				transport.PlayerRequest:    true,
				transport.RulesRequest:     true,
				transport.PingRequest:      false,
				transport.ChallengeRequest: false,
			},
		},
		{
			name:   "legacy",
			policy: server.LegacyChallengePolicy(),
			required: map[transport.QueryType]bool{
				transport.InfoRequest:      false,
				transport.PlayerRequest:    true,
				transport.RulesRequest:     true,
				transport.PingRequest:      false,
				transport.ChallengeRequest: false,
			},
		},
		{
			name:   "none",
			policy: server.NoChallengePolicy(),
			required: map[transport.QueryType]bool{
				transport.InfoRequest:      false,
				transport.PlayerRequest:    false,
				transport.RulesRequest:     false,
				transport.PingRequest:      false,
				transport.ChallengeRequest: false,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for query, want := range test.required {
				if got := test.policy.Required(query); got != want {
					t.Errorf("Required(0x%X) = %t, want %t", query, got, want)
				}
			}
		})
	}
}

func TestChallengePolicyFunc(t *testing.T) {
	policy := server.ChallengePolicyFunc(func(query transport.QueryType) bool {
		return query == transport.InfoRequest
	})

	if !policy.Required(transport.InfoRequest) {
		t.Error("ChallengePolicyFunc did not require INFO challenge")
	}
	if policy.Required(transport.RulesRequest) {
		t.Error("ChallengePolicyFunc unexpectedly required RULES challenge")
	}
}
