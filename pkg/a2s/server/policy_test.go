// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package server_test

import (
	"testing"

	"github.com/woozymasta/a2s/pkg/a2s"
	"github.com/woozymasta/a2s/pkg/a2s/server"
)

func TestChallengePolicyPresets(t *testing.T) {
	tests := []struct {
		name     string
		policy   server.ChallengePolicy
		required map[a2s.QueryType]bool
	}{
		{
			name:   "secure",
			policy: server.SecureChallengePolicy(),
			required: map[a2s.QueryType]bool{
				a2s.InfoRequest:      true,
				a2s.PlayerRequest:    true,
				a2s.RulesRequest:     true,
				a2s.PingRequest:      false,
				a2s.ChallengeRequest: false,
			},
		},
		{
			name:   "legacy",
			policy: server.LegacyChallengePolicy(),
			required: map[a2s.QueryType]bool{
				a2s.InfoRequest:      false,
				a2s.PlayerRequest:    true,
				a2s.RulesRequest:     true,
				a2s.PingRequest:      false,
				a2s.ChallengeRequest: false,
			},
		},
		{
			name:   "none",
			policy: server.NoChallengePolicy(),
			required: map[a2s.QueryType]bool{
				a2s.InfoRequest:      false,
				a2s.PlayerRequest:    false,
				a2s.RulesRequest:     false,
				a2s.PingRequest:      false,
				a2s.ChallengeRequest: false,
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
	policy := server.ChallengePolicyFunc(func(query a2s.QueryType) bool {
		return query == a2s.InfoRequest
	})

	if !policy.Required(a2s.InfoRequest) {
		t.Error("ChallengePolicyFunc did not require INFO challenge")
	}
	if policy.Required(a2s.RulesRequest) {
		t.Error("ChallengePolicyFunc unexpectedly required RULES challenge")
	}
}
