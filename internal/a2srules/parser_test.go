package a2srules

import (
	"bytes"
	"errors"
	"testing"
)

func TestParsePreservesOrderDuplicatesAndRemaining(t *testing.T) {
	data := []byte{
		2, 0,
		'a', 0, '1', 0,
		'a', 0, '2', 0,
		0xFF,
	}

	result, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if len(result.Entries) != 2 {
		t.Fatalf("entry count = %d, want 2", len(result.Entries))
	}
	if string(result.Entries[0].Key) != "a" || string(result.Entries[0].Value) != "1" {
		t.Fatalf("first entry = %q/%q, want %q/%q", result.Entries[0].Key, result.Entries[0].Value, "a", "1")
	}
	if string(result.Entries[1].Value) != "2" {
		t.Fatalf("second value = %q, want %q", result.Entries[1].Value, "2")
	}
	if !bytes.Equal(result.Remaining, []byte{0xFF}) {
		t.Fatalf("remaining = %X, want FF", result.Remaining)
	}

	rules := Map(result.Entries)
	if rules["a"] != "2" {
		t.Fatalf("mapped duplicate = %q, want %q", rules["a"], "2")
	}
}

func TestParseRejectsMalformedEntries(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want error
	}{
		{name: "count", data: []byte{1}, want: ErrCount},
		{name: "entry header", data: []byte{1, 0, 'a'}, want: ErrInsufficientData},
		{name: "key", data: []byte{1, 0, 'a', 'v', 'x', 'y'}, want: ErrKey},
		{name: "value", data: []byte{1, 0, 'a', 0, 'v', 'x'}, want: ErrValue},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := Parse(test.data)
			if !errors.Is(err, test.want) {
				t.Fatalf("Parse error = %v, want %v", err, test.want)
			}
		})
	}
}

func TestParseZeroCountKeepsRemaining(t *testing.T) {
	result, err := Parse([]byte{0, 0, 1, 2})
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if !bytes.Equal(result.Remaining, []byte{1, 2}) {
		t.Fatalf("remaining = %X, want 0102", result.Remaining)
	}
	if Map(result.Entries) != nil {
		t.Fatal("Map returned non-nil map for empty entries")
	}
}
