package server

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"sync/atomic"
)

const (
	// DefaultSourceSplitSize is the default maximum Source split datagram size.
	DefaultSourceSplitSize = 1248
	// DefaultSourceMaxResponseSize bounds one logical response before packetization.
	DefaultSourceMaxResponseSize = 16 * 1024 * 1024

	// sourceSplitHeaderSize is the uncompressed Source split header size.
	sourceSplitHeaderSize = 12
	// sourceSplitCountMax is the maximum fragment count representable by one byte.
	sourceSplitCountMax = 255
	// sourceSplitSizeMax is the largest value representable by the wire uint16 size field.
	sourceSplitSizeMax = 1<<16 - 1
	// sourceSplitMarker identifies a Source multi-packet datagram.
	sourceSplitMarker = uint32(0xFFFFFFFE)
	// sourceSplitReservedID is the compression flag and must be clear for uncompressed packets.
	sourceSplitReservedID = uint32(0x80000000)
)

// defaultSourceSplitSequence provides IDs for zero-value packetizers.
var defaultSourceSplitSequence atomic.Uint32

// SourcePacketizer wraps one logical A2S response in Source single or split datagrams.
type SourcePacketizer struct {

	// nextID returns the next split ID for packetizer instances created by the constructor.
	nextID func() uint32
	// SplitSize is the maximum size of each Source datagram, including its split header.
	// Zero uses DefaultSourceSplitSize.
	SplitSize int

	// MaxResponseSize is the maximum size of one logical response.
	// Zero uses DefaultSourceMaxResponseSize.
	MaxResponseSize int
}

// NewSourcePacketizer creates a packetizer with a randomized split ID sequence.
func NewSourcePacketizer() (*SourcePacketizer, error) {
	var seedBytes [4]byte
	if _, err := rand.Read(seedBytes[:]); err != nil {
		return nil, fmt.Errorf("%w: initialize split ID generator: %w", ErrPacketizer, err)
	}

	var sequence atomic.Uint32
	sequence.Store(binary.LittleEndian.Uint32(seedBytes[:]) &^ sourceSplitReservedID)

	return &SourcePacketizer{
		SplitSize:       DefaultSourceSplitSize,
		MaxResponseSize: DefaultSourceMaxResponseSize,
		nextID: func() uint32 {
			return sequence.Add(1) &^ sourceSplitReservedID
		},
	}, nil
}

// Packetize converts one complete logical A2S response into UDP datagrams.
// The input must include the logical packet marker and response type.
func (p *SourcePacketizer) Packetize(data []byte) ([][]byte, error) {
	if p == nil {
		return nil, fmt.Errorf("%w: packetizer is nil", ErrPacketizer)
	}

	splitSize := p.SplitSize
	if splitSize == 0 {
		splitSize = DefaultSourceSplitSize
	}

	if splitSize <= sourceSplitHeaderSize || splitSize > sourceSplitSizeMax {
		return nil, fmt.Errorf(
			"%w: got %d, want %d..%d",
			ErrPacketizerSplitSize,
			splitSize,
			sourceSplitHeaderSize+1,
			sourceSplitSizeMax,
		)
	}

	maxResponseSize := p.MaxResponseSize
	if maxResponseSize == 0 {
		maxResponseSize = DefaultSourceMaxResponseSize
	}
	if maxResponseSize <= 0 {
		return nil, fmt.Errorf("%w: got %d", ErrPacketizerResponseSize, maxResponseSize)
	}
	if len(data) < 5 || binary.LittleEndian.Uint32(data[:4]) != ^uint32(0) {
		return nil, fmt.Errorf("%w: logical single-packet framing is invalid", ErrPacketizerInput)
	}
	if len(data) > maxResponseSize {
		return nil, fmt.Errorf(
			"%w: got %d, maximum %d",
			ErrPacketizerResponseSize,
			len(data),
			maxResponseSize,
		)
	}

	if len(data) <= splitSize {
		return [][]byte{bytes.Clone(data)}, nil
	}

	chunkSize := splitSize - sourceSplitHeaderSize
	count := (len(data) + chunkSize - 1) / chunkSize
	if count > sourceSplitCountMax {
		return nil, fmt.Errorf(
			"%w: got %d, maximum %d",
			ErrPacketizerFragmentCount,
			count,
			sourceSplitCountMax,
		)
	}

	id := nextSourceSplitID(p.nextID)
	packets := make([][]byte, 0, count)
	for index, start := 0, 0; start < len(data); index, start = index+1, start+chunkSize {
		end := start + chunkSize
		if end > len(data) {
			end = len(data)
		}

		packet := make([]byte, sourceSplitHeaderSize+end-start)
		binary.LittleEndian.PutUint32(packet[:4], sourceSplitMarker)
		binary.LittleEndian.PutUint32(packet[4:8], id)
		// #nosec G115 -- count is validated to fit uint8 above.
		packet[8] = byte(count)
		// #nosec G115 -- index is bounded by the validated fragment count.
		packet[9] = byte(index)
		// #nosec G115 -- splitSize is validated to fit uint16 above.
		binary.LittleEndian.PutUint16(packet[10:12], uint16(splitSize))
		copy(packet[sourceSplitHeaderSize:], data[start:end])
		packets = append(packets, packet)
	}

	return packets, nil
}

// nextSourceSplitID returns a non-compression Source split ID.
func nextSourceSplitID(nextID func() uint32) uint32 {
	if nextID != nil {
		return nextID() &^ sourceSplitReservedID
	}

	return defaultSourceSplitSequence.Add(1) &^ sourceSplitReservedID
}
