package wire

import (
	"errors"
	"io"
	"testing"
)

func TestDecoderByte(t *testing.T) {
	decoder := NewDecoder([]byte{0x12})

	value, err := decoder.Byte()
	if err != nil {
		t.Fatalf("Byte() error = %v", err)
	}
	if value != 0x12 {
		t.Fatalf("Byte() = %#x, want %#x", value, 0x12)
	}
	if got := decoder.Offset(); got != 1 {
		t.Fatalf("Offset() = %d, want 1", got)
	}

	if _, err := decoder.Byte(); !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("second Byte() error = %v, want io.ErrUnexpectedEOF", err)
	}
	if got := decoder.Offset(); got != 1 {
		t.Fatalf("Offset() after failed Byte() = %d, want 1", got)
	}
}

func TestDecoderBytes(t *testing.T) {
	input := []byte{1, 2, 3}

	tests := []struct {
		name       string
		length     int
		want       []byte
		wantError  error
		wantOffset int
	}{
		{
			name:       "zero length",
			length:     0,
			want:       input[:0],
			wantOffset: 0,
		},
		{
			name:       "exact remaining length",
			length:     len(input),
			want:       input,
			wantOffset: len(input),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			decoder := NewDecoder(input)
			got, err := decoder.Bytes(test.length)
			if !errors.Is(err, test.wantError) {
				t.Fatalf("Bytes(%d) error = %v, want %v", test.length, err, test.wantError)
			}
			if string(got) != string(test.want) {
				t.Fatalf("Bytes(%d) = %v, want %v", test.length, got, test.want)
			}
			if got := decoder.Offset(); got != test.wantOffset {
				t.Fatalf("Offset() = %d, want %d", got, test.wantOffset)
			}
		})
	}
}

func TestDecoderBytesFailuresPreserveOffset(t *testing.T) {
	for _, test := range []struct {
		name   string
		length int
		want   error
	}{
		{name: "negative length", length: -1, want: ErrInvalidLength},
		{name: "one byte too many", length: 4, want: io.ErrUnexpectedEOF},
	} {
		t.Run(test.name, func(t *testing.T) {
			decoder := NewDecoder([]byte{1, 2, 3})
			if _, err := decoder.Bytes(test.length); !errors.Is(err, test.want) {
				t.Fatalf("Bytes(%d) error = %v, want %v", test.length, err, test.want)
			}
			if got := decoder.Offset(); got != 0 {
				t.Fatalf("Offset() after failed Bytes() = %d, want 0", got)
			}
			if got := decoder.Remaining(); got != 3 {
				t.Fatalf("Remaining() after failed Bytes() = %d, want 3", got)
			}
		})
	}
}

func TestDecoderTail(t *testing.T) {
	input := []byte{1, 2, 3}
	decoder := NewDecoder(input)

	if _, err := decoder.Bytes(1); err != nil {
		t.Fatalf("Bytes(1) error = %v", err)
	}

	tail := decoder.Tail()
	if string(tail) != string(input[1:]) {
		t.Fatalf("Tail() = %v, want %v", tail, input[1:])
	}
	if got := decoder.Offset(); got != 1 {
		t.Fatalf("Offset() after Tail() = %d, want 1", got)
	}
	if got := decoder.Remaining(); got != 2 {
		t.Fatalf("Remaining() = %d, want 2", got)
	}
	if decoder.Empty() {
		t.Fatal("Empty() = true with unread bytes")
	}

	if _, err := decoder.Bytes(2); err != nil {
		t.Fatalf("Bytes(2) error = %v", err)
	}
	if !decoder.Empty() {
		t.Fatal("Empty() = false after consuming all bytes")
	}
	if got := decoder.Tail(); len(got) != 0 {
		t.Fatalf("Tail() after consuming input = %v, want empty", got)
	}
}
