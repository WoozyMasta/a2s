// Package wire provides a small, protocol-neutral decoder
// for little-endian binary data backed by an in-memory byte slice.
// Protocol-specific semantics remain in the owning A2S/A3SB packages;
// this package intentionally provides no generic writer or encoder.
// As an internal package, its API has no compatibility guarantee.
package wire

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"math"
)

var (
	// ErrInvalidLength indicates that a read was requested with a negative length.
	ErrInvalidLength = errors.New("wire: invalid negative length")
)

// Decoder reads protocol-neutral binary values from an in-memory byte slice.
// All reads are little-endian and advance the offset only after succeeding.
// Slices returned by Bytes and Tail borrow the input passed to NewDecoder.
type Decoder struct {
	data []byte
	off  int
}

// NewDecoder creates a decoder over data without copying it.
func NewDecoder(data []byte) Decoder {
	return Decoder{data: data}
}

// Byte reads one byte.
func (d *Decoder) Byte() (byte, error) {
	data, err := d.take(1)
	if err != nil {
		return 0, err
	}

	return data[0], nil
}

// Uint16 reads a little-endian uint16.
func (d *Decoder) Uint16() (uint16, error) {
	data, err := d.take(2)
	if err != nil {
		return 0, err
	}

	return binary.LittleEndian.Uint16(data), nil
}

// Uint32 reads a little-endian uint32.
func (d *Decoder) Uint32() (uint32, error) {
	data, err := d.take(4)
	if err != nil {
		return 0, err
	}

	return binary.LittleEndian.Uint32(data), nil
}

// Uint64 reads a little-endian uint64.
func (d *Decoder) Uint64() (uint64, error) {
	data, err := d.take(8)
	if err != nil {
		return 0, err
	}

	return binary.LittleEndian.Uint64(data), nil
}

// Int32 reads a little-endian signed int32.
func (d *Decoder) Int32() (int32, error) {
	data, err := d.take(4)
	if err != nil {
		return 0, err
	}

	// Preserve the two's-complement bit pattern of the signed wire value.
	// #nosec G115 -- intentional signed wire conversion
	return int32(binary.LittleEndian.Uint32(data)), nil
}

// Float32 reads a little-endian IEEE 754 float32.
func (d *Decoder) Float32() (float32, error) {
	data, err := d.take(4)
	if err != nil {
		return 0, err
	}

	return math.Float32frombits(binary.LittleEndian.Uint32(data)), nil
}

// Bytes reads n bytes and returns a borrowed slice backed by the decoder input.
// The returned slice must be treated as read-only.
func (d *Decoder) Bytes(n int) ([]byte, error) {
	return d.take(n)
}

// CStringBytes reads a NUL-terminated byte sequence
// and returns a borrowed slice backed by the decoder input.
// The returned slice must be treated as read-only.
func (d *Decoder) CStringBytes() ([]byte, error) {
	tail := d.data[d.off:]
	n := bytes.IndexByte(tail, 0)
	if n < 0 {
		return nil, io.ErrUnexpectedEOF
	}

	d.off += n + 1
	return tail[:n], nil
}

// CString reads a NUL-terminated string.
func (d *Decoder) CString() (string, error) {
	value, err := d.CStringBytes()
	if err != nil {
		return "", err
	}

	return string(value), nil
}

// FixedString reads exactly n bytes and returns them as a string.
func (d *Decoder) FixedString(n int) (string, error) {
	value, err := d.Bytes(n)
	if err != nil {
		return "", err
	}

	return string(value), nil
}

// Offset returns the number of bytes consumed from the input.
func (d *Decoder) Offset() int {
	return d.off
}

// Remaining returns the number of unread bytes.
func (d *Decoder) Remaining() int {
	return len(d.data) - d.off
}

// Empty reports whether all input bytes have been consumed.
func (d *Decoder) Empty() bool {
	return d.Remaining() == 0
}

// Tail returns all unread bytes without advancing the decoder.
// The returned slice borrows the decoder input and must be treated as read-only.
func (d *Decoder) Tail() []byte {
	return d.data[d.off:]
}

// take validates and consumes n bytes.
// It updates the offset only after the complete range has been validated.
func (d *Decoder) take(n int) ([]byte, error) {
	if n < 0 {
		return nil, ErrInvalidLength
	}
	if n > d.Remaining() {
		return nil, io.ErrUnexpectedEOF
	}

	start := d.off
	d.off += n

	return d.data[start:d.off], nil
}
