// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package a2s

import (
	"bytes"
	"errors"
	"testing"

	"github.com/woozymasta/a2s/pkg/a2s/transport"
)

func TestPacketCodecRoundTrip(t *testing.T) {
	tests := []struct {
		name   string
		packet transport.Packet
	}{
		{name: "empty payload", packet: transport.Packet{Type: transport.ResponsePing}},
		{name: "known payload", packet: transport.Packet{
			Type:    transport.ResponseRules,
			Payload: []byte{0x00, 0xFF, 0x7F},
		}},
		{name: "unknown response", packet: transport.Packet{
			Type:    transport.ResponseType(0xFE),
			Payload: []byte("opaque response payload"),
		}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			encoded := transport.AppendPacket(nil, test.packet)

			decoded, err := transport.DecodePacket(encoded)
			if err != nil {
				t.Fatalf("transport.DecodePacket() error = %v", err)
			}
			if decoded.Type != test.packet.Type || !bytes.Equal(decoded.Payload, test.packet.Payload) {
				t.Fatalf("decoded packet = %+v, want %+v", decoded, test.packet)
			}
		})
	}
}

func TestDecodePacketFixtures(t *testing.T) {
	data := readProtocolFixture(t, "response_challenge.hex")
	got, err := transport.DecodePacket(data)
	if err != nil {
		t.Fatalf("transport.DecodePacket() error = %v", err)
	}

	want := transport.Packet{
		Type:    transport.ResponseChallenge,
		Payload: []byte{0x78, 0x56, 0x34, 0x12},
	}
	if got.Type != want.Type || !bytes.Equal(got.Payload, want.Payload) {
		t.Fatalf("decoded packet = %+v, want %+v", got, want)
	}
}

func TestDecodePacketOwnsPayload(t *testing.T) {
	data := singlePacketFixture(transport.ResponseRules, []byte("payload"))
	got, err := transport.DecodePacket(data)
	if err != nil {
		t.Fatalf("transport.DecodePacket() error = %v", err)
	}

	data[len(data)-1] = 'X'
	if string(got.Payload) != "payload" {
		t.Fatalf("decoded payload changed with input buffer: %q", got.Payload)
	}
}

func TestDecodePacketRejectsNonSingleFraming(t *testing.T) {
	valid := singlePacketFixture(transport.ResponseRules, []byte("payload"))
	tests := []struct {
		name string
		data []byte
		want error
	}{
		{name: "short header", data: valid[:4], want: transport.ErrPacketHeader},
		{name: "wrong marker", data: append([]byte{0xFE, 0xFF, 0xFF, 0xFF}, valid[4:]...), want: transport.ErrPacketHeader},
		{name: "split response", data: sourceSplitPacketSequence(1, valid, 3)[0], want: transport.ErrPacketHeader},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := transport.DecodePacket(test.data); !errors.Is(err, test.want) {
				t.Fatalf("transport.DecodePacket() error = %v, want %v", err, test.want)
			}
		})
	}
}

func TestAppendPacketPreservesPrefix(t *testing.T) {
	prefix := []byte{0xAA, 0xBB}
	got := transport.AppendPacket(prefix, transport.Packet{Type: transport.ResponseInfo, Payload: []byte("info")})

	want := append([]byte{0xAA, 0xBB}, singlePacketFixture(transport.ResponseInfo, []byte("info"))...)
	if !bytes.Equal(got, want) {
		t.Fatalf("encoded packet = %X, want %X", got, want)
	}
}
