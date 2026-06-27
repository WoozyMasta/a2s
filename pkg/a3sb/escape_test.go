package a3sb

import (
	"bytes"
	"testing"
)

func TestAppendDecodedEscapeSequences(t *testing.T) {
	for _, test := range []struct {
		name string
		data []byte
		want []byte
	}{
		{name: "01 01", data: []byte{0x01, 0x01}, want: []byte{0x01}},
		{name: "01 02", data: []byte{0x01, 0x02}, want: []byte{0x00}},
		{name: "01 03", data: []byte{0x01, 0x03}, want: []byte{0xFF}},
		{name: "ordinary bytes", data: []byte{0x00, 0x02, 0xFF}, want: []byte{0x00, 0x02, 0xFF}},
		{name: "trailing 01", data: []byte{0x01}, want: []byte{0x01}},
		{name: "unknown escape", data: []byte{0x01, 0x04}, want: []byte{0x01, 0x04}},
	} {
		t.Run(test.name, func(t *testing.T) {
			got := appendDecodedEscapeSequences([]byte{0xAA}, test.data)
			want := append([]byte{0xAA}, test.want...)
			if !bytes.Equal(got, want) {
				t.Fatalf("appendDecodedEscapeSequences() = %X, want %X", got, want)
			}
		})
	}
}

func TestAppendDecodedEscapeSequencesInPlace(t *testing.T) {
	data := []byte{0xAA, 0x01, 0x01, 0x01, 0x02, 0x01, 0x03, 0x01, 0x04, 0x01}

	got := appendDecodedEscapeSequences(nil, data)
	want := []byte{0xAA, 0x01, 0x00, 0xFF, 0x01, 0x04, 0x01}
	if !bytes.Equal(got, want) {
		t.Fatalf("appendDecodedEscapeSequences(nil, data) = %X, want %X", got, want)
	}
}

func TestAppendEscapeSequencesRoundTripAllBytes(t *testing.T) {
	data := make([]byte, 256)
	for value := range data {
		data[value] = byte(value)
	}

	encoded := AppendEscapeSequences(nil, data)
	decoded := appendDecodedEscapeSequences(nil, append([]byte(nil), encoded...))
	if !bytes.Equal(decoded, data) {
		t.Fatalf("decoded escaped bytes = %X, want %X", decoded, data)
	}
}

func TestAppendEscapeSequencesUsesReservedMappings(t *testing.T) {
	got := AppendEscapeSequences([]byte{0xAA}, []byte{0x01, 0x00, 0xFF, 0x02})
	want := []byte{0xAA, 0x01, 0x01, 0x01, 0x02, 0x01, 0x03, 0x02}
	if !bytes.Equal(got, want) {
		t.Fatalf("AppendEscapeSequences() = %X, want %X", got, want)
	}
}
