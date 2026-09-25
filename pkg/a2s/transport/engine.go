// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package transport

import (
	"context"
	"errors"
	"time"
)

// Execute runs one complete logical A2S transaction over io.
//
// The caller owns readBuf and must serialize Execute calls
// that use the same buffer and socket.
// The returned payload owns its backing storage.
func Execute(
	ctx context.Context,
	io DatagramIO,
	requestType QueryType,
	readBuf []byte,
) (Packet, QueryMeta, error) {
	if len(readBuf) == 0 {
		readBuf = make([]byte, DefaultBufferSize)
	}

	started := time.Now()
	var (
		lastUnexpectedErr      error
		lastUnexpectedResponse ResponseType
		usedChallenge          bool
	)

	for range maxUnsupportedResponses {
		resp, responseType, attemptUsedChallenge, err := requestWithChallenge(
			ctx,
			io,
			requestType,
			readBuf,
		)
		usedChallenge = usedChallenge || attemptUsedChallenge
		meta := transactionMeta(started, usedChallenge)
		if err != nil {
			if lastUnexpectedErr != nil {
				return Packet{Type: lastUnexpectedResponse}, meta, errors.Join(lastUnexpectedErr, err)
			}

			return Packet{Type: responseType}, meta, err
		}

		packet, err := DecodePacket(resp)
		if err != nil {
			return Packet{}, meta, err
		}

		validationErr := ValidateResponseType(requestType, packet.Type)
		if validationErr == nil {
			return packet, meta, nil
		}

		classifiedErr := classifyResponseError(requestType, packet.Type, validationErr)
		if shouldRetryResponse(requestType, packet.Type) {
			// Some servers answer an unsupported query with A2S_INFO or ask for a challenge again.
			// Keep the error and try the query once more.
			lastUnexpectedErr = classifiedErr
			lastUnexpectedResponse = packet.Type
			continue
		}

		return packet, meta, classifiedErr
	}

	if lastUnexpectedErr != nil {
		return Packet{Type: lastUnexpectedResponse}, transactionMeta(started, usedChallenge), lastUnexpectedErr
	}

	return Packet{}, transactionMeta(started, usedChallenge), validationErrForRequest(requestType)
}

// transactionMeta snapshots the cumulative transaction state for an error
// or successful response returned from the current attempt.
func transactionMeta(started time.Time, usedChallenge bool) QueryMeta {
	return QueryMeta{
		Duration:      time.Since(started),
		UsedChallenge: usedChallenge,
	}
}

// requestWithChallenge repeats a request when the server returns a challenge.
func requestWithChallenge(
	ctx context.Context,
	io DatagramIO,
	requestType QueryType,
	readBuf []byte,
) ([]byte, ResponseType, bool, error) {
	challenge := initialChallenge

	for attempt := range maxChallengeResponses {
		resp, err := request(ctx, io, requestType, challenge, readBuf)
		if err != nil {
			return nil, 0, attempt > 0, err
		}
		if len(resp) < PacketHeaderSize {
			return resp, 0, attempt > 0, errors.Join(ErrPacketHeader, ErrInsufficientData)
		}

		responseType := ResponseType(resp[4])
		if responseType != ResponseChallenge ||
			requestType == ChallengeRequest || requestType == PingRequest {
			return resp, responseType, attempt > 0, nil
		}

		if attempt == maxChallengeResponses-1 {
			return resp, ResponseChallenge, true, ErrChallengeLoop
		}

		// The next request carries the token returned by this challenge packet.
		challenge, err = ParseChallengeResponse(resp)
		if err != nil {
			return resp, ResponseChallenge, true, err
		}
	}

	return nil, ResponseChallenge, true, ErrChallengeLoop
}

// classifyResponseError adds the reason that makes a validation failure
// actionable to callers without changing the original validation error.
func classifyResponseError(
	requestType QueryType,
	responseType ResponseType,
	validationErr error,
) error {
	switch {
	case responseType == ResponseChallenge:
		return errors.Join(validationErr, ErrChallengeLoop)

	case requestType != InfoRequest &&
		(responseType == ResponseInfo || responseType == ResponseInfoGoldSource):
		return errors.Join(validationErr, ErrQueryUnsupported)

	default:
		return validationErr
	}
}

// shouldRetryResponse identifies alternate responses that may be transient
// or indicate that this particular query is unsupported by the server.
func shouldRetryResponse(requestType QueryType, responseType ResponseType) bool {
	if requestType == InfoRequest {
		return false
	}

	return responseType == ResponseChallenge ||
		responseType == ResponseInfo ||
		responseType == ResponseInfoGoldSource
}
