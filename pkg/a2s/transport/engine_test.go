// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package transport

import (
	"context"
	"errors"
	"testing"
)

func TestExecuteRejectsShortAssembledSplitResponse(t *testing.T) {
	io := &scriptedDatagramIO{
		responses: [][]byte{{
			0xFE, 0xFF, 0xFF, 0xFF,
			0x01, 0x00, 0x00, 0x00,
			0x01,
		}},
	}

	_, _, err := Execute(context.Background(), io, InfoRequest, make([]byte, DefaultBufferSize))
	if !errors.Is(err, ErrPacketHeader) {
		t.Fatalf("Execute() error = %v, want ErrPacketHeader", err)
	}
	if !errors.Is(err, ErrInsufficientData) {
		t.Fatalf("Execute() error = %v, want ErrInsufficientData", err)
	}
}

func TestParseChallengeResponseValidatesPacket(t *testing.T) {
	valid := AppendPacket(nil, Packet{
		Type:    ResponseChallenge,
		Payload: []byte{0x78, 0x56, 0x34, 0x12},
	})

	got, err := ParseChallengeResponse(valid)
	if err != nil {
		t.Fatalf("ParseChallengeResponse() error = %v", err)
	}
	if got != (Challenge{0x78, 0x56, 0x34, 0x12}) {
		t.Fatalf("challenge = %X, want 78563412", got)
	}

	tests := []struct {
		name string
		data []byte
		want error
	}{
		{
			name: "wrong response type",
			data: AppendPacket(nil, Packet{
				Type:    ResponseRules,
				Payload: valid[5:],
			}),
			want: ErrValidatorChallenge,
		},
		{
			name: "wrong packet marker",
			data: append([]byte{0x00, 0x00, 0x00, 0x00}, valid[4:]...),
			want: ErrPacketHeader,
		},
		{
			name: "short payload",
			data: AppendPacket(nil, Packet{
				Type:    ResponseChallenge,
				Payload: []byte{0x78, 0x56, 0x34},
			}),
			want: ErrChallengeValue,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := ParseChallengeResponse(test.data); !errors.Is(err, test.want) {
				t.Fatalf("ParseChallengeResponse() error = %v, want %v", err, test.want)
			}
		})
	}
}

// scriptedDatagramIO returns predefined datagrams to the transport engine.
type scriptedDatagramIO struct {
	responses [][]byte
}

// Write accepts a request without modifying the scripted response sequence.
func (io *scriptedDatagramIO) Write(context.Context, []byte) error {
	return nil
}

// Read copies the next scripted datagram into the caller's buffer.
func (io *scriptedDatagramIO) Read(_ context.Context, dst []byte) (int, error) {
	if len(io.responses) == 0 {
		return 0, errors.New("no scripted response")
	}

	response := io.responses[0]
	io.responses = io.responses[1:]
	return copy(dst, response), nil
}
