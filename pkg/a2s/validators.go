package a2s

import (
	"encoding/binary"
	"fmt"
)

// Checks the packet header and returns true if the response is multi-packet
func isMultiPacket(data []byte) (bool, error) {
	header := binary.LittleEndian.Uint32(data[:4])

	switch header {
	case singlePacket:
		if len(data) < 5 {
			return false, ErrSinglePacket
		}
		return false, nil

	case multiPacket:
		if len(data) < 10 {
			return true, ErrMultiPacket
		}
		return true, nil

	default:
		return false, fmt.Errorf("%w in header: 0x%X", ErrWrongByte, header)
	}
}

// Checks the response type in the packet against the type that was sent
func validateResponseType(request, response Flag) error {
	switch request {
	case InfoRequest:
		if response != infoResponseSource && response != infoResponseGoldSource {
			return fmt.Errorf("%w: 0x%X for A2S_INFO", ErrWrongByte, response)
		}

	case PlayerRequest:
		if response != playerResponse {
			return fmt.Errorf("%w: 0x%X for A2S_PLAYER", ErrWrongByte, response)
		}

	case RulesRequest:
		if response != rulesResponse {
			return fmt.Errorf("%w: 0x%X for A2S_RULES", ErrWrongByte, response)
		}

	case PingRequest:
		if response != pingResponse {
			return fmt.Errorf("%w: 0x%X for A2A_PING", ErrWrongByte, response)
		}

	case ChallengeRequest:
		if response != challengeResponse {
			return fmt.Errorf("%w: 0x%X for A2S_SERVERQUERY_GETCHALLENGE", ErrWrongByte, response)
		}

	default:
		return fmt.Errorf("%w: 0x%X", ErrWrongRequest, request)
	}

	return nil
}
