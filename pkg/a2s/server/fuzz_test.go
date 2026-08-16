// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package server

import (
	"context"
	"encoding/binary"
	"net/netip"
	"testing"
	"time"

	"github.com/woozymasta/a2s/pkg/a2s"
)

const (
	fuzzInputMax   = 64 * 1024
	fuzzStringMax  = 256
	fuzzEntriesMax = 64
)

func FuzzDecodeRequest(f *testing.F) {
	f.Add([]byte{0xFF, 0xFF, 0xFF, 0xFF, byte(a2s.PingRequest)})
	f.Add([]byte{0xFF, 0xFF, 0xFF, 0xFF, byte(a2s.InfoRequest), 'S', 'o', 'u', 'r', 'c', 'e'})
	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = a2s.DecodeRequest(limitFuzzBytes(data, fuzzInputMax))
	})
}

func FuzzDecodePacket(f *testing.F) {
	f.Add([]byte{0xFF, 0xFF, 0xFF, 0xFF, byte(a2s.ResponseRules)})
	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = a2s.DecodePacket(limitFuzzBytes(data, fuzzInputMax))
	})
}

func FuzzEncodeInfo(f *testing.F) {
	f.Add([]byte("server name"))
	f.Fuzz(func(t *testing.T, data []byte) {
		value := fuzzString(data)
		gameID := uint64(len(data))
		_, _ = a2s.AppendInfo(nil, a2s.Info{
			Name:         value,
			Map:          value,
			Folder:       value,
			Game:         value,
			Version:      value,
			SourceTVName: value,
			Keywords:     []string{value},
			GameID:       &gameID,
		})
	})
}

func FuzzEncodePlayers(f *testing.F) {
	f.Add([]byte("player"))
	f.Fuzz(func(t *testing.T, data []byte) {
		value := fuzzString(data)
		players := make([]a2s.Player, minInt(len(data), fuzzEntriesMax))
		for index := range players {
			players[index] = a2s.Player{
				Name:     value,
				Index:    byte(index),
				Score:    int32(index),
				Duration: time.Duration(index) * time.Second,
			}
		}

		_, _ = a2s.AppendPlayers(nil, players)
	})
}

func FuzzEncodeRules(f *testing.F) {
	f.Add([]byte("hostname"))
	f.Fuzz(func(t *testing.T, data []byte) {
		value := fuzzString(data)
		rules := make(a2s.Rules, minInt(len(data), fuzzEntriesMax))
		for index := range rules {
			rules[index] = a2s.Rule{
				Name:  value,
				Value: value,
			}
		}

		_, _ = a2s.AppendRules(nil, rules)
	})
}

func FuzzPacketizers(f *testing.F) {
	f.Add([]byte("packetizer payload"), uint8(32))
	f.Add([]byte("this payload exercises split packetization"), uint8(8))
	f.Fuzz(func(t *testing.T, data []byte, splitSize uint8) {
		payload := limitFuzzBytes(data, fuzzInputMax)
		logical := make([]byte, 5+len(payload))
		binary.LittleEndian.PutUint32(logical[:4], 0xFFFFFFFF)
		logical[4] = byte(a2s.ResponseRules)
		copy(logical[5:], payload)

		source := &SourcePacketizer{
			SplitSize:       sourceSplitHeaderSize + 1 + int(splitSize),
			MaxResponseSize: fuzzInputMax + 5,
			nextID:          func() uint32 { return 1 },
		}
		goldSource := &GoldSourcePacketizer{
			SplitSize:       goldSourceSplitHeaderSize + 1 + int(splitSize),
			MaxResponseSize: fuzzInputMax + 5,
			nextID:          func() uint32 { return 0xD9D51BBC },
		}

		_, _ = source.Packetize(logical)
		_, _ = goldSource.Packetize(logical)
	})
}

func FuzzChallengeGate(f *testing.F) {
	provider, err := NewStatelessChallengeProvider()
	if err != nil {
		f.Fatalf("NewStatelessChallengeProvider() error = %v", err)
	}

	gate, err := NewChallengeGate(
		HandlerFunc(func(context.Context, *Request) (Response, error) {
			return PacketResponse{Packet: a2s.Packet{Type: a2s.ResponseInfo}}, nil
		}),
		SecureChallengePolicy(),
		provider,
	)
	if err != nil {
		f.Fatalf("NewChallengeGate() error = %v", err)
	}

	f.Add([]byte{0x78, 0x56, 0x34, 0x12})
	f.Fuzz(func(t *testing.T, data []byte) {
		var challenge a2s.Challenge
		copy(challenge[:], data)
		remote := netip.AddrPortFrom(netip.AddrFrom4([4]byte{127, 0, 0, 1}), uint16(len(data)))
		request := &Request{
			Remote: remote,
			Query: a2s.Request{
				Type:         a2s.QueryType(byteAt(data, 0)),
				Challenge:    challenge,
				HasChallenge: len(data)%2 == 0,
			},
		}

		_, _ = gate.Handle(context.Background(), request)
	})
}

func fuzzString(data []byte) string {
	return string(limitFuzzBytes(data, fuzzStringMax))
}

func limitFuzzBytes(data []byte, max int) []byte {
	if len(data) > max {
		return data[:max]
	}

	return data
}

func minInt(left, right int) int {
	if left < right {
		return left
	}

	return right
}

func byteAt(data []byte, index int) byte {
	if index < 0 || index >= len(data) {
		return 0
	}

	return data[index]
}
