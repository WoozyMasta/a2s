// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package a2s

// createHeader builds A2S protocol request header.
//
// InfoRequest includes "Source Engine Query" payload, other requests include challenge value.
//   - InfoRequest      = 0x54
//   - PlayerRequest    = 0x55
//   - RulesRequest     = 0x56
//   - ChallengeRequest = 0x57 (DEPRECATED)
//   - PingRequest      = 0x69 (DEPRECATED)
func createHeader(requestType QueryType, challenge Challenge) ([]byte, error) {
	var hasChallenge bool
	switch requestType {
	case InfoRequest:
		hasChallenge = challenge != InitialChallenge

	case PlayerRequest, RulesRequest:
		hasChallenge = true

	case ChallengeRequest, PingRequest:

	default:
		return nil, ErrHeaderWrongRequest
	}

	request := Request{
		Type:         requestType,
		Challenge:    challenge,
		HasChallenge: hasChallenge,
	}

	return AppendRequest(nil, request)
}
