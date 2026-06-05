package a2s

import (
	"errors"
	"io"
	"testing"
)

func TestParsePing(t *testing.T) {
	if err := parsePing([]byte("pong\x00")); err != nil {
		t.Fatalf("parsePing() error = %v, want nil", err)
	}
}

func TestParsePingMalformedPayload(t *testing.T) {
	err := parsePing([]byte("unterminated"))
	if !errors.Is(err, ErrPingRead) {
		t.Fatalf("parsePing() error = %v, want ErrPingRead", err)
	}
	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("parsePing() error = %v, want io.ErrUnexpectedEOF", err)
	}
}
