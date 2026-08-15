package main

import (
	"io"

	"github.com/woozymasta/flags"
)

// Application contains process-owned streams used by CLI commands.
type Application struct {
	// Out receives successful command output.
	Out io.Writer
	// Err receives diagnostics and cleanup warnings.
	Err io.Writer
	// Localizer resolves application messages using the parser's locale config.
	Localizer *flags.Localizer
}

// NewApplication creates an application with explicit output destinations.
func NewApplication(out, err io.Writer) *Application {
	if out == nil {
		out = io.Discard
	}
	if err == nil {
		err = io.Discard
	}

	return &Application{Out: out, Err: err}
}
