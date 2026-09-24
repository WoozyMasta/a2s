// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package a2s

import (
	"encoding/binary"
	"errors"
	"fmt"
)

// isMultiPacket checks if response uses multi-packet format.
func isMultiPacket(data []byte) (bool, error) {
	// Some servers can send a truncated packet first; avoid panics on short reads.
	if len(data) < 4 {
		return true, ErrMultiPacket
	}

	header := binary.LittleEndian.Uint32(data[:4])

	switch header {
	case singlePacket:
		if len(data) < 5 {
			return false, ErrSinglePacket
		}
		return false, nil

	case multiPacket:
		if len(data) < 9 {
			return true, ErrMultiPacket
		}
		return true, nil

	default:
		return false, errors.Join(ErrValidatorHeader, fmt.Errorf("0x%X", header))
	}
}

// validateResponseType verifies response type matches the request type.
func validateResponseType(request QueryType, response ResponseType) error {
	switch request {
	case InfoRequest:
		if response != ResponseInfo && response != ResponseInfoGoldSource {
			return errors.Join(ErrValidatorInfo, fmt.Errorf("0x%X", response))
		}

	case PlayerRequest:
		if response != ResponsePlayers {
			return errors.Join(ErrValidatorPlayer, fmt.Errorf("0x%X", response))
		}

	case RulesRequest:
		if response != ResponseRules {
			return errors.Join(ErrValidatorRules, fmt.Errorf("0x%X", response))
		}

	case PingRequest:
		if response != ResponsePing {
			return errors.Join(ErrValidatorPing, fmt.Errorf("0x%X", response))
		}

	case ChallengeRequest:
		if response != ResponseChallenge {
			return errors.Join(ErrValidatorChallenge, fmt.Errorf("0x%X", response))
		}

	default:
		return errors.Join(ErrValidatorRequest, fmt.Errorf("0x%X", request))
	}

	return nil
}
