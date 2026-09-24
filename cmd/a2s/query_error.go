// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package main

import (
	"context"
	"errors"
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

// friendlyQueryError adds localized context for query failures without hiding
// the original cause from programmatic error inspection.
func friendlyQueryError(app *Application, key, prefix string, err error, timeout time.Duration) error {
	if errors.Is(err, context.DeadlineExceeded) {
		return &userFacingError{
			message: app.localize(
				"error.query_timeout",
				"%s: server did not respond within %s; it may be offline or unreachable",
				app.localize(key, prefix),
				humanTimeout(app, timeout),
			),
			cause: err,
		}
	}

	return app.wrapError(key, prefix, err)
}

func humanTimeout(app *Application, timeout time.Duration) string {
	if timeout > 0 && timeout%time.Second == 0 {
		seconds := int(timeout / time.Second)
		key := "time.seconds"
		fallback := "%d seconds"

		if seconds == 1 {
			key = "time.second"
			fallback = "%d second"
		}

		return app.localize(key, fallback, seconds)
	}

	return timeout.String()
}
