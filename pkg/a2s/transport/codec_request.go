// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package transport

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// Request represents one complete A2S request datagram.
//
// INFO may omit its challenge on the initial request.
// PLAYER and RULES always require a challenge.
// HasChallenge preserves that wire-level distinction.
type Request struct {
	// Type is the request type byte.
	Type QueryType
	// Challenge is the opaque four-byte challenge token.
	Challenge Challenge
	// HasChallenge reports whether Challenge is present on the wire.
	HasChallenge bool
}

// DecodeRequest parses one complete A2S request datagram.
func DecodeRequest(data []byte) (Request, error) {
	if len(data) < PacketHeaderSize {
		return Request{}, fmt.Errorf("%w: %w", ErrRequestHeader, ErrInsufficientData)
	}
	if binary.BigEndian.Uint32(data[:4]) != SinglePacketMarker {
		return Request{}, ErrRequestHeader
	}

	request := Request{Type: QueryType(data[4])}
	payload := data[PacketHeaderSize:]

	switch request.Type {
	case InfoRequest:
		return decodeInfoRequest(payload, request)

	case PlayerRequest, RulesRequest:
		if len(payload) != len(initialChallenge) {
			return Request{}, invalidRequestPayload(len(payload), len(initialChallenge))
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

// decodeInfoRequest validates the fixed INFO payload and optional challenge.
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

// appendRequestHeader appends the common A2S request framing.
func appendRequestHeader(dst []byte, requestType QueryType) []byte {
	dst = binary.BigEndian.AppendUint32(dst, SinglePacketMarker)
	return append(dst, byte(requestType))
}

// invalidRequestPayload describes an exact wire-length mismatch.
func invalidRequestPayload(got, want int) error {
	return fmt.Errorf("%w: got %d bytes, want %d", ErrRequestPayload, got, want)
}
