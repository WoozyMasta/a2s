package bread

import "errors"

var (
	ErrUnderflow = errors.New("buffer underflow: not enough data to read")
	ErrBool      = errors.New("unsupported boolean byte in buffer")
	ErrString    = errors.New("length of the string from the buffer is less than expected")
)
