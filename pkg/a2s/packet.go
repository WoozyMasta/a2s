package a2s

import (
	"bytes"
	"encoding/binary"
	"errors"
)

const packetHeaderSize = 5

// DecodePacket parses one complete logical single-packet A2S response.
//
// Response types are intentionally not validated here.
// This codec is the lossless boundary for proxy and cache code,
// so unknown response types and payload bytes must remain representable.
func DecodePacket(data []byte) (Packet, error) {
	if len(data) < packetHeaderSize {
		return Packet{}, errors.Join(ErrPacketHeader, ErrInsufficientData)
	}

	if binary.LittleEndian.Uint32(data[:4]) != singlePacket {
		return Packet{}, ErrPacketHeader
	}

	return Packet{
		Type:    ResponseType(data[4]),
		Payload: bytes.Clone(data[packetHeaderSize:]),
	}, nil
}

// AppendPacket appends one complete logical single-packet A2S response to dst.
func AppendPacket(dst []byte, packet Packet) ([]byte, error) {
	dst = binary.LittleEndian.AppendUint32(dst, singlePacket)
	dst = append(dst, byte(packet.Type))
	dst = append(dst, packet.Payload...)

	return dst, nil
}
