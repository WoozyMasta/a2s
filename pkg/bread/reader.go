// Package bread is a set of helper functions for reading bytes into different types from a buffer.
package bread

import (
	"bytes"
	"fmt"
	"math"
	"time"
)

// Read bytes buffer and return byte
func Byte(buf *bytes.Buffer) (byte, error) {
	if err := checkLength(buf, 1); err != nil {
		return 0, err
	}

	return buf.ReadByte()
}

// Read bytes buffer and return boolean, where 1 is true, 0 is false, other error
func Bool(buf *bytes.Buffer) (bool, error) {
	value, err := Byte(buf)
	if err != nil {
		return false, err
	}

	switch value {
	case 1:
		return true, nil
	case 0:
		return false, nil
	default:
		return false, fmt.Errorf("%w: 0x%X", ErrBool, value)
	}
}

// Read bytes buffer and return int8
func Int8(buf *bytes.Buffer) (int8, error) {
	value, err := Byte(buf)
	if err != nil {
		return 0, err
	}

	return int8(value), nil
}

// Read bytes buffer and return int16
func Int16(buf *bytes.Buffer) (int16, error) {
	var value int16
	if err := readNumber(buf, &value, 2); err != nil {
		return 0, err
	}

	return value, nil
}

// Read bytes buffer and return int32
func Int32(buf *bytes.Buffer) (int32, error) {
	var value int32
	if err := readNumber(buf, &value, 4); err != nil {
		return 0, err
	}

	return value, nil
}

// Read bytes buffer and return int64
func Int64(buf *bytes.Buffer) (int64, error) {
	var value int64
	if err := readNumber(buf, &value, 8); err != nil {
		return 0, err
	}

	return value, nil
}

// Read bytes buffer and return uint16
func Uint16(buf *bytes.Buffer) (uint16, error) {
	var value uint16
	if err := readNumber(buf, &value, 2); err != nil {
		return 0, err
	}

	return value, nil
}

// Read bytes buffer and return uint32
func Uint32(buf *bytes.Buffer) (uint32, error) {
	var value uint32
	if err := readNumber(buf, &value, 4); err != nil {
		return 0, err
	}

	return value, nil
}

// Read bytes buffer and return uint64
func Uint64(buf *bytes.Buffer) (uint64, error) {
	var value uint64
	if err := readNumber(buf, &value, 8); err != nil {
		return 0, err
	}

	return value, nil
}

// Read bytes buffer and return float32
func Float32(buf *bytes.Buffer) (float32, error) {
	var value float32
	if err := readNumber(buf, &value, 4); err != nil {
		return 0, err
	}

	return value, nil
}

// Read bytes buffer and return float64
func Float64(buf *bytes.Buffer) (float64, error) {
	var value float64
	if err := readNumber(buf, &value, 8); err != nil {
		return 0, err
	}

	return value, nil
}

// Read bytes buffer to first 0x00 delimiter and return as string
func String(buf *bytes.Buffer) (string, error) {
	value, err := buf.ReadBytes(0x00)
	if err != nil {
		return "", err
	}

	return string(value[:len(value)-1]), nil
}

// Read bytes buffer by size count and return as string
func StringLen(buf *bytes.Buffer, size int) (string, error) {
	if err := checkLength(buf, size); err != nil {
		return "", err
	}

	value := make([]byte, size)
	n, err := buf.Read(value)
	if err != nil {
		return "", err
	}

	if n != size {
		return "", fmt.Errorf("%w: got '%s' with %d length but expected %d", ErrString, value, n, size)
	}

	return string(value), nil
}

// Read bytes buffer as float32 and return as time.Duration
func Duration32(buf *bytes.Buffer) (time.Duration, error) {
	f, err := Float32(buf)
	if err != nil {
		return 0, err
	}

	seconds := int64(f)
	nanoseconds := int64(math.Round(float64(f-float32(seconds)) * 1e9))

	return time.Duration(seconds)*time.Second + time.Duration(nanoseconds)*time.Nanosecond, nil
}

// Read bytes buffer as float64 and return as time.Duration
func Duration64(buf *bytes.Buffer) (time.Duration, error) {
	f, err := Float64(buf)
	if err != nil {
		return 0, err
	}

	seconds := int64(f)
	nanoseconds := int64(math.Round((f - float64(seconds)) * 1e9))

	return time.Duration(seconds)*time.Second + time.Duration(nanoseconds)*time.Nanosecond, nil
}

// Read bytes buffer to first 0x00 delimiter
func BytesPage(buf *bytes.Buffer) ([]byte, error) {
	value, err := buf.ReadBytes(0x00)
	if err != nil {
		return nil, err
	}
	n := len(value) - 1

	return value[:n], nil
}

// Replace Escape sequence to Escape value
//
//	{0x01, 0x01} -> 0x01
//	{0x01, 0x02} -> 0x00
//	{0x01, 0x03} -> 0xFF
func EscapeSequences(data []byte) []byte {
	var buf bytes.Buffer

	for i := 0; i < len(data); i++ {
		if data[i] == 0x01 && i+1 < len(data) {
			switch data[i+1] {
			case 0x01:
				buf.WriteByte(0x01)
				i++
			case 0x02:
				buf.WriteByte(0x00)
				i++
			case 0x03:
				buf.WriteByte(0xFF)
				i++
			default:
				buf.WriteByte(data[i])
			}
		} else {
			buf.WriteByte(data[i])
		}
	}

	return buf.Bytes()
}
