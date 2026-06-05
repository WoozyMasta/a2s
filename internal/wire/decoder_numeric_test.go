package wire

import (
	"errors"
	"io"
	"math"
	"testing"
)

func TestDecoderNumeric(t *testing.T) {
	decoder := NewDecoder([]byte{
		0x34, 0x12,
		0x78, 0x56, 0x34, 0x12,
		0xef, 0xcd, 0xab, 0x89, 0x67, 0x45, 0x23, 0x01,
		0x00, 0x00, 0x00, 0x80,
		0x00, 0x00, 0x20, 0xc0,
	})

	if got, err := decoder.Uint16(); err != nil || got != 0x1234 {
		t.Fatalf("Uint16() = %#x, %v; want %#x, nil", got, err, uint16(0x1234))
	}
	if got, err := decoder.Uint32(); err != nil || got != 0x12345678 {
		t.Fatalf("Uint32() = %#x, %v; want %#x, nil", got, err, uint32(0x12345678))
	}
	if got, err := decoder.Uint64(); err != nil || got != 0x0123456789abcdef {
		t.Fatalf("Uint64() = %#x, %v; want %#x, nil", got, err, uint64(0x0123456789abcdef))
	}
	if got, err := decoder.Int32(); err != nil || got != math.MinInt32 {
		t.Fatalf("Int32() = %d, %v; want %d, nil", got, err, int32(math.MinInt32))
	}
	if got, err := decoder.Float32(); err != nil || got != -2.5 {
		t.Fatalf("Float32() = %f, %v; want -2.5, nil", got, err)
	}
	if got := decoder.Offset(); got != 22 {
		t.Fatalf("Offset() = %d, want 22", got)
	}
}

func TestDecoderFloat32BitPatterns(t *testing.T) {
	for _, test := range []struct {
		name string
		bits uint32
	}{
		{name: "zero", bits: 0x00000000},
		{name: "negative zero", bits: 0x80000000},
		{name: "infinity", bits: 0x7f800000},
		{name: "NaN", bits: 0x7fc00001},
	} {
		t.Run(test.name, func(t *testing.T) {
			data := []byte{byte(test.bits), byte(test.bits >> 8), byte(test.bits >> 16), byte(test.bits >> 24)}
			decoder := NewDecoder(data)
			got, err := decoder.Float32()
			if err != nil {
				t.Fatalf("Float32() error = %v", err)
			}
			if gotBits := math.Float32bits(got); gotBits != test.bits {
				t.Fatalf("Float32() bits = %#x, want %#x", gotBits, test.bits)
			}
		})
	}
}

func TestDecoderNumericFailuresPreserveOffset(t *testing.T) {
	for _, test := range []struct {
		name string
		read func(*Decoder) error
	}{
		{name: "Uint16", read: func(decoder *Decoder) error { _, err := decoder.Uint16(); return err }},
		{name: "Uint32", read: func(decoder *Decoder) error { _, err := decoder.Uint32(); return err }},
		{name: "Uint64", read: func(decoder *Decoder) error { _, err := decoder.Uint64(); return err }},
		{name: "Int32", read: func(decoder *Decoder) error { _, err := decoder.Int32(); return err }},
		{name: "Float32", read: func(decoder *Decoder) error { _, err := decoder.Float32(); return err }},
	} {
		t.Run(test.name, func(t *testing.T) {
			decoder := NewDecoder([]byte{1})
			if err := test.read(&decoder); !errors.Is(err, io.ErrUnexpectedEOF) {
				t.Fatalf("read error = %v, want io.ErrUnexpectedEOF", err)
			}
			if got := decoder.Offset(); got != 0 {
				t.Fatalf("Offset() after failed read = %d, want 0", got)
			}
		})
	}
}
