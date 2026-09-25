// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package transport

import (
	"encoding/binary"
)

// AppendPacket appends one complete logical single-packet A2S response to dst.
func AppendPacket(dst []byte, packet Packet) []byte {
	dst = appendPacketHeader(dst, packet.Type)
	dst = append(dst, packet.Payload...)

	return dst
}

// AppendPacketHeader appends the framing for one logical response.
// The caller may append the payload directly after the returned slice.
func AppendPacketHeader(dst []byte, responseType ResponseType) []byte {
	return appendPacketHeader(dst, responseType)
}

// appendPacketHeader appends the common logical response framing.
func appendPacketHeader(dst []byte, responseType ResponseType) []byte {
	dst = binary.LittleEndian.AppendUint32(dst, SinglePacketMarker)

	return append(dst, byte(responseType))
}
