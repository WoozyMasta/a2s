package bread

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
)

var (
	ErrUnderflow = errors.New("buffer underflow: not enough data to read")                  // Error if buffer underflow
	ErrNumber    = errors.New("failed to read number from buffer")                          // Error with read number
	ErrBool      = errors.New("unsupported boolean byte in buffer")                         // Error with read boolean
	ErrString    = errors.New("length of the string from the buffer is less than expected") // Error with read string
)

// Helper for check bytes in buffer not underflow
//
//nolint:unused // used in reader.go
func checkLength(buf *bytes.Buffer, size int) error {
	if buf.Len() < size {
		return fmt.Errorf("%w: got %d of expected %d bytes", ErrUnderflow, buf.Len(), size)
	}

	return nil
}

// Helper for read number in LittleEndian
// Optimized to avoid reflection overhead of binary.Read
//
//nolint:unused // used in reader.go
func readNumber(buf *bytes.Buffer, data any, size int) error {
	if buf.Len() < size {
		return fmt.Errorf("%w: got %d of expected %d bytes", ErrUnderflow, buf.Len(), size)
	}

	// Read bytes directly without reflection
	b := buf.Next(size)
	if len(b) != size {
		return fmt.Errorf("%w: read %d bytes, expected %d", ErrNumber, len(b), size)
	}

	// Use type switch for direct conversion (faster than reflection)
	switch v := data.(type) {
	case *int16:
		//nolint:gosec // intentional conversion for binary protocol parsing (A2S protocol uses signed integers)
		*v = int16(binary.LittleEndian.Uint16(b))
	case *int32:
		//nolint:gosec // intentional conversion for binary protocol parsing (A2S protocol uses signed integers)
		*v = int32(binary.LittleEndian.Uint32(b))
	case *int64:
		//nolint:gosec // intentional conversion for binary protocol parsing (A2S protocol uses signed integers)
		*v = int64(binary.LittleEndian.Uint64(b))
	case *uint16:
		*v = binary.LittleEndian.Uint16(b)
	case *uint32:
		*v = binary.LittleEndian.Uint32(b)
	case *uint64:
		*v = binary.LittleEndian.Uint64(b)
	case *float32:
		*v = math.Float32frombits(binary.LittleEndian.Uint32(b))
	case *float64:
		*v = math.Float64frombits(binary.LittleEndian.Uint64(b))
	default:
		// Fallback to reflection for unsupported types
		reader := bytes.NewReader(b)
		if err := binary.Read(reader, binary.LittleEndian, data); err != nil {
			return fmt.Errorf("%w: %w", ErrNumber, err)
		}
	}

	return nil
}
