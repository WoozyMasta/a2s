// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package wire

import (
	"errors"
	"io"
	"testing"
)

func TestDecoderCString(t *testing.T) {
	input := []byte("hello\x00tail")
	decoder := NewDecoder(input)

	value, err := decoder.CStringBytes()
	if err != nil {
		t.Fatalf("CStringBytes() error = %v", err)
	}
	if string(value) != "hello" {
		t.Fatalf("CStringBytes() = %q, want hello", value)
	}
	if got := decoder.Offset(); got != len("hello")+1 {
		t.Fatalf("Offset() = %d, want 6", got)
	}

	valueString, err := decoder.CString()
	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("CString() error = %v, want io.ErrUnexpectedEOF", err)
	}
	if valueString != "" {
		t.Fatalf("CString() = %q, want empty string on error", valueString)
	}
	if got := decoder.Offset(); got != 6 {
		t.Fatalf("Offset() after failed CString() = %d, want 6", got)
	}
}

func TestDecoderCStringEmptyAndFailure(t *testing.T) {
	decoder := NewDecoder([]byte{0, 'x'})

	if got, err := decoder.CString(); err != nil || got != "" {
		t.Fatalf("empty CString() = %q, %v; want empty string, nil", got, err)
	}
	if got := decoder.Offset(); got != 1 {
		t.Fatalf("Offset() after empty CString() = %d, want 1", got)
	}

	if _, err := decoder.CStringBytes(); !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("unterminated CStringBytes() error = %v, want io.ErrUnexpectedEOF", err)
	}
	if got := decoder.Offset(); got != 1 {
		t.Fatalf("Offset() after failed CStringBytes() = %d, want 1", got)
	}
}

func TestDecoderFixedString(t *testing.T) {
	input := []byte("hello")

	for _, test := range []struct {
		name       string
		length     int
		want       string
		wantError  error
		wantOffset int
	}{
		{name: "zero length", length: 0, want: "", wantOffset: 0},
		{name: "exact length", length: len(input), want: "hello", wantOffset: 5},
		{name: "negative length", length: -1, wantError: ErrInvalidLength, wantOffset: 0},
		{name: "one byte too many", length: len(input) + 1, wantError: io.ErrUnexpectedEOF, wantOffset: 0},
	} {
		t.Run(test.name, func(t *testing.T) {
			decoder := NewDecoder(input)
			got, err := decoder.FixedString(test.length)
			if !errors.Is(err, test.wantError) {
				t.Fatalf("FixedString(%d) error = %v, want %v", test.length, err, test.wantError)
			}
			if got != test.want {
				t.Fatalf("FixedString(%d) = %q, want %q", test.length, got, test.want)
			}
			if got := decoder.Offset(); got != test.wantOffset {
				t.Fatalf("Offset() = %d, want %d", got, test.wantOffset)
			}
		})
	}
}
