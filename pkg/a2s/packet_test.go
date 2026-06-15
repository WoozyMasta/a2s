package a2s

import (
	"bytes"
	"errors"
	"testing"
)

func TestPacketCodecRoundTrip(t *testing.T) {
	tests := []struct {
		name   string
		packet Packet
	}{
		{name: "empty payload", packet: Packet{Type: ResponsePing}},
		{name: "known payload", packet: Packet{
			Type:    ResponseRules,
			Payload: []byte{0x00, 0xFF, 0x7F},
		}},
		{name: "unknown response", packet: Packet{
			Type:    ResponseType(0xFE),
			Payload: []byte("opaque response payload"),
		}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			encoded, err := AppendPacket(nil, test.packet)
			if err != nil {
				t.Fatalf("AppendPacket() error = %v", err)
			}

			decoded, err := DecodePacket(encoded)
			if err != nil {
				t.Fatalf("DecodePacket() error = %v", err)
			}
			if decoded.Type != test.packet.Type || !bytes.Equal(decoded.Payload, test.packet.Payload) {
				t.Fatalf("decoded packet = %+v, want %+v", decoded, test.packet)
			}
		})
	}
}

func TestDecodePacketFixtures(t *testing.T) {
	data := readProtocolFixture(t, "response_challenge.hex")
	got, err := DecodePacket(data)
	if err != nil {
		t.Fatalf("DecodePacket() error = %v", err)
	}

	want := Packet{
		Type:    ResponseChallenge,
		Payload: []byte{0x78, 0x56, 0x34, 0x12},
	}
	if got.Type != want.Type || !bytes.Equal(got.Payload, want.Payload) {
		t.Fatalf("decoded packet = %+v, want %+v", got, want)
	}
}

func TestDecodePacketOwnsPayload(t *testing.T) {
	data := singlePacketFixture(ResponseRules, []byte("payload"))
	got, err := DecodePacket(data)
	if err != nil {
		t.Fatalf("DecodePacket() error = %v", err)
	}

	data[len(data)-1] = 'X'
	if string(got.Payload) != "payload" {
		t.Fatalf("decoded payload changed with input buffer: %q", got.Payload)
	}
}

func TestDecodePacketRejectsNonSingleFraming(t *testing.T) {
	valid := singlePacketFixture(ResponseRules, []byte("payload"))
	tests := []struct {
		name string
		data []byte
		want error
	}{
		{name: "short header", data: valid[:4], want: ErrPacketHeader},
		{name: "wrong marker", data: append([]byte{0xFE, 0xFF, 0xFF, 0xFF}, valid[4:]...), want: ErrPacketHeader},
		{name: "split response", data: sourceSplitPacketSequence(1, valid, 3)[0], want: ErrPacketHeader},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := DecodePacket(test.data); !errors.Is(err, test.want) {
				t.Fatalf("DecodePacket() error = %v, want %v", err, test.want)
			}
		})
	}
}

func TestAppendPacketPreservesPrefix(t *testing.T) {
	prefix := []byte{0xAA, 0xBB}
	got, err := AppendPacket(prefix, Packet{Type: ResponseInfo, Payload: []byte("info")})
	if err != nil {
		t.Fatalf("AppendPacket() error = %v", err)
	}

	want := append(prefix, singlePacketFixture(ResponseInfo, []byte("info"))...)
	if !bytes.Equal(got, want) {
		t.Fatalf("encoded packet = %X, want %X", got, want)
	}
}
