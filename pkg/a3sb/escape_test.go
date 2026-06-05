package a3sb

import (
	"bytes"
	"testing"
)

func TestAppendEscapeSequences(t *testing.T) {
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
			got := appendEscapeSequences([]byte{0xAA}, test.data)
			want := append([]byte{0xAA}, test.want...)
			if !bytes.Equal(got, want) {
				t.Fatalf("appendEscapeSequences() = %X, want %X", got, want)
			}
		})
	}
}

func TestAppendEscapeSequencesInPlace(t *testing.T) {
	data := []byte{0xAA, 0x01, 0x01, 0x01, 0x02, 0x01, 0x03, 0x01, 0x04, 0x01}

	got := appendEscapeSequences(nil, data)
	want := []byte{0xAA, 0x01, 0x00, 0xFF, 0x01, 0x04, 0x01}
	if !bytes.Equal(got, want) {
		t.Fatalf("appendEscapeSequences(nil, data) = %X, want %X", got, want)
	}
}
