package bread

import "errors"

var (
	// ErrUnderflow indicates that a fixed-width read exceeded remaining data.
	ErrUnderflow = errors.New("buffer underflow: not enough data to read")

	// ErrBool indicates that a boolean field contained a value other than 0 or 1.
	ErrBool = errors.New("unsupported boolean byte in buffer")

	// ErrString indicates that a string field was not terminated as expected.
	ErrString = errors.New("length of the string from the buffer is less than expected")
)
