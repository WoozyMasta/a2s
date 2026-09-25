// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package transport

import "errors"

// Transaction, framing, validation, and decompression errors are part
// of the public low-level transport contract.
var (
	// ErrChallengeRead indicates that a challenge response could not be read.
	ErrChallengeRead = errors.New("A2S_SERVERQUERY_GETCHALLENGE: failed to read")
	// ErrChallengeValue indicates that a challenge value is invalid or missing.
	ErrChallengeValue = errors.New("A2S_SERVERQUERY_GETCHALLENGE: value read failed")

	// ErrSinglePacket indicates that a single-packet response is too short.
	ErrSinglePacket = errors.New("received single packet data is too short")
	// ErrMultiPacket indicates that a multi-packet response is too short.
	ErrMultiPacket = errors.New("received multi packet data is too short")
	// ErrHeaderWrongRequest indicates that a packet header contains an unsupported request type.
	ErrHeaderWrongRequest = errors.New("unsupported request type in header")
	// ErrWrongRequest indicates that an A2S request type is unsupported.
	ErrWrongRequest = errors.New("unsupported request type")
	// ErrRequestHeader indicates that an A2S request header is invalid.
	ErrRequestHeader = errors.New("invalid A2S request header")
	// ErrRequestPayload indicates that an A2S request payload is invalid.
	ErrRequestPayload = errors.New("invalid A2S request payload")
	// ErrPacketHeader indicates that an A2S packet header is invalid.
	ErrPacketHeader = errors.New("invalid A2S packet header")
	// ErrInsufficientData indicates that a response has insufficient data.
	ErrInsufficientData = errors.New("insufficient data length")
	// ErrMultiPacketMismatch indicates that multi-packet metadata does not match.
	ErrMultiPacketMismatch = errors.New("mismatched number of packets received")
	// ErrMultiPacketSize indicates that a multi-packet response exceeds the size limit.
	ErrMultiPacketSize = errors.New("multi packet response exceeds size limit")
	// ErrMultiPacketConflict indicates that duplicate multi-packet fragments conflict.
	ErrMultiPacketConflict = errors.New("conflicting duplicate multi packet fragment")
	// ErrMultiPacketInconsistent indicates that multi-packet fragment metadata is inconsistent.
	ErrMultiPacketInconsistent = errors.New("inconsistent multi packet fragment metadata")
	// ErrChallengeLoop indicates that a server keeps returning challenge responses.
	ErrChallengeLoop = errors.New("server keeps returning challenge response")
	// ErrQueryUnsupported indicates that the server does not support a requested query.
	ErrQueryUnsupported = errors.New("server does not support requested query")

	// ErrValidatorHeader indicates that a response has an invalid header byte.
	ErrValidatorHeader = errors.New("validator: wrong header byte")
	// ErrValidatorInfo indicates that a response is not a valid A2S_INFO response.
	ErrValidatorInfo = errors.New("validator: wrong A2S_INFO: response")
	// ErrValidatorPlayer indicates that a response is not a valid A2S_PLAYER response.
	ErrValidatorPlayer = errors.New("validator: wrong A2S_PLAYER: response")
	// ErrValidatorRules indicates that a response is not a valid A2S_RULES response.
	ErrValidatorRules = errors.New("validator: wrong A2S_RULES: response")
	// ErrValidatorPing indicates that a response is not a valid A2A_PING response.
	ErrValidatorPing = errors.New("validator: wrong A2A_PING response")
	// ErrValidatorChallenge indicates that a response is not a valid challenge response.
	ErrValidatorChallenge = errors.New("validator: wrong A2S_SERVERQUERY_GETCHALLENGE: response")
	// ErrValidatorRequest indicates that an A2S request type is invalid.
	ErrValidatorRequest = errors.New("validator: wrong request type")

	// ErrDecompressSize indicates that decompressed data exceeds the size limit.
	ErrDecompressSize = errors.New("bz2 decompressed size exceeds limit")
	// ErrDecompressFailed indicates that bzip2 decompression failed.
	ErrDecompressFailed = errors.New("bz2 decompression failed")
	// ErrDecompressSizeMismatch indicates that the decompressed size is unexpected.
	ErrDecompressSizeMismatch = errors.New("bz2 decompressed size mismatch")
	// ErrDecompressCRC indicates that the decompressed data has an invalid CRC32.
	ErrDecompressCRC = errors.New("bz2 CRC32 checksum mismatch")
)
