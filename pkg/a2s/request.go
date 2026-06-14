package a2s

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
)

const requestHeaderSize = 5

// DecodeRequest parses one complete A2S request datagram.
//
// The decoder accepts only the documented exact wire forms.
// It does not accept trailing bytes because they could change the meaning of a request
// when the datagram is forwarded or re-encoded.
func DecodeRequest(data []byte) (Request, error) {
	if len(data) < requestHeaderSize {
		return Request{}, errors.Join(ErrRequestHeader, ErrInsufficientData)
	}

	if binary.BigEndian.Uint32(data[:4]) != singlePacket {
		return Request{}, ErrRequestHeader
	}

	request := Request{Type: QueryType(data[4])}
	payload := data[requestHeaderSize:]

	switch request.Type {
	case InfoRequest:
		return decodeInfoRequest(payload, request)

	case PlayerRequest, RulesRequest:
		if len(payload) != len(InitialChallenge) {
			return Request{}, invalidRequestPayload(len(payload), len(InitialChallenge))
		}

		copy(request.Challenge[:], payload)
		request.HasChallenge = true
		return request, nil

	case ChallengeRequest, PingRequest:
		if len(payload) != 0 {
			return Request{}, invalidRequestPayload(len(payload), 0)
		}

		return request, nil

	default:
		return Request{}, fmt.Errorf("%w: 0x%X", ErrWrongRequest, request.Type)
	}
}

func decodeInfoRequest(payload []byte, request Request) (Request, error) {
	base := append([]byte(infoPayload), 0)

	switch len(payload) {
	case len(base):
		if !bytes.Equal(payload, base) {
			return Request{}, invalidRequestPayload(len(payload), len(base))
		}

		return request, nil

	case len(base) + len(request.Challenge):
		if !bytes.Equal(payload[:len(base)], base) {
			return Request{}, invalidRequestPayload(len(payload), len(base))
		}

		copy(request.Challenge[:], payload[len(base):])
		request.HasChallenge = true
		return request, nil

	default:
		return Request{}, invalidRequestPayload(len(payload), len(base))
	}
}

// AppendRequest appends one complete A2S request datagram to dst.
func AppendRequest(dst []byte, request Request) ([]byte, error) {
	switch request.Type {
	case InfoRequest:
		dst = appendRequestHeader(dst, request.Type)
		dst = append(dst, infoPayload...)
		dst = append(dst, 0)
		if request.HasChallenge {
			dst = append(dst, request.Challenge[:]...)
		}
		return dst, nil

	case PlayerRequest, RulesRequest:
		if !request.HasChallenge {
			return dst, fmt.Errorf("%w: challenge is required", ErrRequestPayload)
		}

		dst = appendRequestHeader(dst, request.Type)
		dst = append(dst, request.Challenge[:]...)
		return dst, nil

	case ChallengeRequest, PingRequest:
		if request.HasChallenge {
			return dst, fmt.Errorf("%w: challenge is not allowed", ErrRequestPayload)
		}

		return appendRequestHeader(dst, request.Type), nil

	default:
		return dst, fmt.Errorf("%w: 0x%X", ErrWrongRequest, request.Type)
	}
}

func appendRequestHeader(dst []byte, requestType QueryType) []byte {
	dst = binary.BigEndian.AppendUint32(dst, singlePacket)
	return append(dst, byte(requestType))
}

func invalidRequestPayload(got, want int) error {
	return fmt.Errorf("%w: got %d bytes, want %d", ErrRequestPayload, got, want)
}
