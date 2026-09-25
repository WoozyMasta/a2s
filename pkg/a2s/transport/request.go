// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package transport

import (
	"context"
	"errors"
)

// request sends one datagram and returns one complete logical response.
func request(
	ctx context.Context,
	io DatagramIO,
	requestType QueryType,
	challenge Challenge,
	readBuf []byte,
) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	req, err := CreateHeader(requestType, challenge)
	if err != nil {
		return nil, err
	}

	if err := io.Write(ctx, req); err != nil {
		return nil, contextError(ctx, err)
	}

	resp, err := readPacket(ctx, io, readBuf)
	if err != nil {
		return resp, err
	}

	multi, err := IsMultiPacket(resp)
	if err != nil {
		return resp, err
	}
	if !multi {
		return bytesClone(resp), nil
	}

	return readSplitResponse(ctx, io, resp, readBuf)
}

// readPacket ignores malformed datagrams
// that can be unrelated traffic on a shared UDP socket,
// but returns a non-retryable framing error immediately.
func readPacket(ctx context.Context, io DatagramIO, readBuf []byte) ([]byte, error) {
	var (
		lastResp []byte
		lastErr  error
	)

	for range maxPacketReads {
		n, err := io.Read(ctx, readBuf)
		if err != nil {
			return nil, contextError(ctx, err)
		}

		resp := readBuf[:n]
		_, packetErr := IsMultiPacket(resp)
		if packetErr == nil {
			// The caller either clones a single packet
			// or copies the first split fragment before reading the next datagram.
			return resp, nil
		}

		lastResp = bytesClone(resp)
		lastErr = packetErr
		if !retryablePacketError(packetErr) {
			break
		}
	}

	if lastErr == nil {
		return nil, ErrSinglePacket
	}

	return lastResp, lastErr
}

// retryablePacketError distinguishes malformed unrelated datagrams
// from errors that indicate a failure in the requested transaction.
func retryablePacketError(err error) bool {
	return errors.Is(err, ErrMultiPacket) ||
		errors.Is(err, ErrSinglePacket) ||
		errors.Is(err, ErrValidatorHeader)
}

// bytesClone keeps packet data independent from the reusable read buffer.
func bytesClone(data []byte) []byte {
	if data == nil {
		return nil
	}

	clone := make([]byte, len(data))
	copy(clone, data)
	return clone
}
