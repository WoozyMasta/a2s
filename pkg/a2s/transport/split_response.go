// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package transport

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
)

// readSplitResponse collects all fragments of one split response
// and joins their payloads into the logical response expected by the decoders.
func readSplitResponse(
	ctx context.Context,
	io DatagramIO,
	first []byte,
	readBuf []byte,
) ([]byte, error) {
	info, err := ParseSplitHeader(first)
	if err != nil {
		return nil, err
	}
	if len(first) > splitResponseSizeMax {
		return nil, ErrMultiPacketSize
	}

	packets := make([][]byte, info.Count)
	packets[info.Index] = bytesClone(first)
	receivedSize := len(first)

	receivedSize, err = collectSplitPackets(
		ctx,
		io,
		readBuf,
		info,
		packets,
		receivedSize,
	)
	if err != nil {
		return nil, err
	}

	info, totalSize, err := validateSplitPackets(info, packets)
	if err != nil {
		return nil, err
	}

	// The collection limit protects the socket loop;
	// this check protects the final allocation after all fragments have been validated.
	if receivedSize > splitResponseSizeMax {
		return nil, ErrMultiPacketSize
	}

	return assembleSplitResponse(info, packets, totalSize)
}

// collectSplitPackets reads fragments until every expected index is present.
func collectSplitPackets(
	ctx context.Context,
	io DatagramIO,
	readBuf []byte,
	info SplitHeader,
	packets [][]byte,
	receivedSize int,
) (int, error) {
	received := 1
	for received < info.Count {
		n, err := io.Read(ctx, readBuf)
		if err != nil {
			return receivedSize, contextError(ctx, err)
		}

		resp := readBuf[:n]
		if n < splitMin {
			// Short datagrams cannot contain a split fragment header.
			continue
		}
		if binary.LittleEndian.Uint32(resp[4:8]) != info.ID {
			// A connected UDP socket can still receive
			// stale fragments from a previous transaction.
			// Ignore a different split identifier.
			continue
		}

		packetInfo, err := ParseSplitHeader(resp)
		if err != nil {
			return receivedSize, errors.Join(ErrMultiPacketInconsistent, err)
		}
		if err := ValidateSplitFragment(info, packetInfo); err != nil {
			return receivedSize, err
		}
		if packetInfo.Index < 0 || packetInfo.Index >= info.Count {
			// The metadata is valid enough to identify the response,
			// but the fragment index cannot contribute to this response.
			continue
		}

		current := packetInfo.Index
		if packets[current] == nil {
			if n > splitResponseSizeMax-receivedSize {
				return receivedSize, ErrMultiPacketSize
			}

			packets[current] = bytesClone(resp)
			receivedSize += n
			received++
			continue
		}

		// Repeated identical fragments are harmless;
		// conflicting duplicates indicate that the response
		// cannot be assembled deterministically.
		if !bytes.Equal(packets[current], resp) {
			return receivedSize, ErrMultiPacketConflict
		}
	}

	return receivedSize, nil
}

// validateSplitPackets verifies ordering, metadata, and payload sizes
// before assembling a response or allocating its final backing array.
func validateSplitPackets(
	info SplitHeader,
	packets [][]byte,
) (SplitHeader, int, error) {
	firstInfo, err := ParseSplitHeader(packets[0])
	if err != nil || firstInfo.Index != 0 {
		return SplitHeader{}, 0, ErrMultiPacket
	}
	info = firstInfo

	totalSize := 0
	for index, packet := range packets {
		if packet == nil {
			return SplitHeader{}, 0, ErrMultiPacketMismatch
		}

		packetInfo, err := ParseSplitHeader(packet)
		if err != nil {
			return SplitHeader{}, 0, errors.Join(ErrMultiPacketInconsistent, err)
		}
		if err := ValidateSplitFragment(info, packetInfo); err != nil || packetInfo.Index != index {
			return SplitHeader{}, 0, ErrMultiPacketInconsistent
		}

		dataOffset := info.HeaderSize
		if index == 0 {
			dataOffset = info.DataOffset
		}
		if len(packet) < dataOffset {
			return SplitHeader{}, 0, ErrMultiPacket
		}

		packetSize := len(packet) - dataOffset
		if packetSize > splitResponseSizeMax-totalSize {
			return SplitHeader{}, 0, ErrMultiPacketSize
		}
		totalSize += packetSize
	}

	return info, totalSize, nil
}

// assembleSplitResponse concatenates validated fragments
// and decompresses a Source split response
// when its compression flag is set.
func assembleSplitResponse(
	info SplitHeader,
	packets [][]byte,
	totalSize int,
) ([]byte, error) {
	assembled := make([]byte, 0, totalSize)
	for index, packet := range packets {
		dataOffset := info.HeaderSize
		if index == 0 {
			dataOffset = info.DataOffset
		}
		assembled = append(assembled, packet[dataOffset:]...)
	}

	if !info.Compressed {
		return assembled, nil
	}

	return DecompressBzip2(assembled, info.UnpackedSize, info.CRC)
}
