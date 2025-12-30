// Package bread is a set of helper functions for reading bytes into different types from a buffer.
// This file contains optimized versions that work directly with []byte to avoid bytes.Buffer allocations.
package bread

import (
	"encoding/binary"
	"math"
	"time"
)

// Reader is a simple byte reader that works directly with []byte
// to avoid bytes.Buffer allocations
type Reader struct {
	data []byte
	pos  int
}

// NewReader creates a new reader from []byte
func NewReader(data []byte) *Reader {
	return &Reader{data: data, pos: 0}
}

// Reset resets the reader to the beginning
func (r *Reader) Reset(data []byte) {
	r.data = data
	r.pos = 0
}

// Pos returns current position
func (r *Reader) Pos() int {
	return r.pos
}

// Len returns remaining bytes
func (r *Reader) Len() int {
	return len(r.data) - r.pos
}

// Byte reads a single byte
func (r *Reader) Byte() (byte, error) {
	if r.pos >= len(r.data) {
		return 0, ErrUnderflow
	}
	b := r.data[r.pos]
	r.pos++
	return b, nil
}

// Bool reads a boolean (1 = true, 0 = false)
func (r *Reader) Bool() (bool, error) {
	value, err := r.Byte()
	if err != nil {
		return false, err
	}

	switch value {
	case 1:
		return true, nil
	case 0:
		return false, nil
	default:
		return false, ErrBool
	}
}

// Uint16 reads uint16 in LittleEndian
func (r *Reader) Uint16() (uint16, error) {
	if r.pos+2 > len(r.data) {
		return 0, ErrUnderflow
	}
	value := binary.LittleEndian.Uint16(r.data[r.pos:])
	r.pos += 2
	return value, nil
}

// Uint32 reads uint32 in LittleEndian
func (r *Reader) Uint32() (uint32, error) {
	if r.pos+4 > len(r.data) {
		return 0, ErrUnderflow
	}
	value := binary.LittleEndian.Uint32(r.data[r.pos:])
	r.pos += 4
	return value, nil
}

// Uint64 reads uint64 in LittleEndian
func (r *Reader) Uint64() (uint64, error) {
	if r.pos+8 > len(r.data) {
		return 0, ErrUnderflow
	}
	value := binary.LittleEndian.Uint64(r.data[r.pos:])
	r.pos += 8
	return value, nil
}

// Float32 reads float32 in LittleEndian
func (r *Reader) Float32() (float32, error) {
	if r.pos+4 > len(r.data) {
		return 0, ErrUnderflow
	}
	bits := binary.LittleEndian.Uint32(r.data[r.pos:])
	r.pos += 4
	return math.Float32frombits(bits), nil
}

// Float64 reads float64 in LittleEndian
func (r *Reader) Float64() (float64, error) {
	if r.pos+8 > len(r.data) {
		return 0, ErrUnderflow
	}
	bits := binary.LittleEndian.Uint64(r.data[r.pos:])
	r.pos += 8
	return math.Float64frombits(bits), nil
}

// String reads a null-terminated string
// Returns the string without the null terminator
// No allocation for the string data itself (Go optimizes string([]byte))
func (r *Reader) String() (string, error) {
	start := r.pos
	// Find null terminator
	for r.pos < len(r.data) && r.data[r.pos] != 0 {
		r.pos++
	}

	if r.pos >= len(r.data) {
		return "", ErrString
	}

	// Extract string (no allocation - Go reuses underlying bytes)
	str := string(r.data[start:r.pos])
	r.pos++ // Skip null terminator
	return str, nil
}

// BytesPage reads bytes until null terminator (returns []byte, not string)
// Returns slice pointing to original data (no copy)
func (r *Reader) BytesPage() ([]byte, error) {
	start := r.pos
	// Find null terminator
	for r.pos < len(r.data) && r.data[r.pos] != 0 {
		r.pos++
	}

	if r.pos >= len(r.data) {
		return nil, ErrString
	}

	result := r.data[start:r.pos]
	r.pos++ // Skip null terminator
	return result, nil
}

// Duration32 reads float32 and converts to time.Duration
func (r *Reader) Duration32() (time.Duration, error) {
	f, err := r.Float32()
	if err != nil {
		return 0, err
	}

	seconds := int64(f)
	nanoseconds := int64(math.Round(float64(f-float32(seconds)) * 1e9))

	return time.Duration(seconds)*time.Second + time.Duration(nanoseconds)*time.Nanosecond, nil
}

// Duration64 reads float64 and converts to time.Duration
func (r *Reader) Duration64() (time.Duration, error) {
	f, err := r.Float64()
	if err != nil {
		return 0, err
	}

	seconds := int64(f)
	nanoseconds := int64(math.Round((f - float64(seconds)) * 1e9))

	return time.Duration(seconds)*time.Second + time.Duration(nanoseconds)*time.Nanosecond, nil
}
