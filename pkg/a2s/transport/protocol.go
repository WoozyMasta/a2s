// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package transport

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
)

// ParseChallengePayload extracts the first four challenge bytes from a response payload.
// Additional payload bytes are ignored for compatibility.
func ParseChallengePayload(data []byte) (Challenge, error) {
	return parseChallenge(data)
}

// ParseChallengeResponse extracts a challenge
// from a complete single-packet challenge response,
// including its packet marker and response byte.
func ParseChallengeResponse(data []byte) (Challenge, error) {
	packet, err := DecodePacket(data)
	if err != nil {
		return Challenge{}, errors.Join(ErrChallengeRead, err)
	}
	if packet.Type != ResponseChallenge {
		return Challenge{}, errors.Join(ErrChallengeRead, ErrValidatorChallenge)
	}

	challenge, err := parseChallenge(packet.Payload)
	if err != nil {
		return Challenge{}, errors.Join(ErrChallengeRead, err)
	}

	return challenge, nil
}

// parseChallenge copies the four-byte opaque challenge token from a payload.
func parseChallenge(data []byte) (Challenge, error) {
	if len(data) < len(Challenge{}) {
		return Challenge{}, errors.Join(
			ErrChallengeValue,
			ErrInsufficientData,
			fmt.Errorf("got %d bytes, want %d", len(data), len(Challenge{})),
		)
	}

	var challenge Challenge
	copy(challenge[:], data[:len(challenge)])

	return challenge, nil
}

// CreateHeader encodes a request and appends the challenge
// where the wire format requires it.
func CreateHeader(requestType QueryType, challenge Challenge) ([]byte, error) {
	var hasChallenge bool
	switch requestType {
	case InfoRequest:
		hasChallenge = challenge != initialChallenge

	case PlayerRequest, RulesRequest:
		hasChallenge = true

	case ChallengeRequest, PingRequest:
		// These requests never carry a challenge value.

	default:
		return nil, ErrHeaderWrongRequest
	}

	request := make([]byte, 0, PacketHeaderSize+len(challenge)+len(infoPayload)+1)
	return AppendRequest(request, Request{
		Type:         requestType,
		Challenge:    challenge,
		HasChallenge: hasChallenge,
	})
}

// DecodePacket validates single-packet framing
// and copies the payload away from the caller's reusable receive buffer.
func DecodePacket(data []byte) (Packet, error) {
	if len(data) < PacketHeaderSize {
		return Packet{}, errors.Join(ErrPacketHeader, ErrInsufficientData)
	}
	if binary.LittleEndian.Uint32(data[:4]) != SinglePacketMarker {
		return Packet{}, ErrPacketHeader
	}

	return Packet{
		Type:    ResponseType(data[4]),
		Payload: bytes.Clone(data[PacketHeaderSize:]),
	}, nil
}

// ValidateResponseType checks that the response byte matches the request.
func ValidateResponseType(request QueryType, response ResponseType) error {
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

// validationErrForRequest returns the package-local error
// for an unsupported request type after all fallback attempts are exhausted.
func validationErrForRequest(requestType QueryType) error {
	switch requestType {
	case InfoRequest:
		return ErrValidatorInfo

	case PlayerRequest:
		return ErrValidatorPlayer

	case RulesRequest:
		return ErrValidatorRules

	case PingRequest:
		return ErrValidatorPing

	case ChallengeRequest:
		return ErrValidatorChallenge

	default:
		return ErrValidatorRequest
	}
}

// contextError prefers cancellation/deadline errors from ctx
// over the platform-specific timeout returned by a UDP socket.
func contextError(ctx context.Context, err error) error {
	if ctxErr := ctx.Err(); ctxErr != nil {
		return ctxErr
	}

	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		if _, ok := ctx.Deadline(); ok {
			return context.DeadlineExceeded
		}
	}

	return err
}
