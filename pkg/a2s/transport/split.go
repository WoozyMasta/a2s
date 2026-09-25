// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package transport

import (
	"bytes"
	"compress/bzip2"
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
)

// SplitHeader contains normalized metadata about one split packet.
type SplitHeader struct {
	// Count is the total number of fragments.
	Count int
	// Index is the zero-based fragment index.
	Index int
	// HeaderSize is the common header size for subsequent fragments.
	HeaderSize int
	// DataOffset is the payload offset for the first fragment.
	DataOffset int
	// ID identifies fragments belonging to one response.
	ID uint32
	// UnpackedSize is the declared size of compressed payload data.
	UnpackedSize uint32
	// CRC is the declared checksum of decompressed payload data.
	CRC uint32
	// Compressed reports whether the response uses Source bzip2 framing.
	Compressed bool
	// GoldSource reports whether the GoldSource split layout was detected.
	GoldSource bool
}

// IsMultiPacket classifies the common A2S packet marker
// and validates the minimum header length needed by the selected framing.
func IsMultiPacket(data []byte) (bool, error) {
	if len(data) < 4 {
		return true, ErrMultiPacket
	}

	header := binary.LittleEndian.Uint32(data[:4])
	switch header {
	case SinglePacketMarker:
		if len(data) < 5 {
			return false, ErrSinglePacket
		}
		return false, nil

	case SplitPacketMarker:
		if len(data) < 9 {
			return true, ErrMultiPacket
		}
		return true, nil

	default:
		return false, errors.Join(ErrValidatorHeader, fmt.Errorf("0x%X", header))
	}
}

// ParseSplitHeader recognizes GoldSource and Source split layouts.
// Source packets are identified from their embedded single-packet marker
// or a plausible split-size field when the marker is not available yet.
func ParseSplitHeader(data []byte) (SplitHeader, error) {
	if len(data) < splitMin {
		return SplitHeader{}, ErrMultiPacket
	}

	packetID := binary.LittleEndian.Uint32(data[4:8])
	countSrc, numberSrc := 0, 0
	if len(data) >= 10 {
		countSrc = int(data[8])
		numberSrc = int(data[9])
	}

	countGold := int(data[8] & 0x0F)
	numberGold := int((data[8] & 0xF0) >> 4)

	// GoldSource and Source put the fragment count/index at different offsets.
	// Probe the markers first, then use the Source size field
	// as a fallback for fragments whose payload does not reach an embedded marker.
	isSource := false
	switch {
	case len(data) >= 13 && binary.LittleEndian.Uint32(data[9:13]) == SinglePacketMarker:

	case len(data) >= 16 && binary.LittleEndian.Uint32(data[12:16]) == SinglePacketMarker:
		isSource = true

	case len(data) >= srcSplitHeader:
		splitSize := binary.LittleEndian.Uint16(data[splitSizeOff:srcSplitHeader])
		isSource = splitSize > 0 && splitSize <= splitSizeMax
	}

	packetCount := countGold
	currentPacket := numberGold
	baseHeaderSize := goldSrcHeader
	useGold := !isSource
	if isSource {
		packetCount = countSrc
		currentPacket = numberSrc
		baseHeaderSize = srcSplitHeader
	}

	info := SplitHeader{
		ID:         packetID,
		Count:      packetCount,
		Index:      currentPacket,
		HeaderSize: baseHeaderSize,
		DataOffset: baseHeaderSize,
		GoldSource: useGold,
	}

	if !useGold && packetID&0x80000000 != 0 {
		info.Compressed = true
		if info.Index == 0 {
			if len(data) < baseHeaderSize+8 {
				return SplitHeader{}, ErrMultiPacket
			}

			info.UnpackedSize = binary.LittleEndian.Uint32(data[baseHeaderSize : baseHeaderSize+4])
			info.CRC = binary.LittleEndian.Uint32(data[baseHeaderSize+4 : baseHeaderSize+8])
			info.DataOffset = baseHeaderSize + 8

			// Some Source servers omit the ordinary split-size word
			// and place compression metadata at the legacy offset instead.
			if baseHeaderSize == srcSplitHeader &&
				info.UnpackedSize > unpackProbeMax && len(data) >= 18 {
				altSize := binary.LittleEndian.Uint32(data[splitSizeOff : splitSizeOff+4])
				altCRC := binary.LittleEndian.Uint32(data[splitSizeOff+4 : splitSizeOff+8])
				if altSize > 0 && altSize <= unpackProbeMax {
					info.UnpackedSize = altSize
					info.CRC = altCRC
					info.HeaderSize = srcSplitNoSize
					info.DataOffset = splitSizeOff + 8
				}
			}
		}
	}

	if info.Count <= 0 || info.Count > splitPacketCountMax ||
		info.Index < 0 || info.Index >= info.Count {
		return SplitHeader{}, ErrMultiPacket
	}

	return info, nil
}

// ValidateSplitFragment checks metadata that must remain identical
// across all fragments belonging to one logical response.
func ValidateSplitFragment(expected, actual SplitHeader) error {
	if actual.ID != expected.ID ||
		actual.Count != expected.Count ||
		actual.Compressed != expected.Compressed ||
		actual.GoldSource != expected.GoldSource ||
		actual.Index < 0 || actual.Index >= expected.Count {
		return ErrMultiPacketInconsistent
	}

	return nil
}

// DecompressBzip2 validates the declared size and checksum
// before returning a decompressed logical response.
func DecompressBzip2(compressed []byte, size uint32, crc uint32) ([]byte, error) {
	if size > maxDecompressedSize {
		return nil, fmt.Errorf("%w: %d > %d", ErrDecompressSize, size, maxDecompressedSize)
	}

	reader := bzip2.NewReader(bytes.NewReader(compressed))
	decompressed := make([]byte, int(size))

	readBytes, err := io.ReadFull(reader, decompressed)
	if err != nil && err != io.ErrUnexpectedEOF {
		return nil, ErrDecompressFailed
	}
	if readBytes != int(size) {
		return nil, ErrDecompressSizeMismatch
	}
	if crc32.ChecksumIEEE(decompressed) != crc {
		return nil, ErrDecompressCRC
	}

	return decompressed, nil
}
