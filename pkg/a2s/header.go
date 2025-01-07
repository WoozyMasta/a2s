package a2s

import (
	"encoding/binary"
	"fmt"
)

// Create header for request type byte
//   - InfoRequest      = 0x54
//   - PlayerRequest    = 0x55
//   - RulesRequest     = 0x56
//   - ChallengeRequest = 0x57 (DEPRECATED)
//   - PingRequest      = 0x69 (DEPRECATED)
func createHeader(requestType Flag, challenge uint32) ([]byte, error) {
	req := make([]byte, 4)
	binary.BigEndian.PutUint32(req, singlePacket)
	req = append(req, byte(requestType))

	switch requestType {
	case InfoRequest:
		req = append(req[:], []byte(infoPayload)...)
		req = append(req[:], 0x00)
		if challenge != singlePacket {
			req = binary.BigEndian.AppendUint32(req, challenge)
		}
		return req, nil

	case PlayerRequest, RulesRequest:
		req = binary.BigEndian.AppendUint32(req, challenge)
		return req, nil

	case PingRequest, ChallengeRequest:
		return req, nil

	default:
		return nil, fmt.Errorf("%w: 0x%X", ErrWrongRequest, requestType)
	}
}
