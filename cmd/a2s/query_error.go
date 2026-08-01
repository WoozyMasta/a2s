package main

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// userFacingError keeps the low-level cause available to callers
// while exposing a concise diagnostic suitable for CLI users.
type userFacingError struct {
	cause   error
	message string
}

func (e *userFacingError) Error() string {
	return e.message
}

func (e *userFacingError) Unwrap() error {
	return e.cause
}

// friendlyQueryError adds a clear explanation for query timeouts
// without hiding the original cause from programmatic error inspection.
func friendlyQueryError(prefix string, err error, timeout time.Duration) error {
	if errors.Is(err, context.DeadlineExceeded) {
		return &userFacingError{
			message: fmt.Sprintf(
				"%s: server did not respond within %s; it may be offline or unreachable",
				prefix,
				humanTimeout(timeout),
			),
			cause: err,
		}
	}

	return fmt.Errorf("%s: %w", prefix, err)
}

func humanTimeout(timeout time.Duration) string {
	if timeout > 0 && timeout%time.Second == 0 {
		seconds := int(timeout / time.Second)
		unit := "seconds"
		if seconds == 1 {
			unit = "second"
		}
		return fmt.Sprintf("%d %s", seconds, unit)
	}

	return timeout.String()
}
