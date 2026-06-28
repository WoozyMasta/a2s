package a3sb

import (
	"bytes"
	"errors"
	"testing"

	"github.com/woozymasta/a2s/internal/testfixtures"
)

func TestEncodePagesUsesDefaultSizeAndWireOrder(t *testing.T) {
	data := bytes.Repeat([]byte("x"), DefaultPageSize+1)
	rules, err := EncodePages(data, 0)
	if err != nil {
		t.Fatalf("EncodePages() error = %v", err)
	}
	if len(rules) != 2 {
		t.Fatalf("page count = %d, want 2", len(rules))
	}
	if rules[0].Name != string([]byte{1, 2}) || rules[1].Name != string([]byte{2, 2}) {
		t.Fatalf("page names = % X, % X, want 01 02, 02 02", rules[0].Name, rules[1].Name)
	}
	if len(rules[0].Value) != DefaultPageSize || len(rules[1].Value) != 1 {
		t.Fatalf("page sizes = %d, %d, want %d, 1", len(rules[0].Value), len(rules[1].Value), DefaultPageSize)
	}
}

func TestEncodePagesRoundTripsEscapedBinary(t *testing.T) {
	binaryData, err := testfixtures.Read("a3sb_arma3_payload.hex")
	if err != nil {
		t.Fatal(err)
	}
	escaped := AppendEscapeSequences(nil, binaryData)
	rules, err := EncodePages(escaped, 7)
	if err != nil {
		t.Fatalf("EncodePages() error = %v", err)
	}

	assembled := make([]byte, 0, len(escaped))
	for _, rule := range rules {
		assembled = append(assembled, rule.Value...)
	}
	decoded := appendDecodedEscapeSequences(nil, assembled)
	if !bytes.Equal(decoded, binaryData) {
		t.Fatalf("decoded paged binary = %X, want %X", decoded, binaryData)
	}
}

func TestEncodePagesRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		size int
		want error
	}{
		{
			name: "negative size",
			data: []byte{1},
			size: -1,
			want: ErrPageSize,
		},
		{
			name: "empty data",
			size: DefaultPageSize,
			want: ErrPageEncode,
		},
		{
			name: "unescaped NUL",
			data: []byte{1, 0, 2},
			size: DefaultPageSize,
			want: ErrPageEncode,
		},
		{
			name: "page count overflow",
			data: bytes.Repeat([]byte{2}, DefaultPageSize*256),
			size: DefaultPageSize,
			want: ErrPageCount,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := EncodePages(test.data, test.size)
			if !errors.Is(err, test.want) {
				t.Fatalf("EncodePages() error = %v, want %v", err, test.want)
			}
		})
	}
}
