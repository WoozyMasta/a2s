package main

import "fmt"

// localize resolves an application message and applies printf-style arguments.
// English source text remains the fallback when no catalog entry is available.
func (app *Application) localize(key, fallback string, args ...any) string {
	if app != nil && app.Localizer != nil {
		fallback = app.Localizer.Localize(key, fallback, nil)
	}

	return fmt.Sprintf(fallback, args...)
}

// formatBool renders a boolean for human-readable localized output.
func (app *Application) formatBool(value bool) string {
	if value {
		return app.localize("value.true", "true")
	}

	return app.localize("value.false", "false")
}

// wrapError adds localized application context without replacing the cause.
func (app *Application) wrapError(key, fallback string, err error) error {
	if err == nil {
		return nil
	}

	return fmt.Errorf("%s: %w", app.localize(key, fallback), err)
}
