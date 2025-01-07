package bread

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
)

var (
	ErrUnderflow = errors.New("buffer underflow: not enough data to read")                  // Error if buffer underflow
	ErrNumber    = errors.New("failed to read number from buffer")                          // Error with read number
	ErrBool      = errors.New("unsupported boolean byte in buffer")                         // Error with read boolean
	ErrString    = errors.New("length of the string from the buffer is less than expected") // Error with read string
)

// Helper for check bytes in buffer not underflow
func checkLength(buf *bytes.Buffer, size int) error {
	if buf.Len() < size {
		return fmt.Errorf("%w: got %d of expected %d bytes", ErrUnderflow, buf.Len(), size)
	}

	return nil
}

// Helper for read number in LittleEndian
func readNumber(buf *bytes.Buffer, data any, size int) error {
	if buf.Len() < size {
		return fmt.Errorf("%w: got %d of expected %d bytes", ErrUnderflow, buf.Len(), size)
	}

	if err := binary.Read(buf, binary.LittleEndian, data); err != nil {
		return fmt.Errorf("%w: %w", ErrNumber, err)
	}

	return nil
}
