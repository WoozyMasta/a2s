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
	var req []byte
	payloadLen := len(infoPayload)

	switch requestType {
	case InfoRequest:
		// Pre-allocate with exact capacity: 4 (header) + 1 (type) + payload + 1 (null) + 4 (challenge, optional)
		capacity := 4 + 1 + payloadLen + 1
		if challenge != singlePacket {
			capacity += 4
		}
		req = make([]byte, 0, capacity)
		req = binary.BigEndian.AppendUint32(req, singlePacket)
		req = append(req, byte(requestType))
		req = append(req, []byte(infoPayload)...)
		req = append(req, 0x00)
		if challenge != singlePacket {
			req = binary.BigEndian.AppendUint32(req, challenge)
		}
		return req, nil

	case PlayerRequest, RulesRequest:
		// Pre-allocate with exact capacity: 4 (header) + 1 (type) + 4 (challenge)
		req = make([]byte, 0, 9)
		req = binary.BigEndian.AppendUint32(req, singlePacket)
		req = append(req, byte(requestType))
		req = binary.BigEndian.AppendUint32(req, challenge)
		return req, nil

	case PingRequest, ChallengeRequest:
		// Pre-allocate with exact capacity: 4 (header) + 1 (type)
		req = make([]byte, 0, 5)
		req = binary.BigEndian.AppendUint32(req, singlePacket)
		req = append(req, byte(requestType))
		return req, nil

	default:
		return nil, fmt.Errorf("%w: 0x%X", ErrWrongRequest, requestType)
	}
}
