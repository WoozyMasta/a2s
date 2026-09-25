// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package a2s

import (
	"bytes"
	"hash/crc32"
	"testing"

	"github.com/woozymasta/a2s/pkg/a2s/transport"
)

func FuzzParseChallenge(f *testing.F) {
	f.Add([]byte{0x78, 0x56, 0x34, 0x12})
	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = transport.ParseChallengePayload(data)
	})
}

func FuzzParseChallengeResponse(f *testing.F) {
	f.Add([]byte{0xff, 0xff, 0xff, 0xff, byte(transport.ResponseChallenge), 0x78, 0x56, 0x34, 0x12})
	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = transport.ParseChallengeResponse(data)
	})
}

func FuzzDecodeInfoSource(f *testing.F) {
	f.Add([]byte{17, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = DecodeInfo(transport.Packet{Type: transport.ResponseInfo, Payload: data})
	})
}

func FuzzDecodeInfoGoldSource(f *testing.F) {
	f.Add([]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = DecodeInfo(transport.Packet{Type: transport.ResponseInfoGoldSource, Payload: data})
	})
}

func FuzzParsePlayers(f *testing.F) {
	f.Add([]byte{0})
	f.Add([]byte{1, 0, 'p', 0, 0, 0, 0, 0, 0, 0, 0, 0})
	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = DecodePlayers(transport.Packet{Type: transport.ResponsePlayers, Payload: data})
	})
}

func FuzzParseTheShipPlayers(f *testing.F) {
	f.Add([]byte{0, 0})
	f.Add([]byte{1, 0, 0, 'p', 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = parseTheShipPlayers(data)
	})
}

func FuzzParseRules(f *testing.F) {
	f.Add([]byte{0, 0})
	f.Add([]byte{1, 0, 'm', 'o', 'd', 'e', 0, 'c', 'o', 'o', 'p', 0})
	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = DecodeRules(transport.Packet{Type: transport.ResponseRules, Payload: data})
	})
}

func FuzzParsePacketHeaders(f *testing.F) {
	f.Add([]byte{0xff, 0xff, 0xff, 0xff, byte(transport.ResponsePlayers)})
	f.Add([]byte{0xfe, 0xff, 0xff, 0xff, 1, 0, 0, 0, 1})
	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = transport.IsMultiPacket(data)
		_, _ = transport.ParseSplitHeader(data)
	})
}

func FuzzDecompressBzip2(f *testing.F) {
	f.Add([]byte("not bzip2"))
	f.Fuzz(func(t *testing.T, data []byte) {
		var size uint32
		if len(data) > 0 {
			size = uint32(len(data) % 1024)
		}

		_, _ = transport.DecompressBzip2(data, size, crc32.ChecksumIEEE(data))
	})
}

func FuzzCreateHeader(f *testing.F) {
	f.Add(byte(transport.InfoRequest), []byte{0xff, 0xff, 0xff, 0xff})
	f.Add(byte(transport.RulesRequest), []byte{0x78, 0x56, 0x34, 0x12})
	f.Fuzz(func(t *testing.T, request byte, challengeData []byte) {
		var challenge transport.Challenge
		copy(challenge[:], challengeData)
		_, _ = transport.CreateHeader(transport.QueryType(request), challenge)
	})
}

func FuzzBinaryChallengeRoundTrip(f *testing.F) {
	f.Add([]byte{0x78, 0x56, 0x34, 0x12})
	f.Fuzz(func(t *testing.T, data []byte) {
		got, err := transport.ParseChallengePayload(data)
		if len(data) < len(transport.Challenge{}) {
			if err == nil {
				t.Fatalf("ParseChallengePayload(%X) returned nil error", data)
			}
			return
		}
		if err != nil || !bytes.Equal(got[:], data[:len(transport.Challenge{})]) {
			t.Fatalf("challenge parse = (%X, %v), want %X", got, err, data[:len(transport.Challenge{})])
		}
	})
}
