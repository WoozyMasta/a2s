package main

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/woozymasta/flags"
)

// Formatter handles output formatting in different formats.
type Formatter struct {
	out       io.Writer
	localizer *flags.Localizer
	format    string
}

// NewFormatter creates a new formatter with the specified format.
func NewFormatter(format string, out io.Writer, localizers ...*flags.Localizer) *Formatter {
	normalized := format
	if normalized == "" {
		normalized = "table"
	}
	normalized = strings.ToLower(normalized)

	if out == nil {
		out = io.Discard
	}

	var localizer *flags.Localizer
	if len(localizers) > 0 {
		localizer = localizers[0]
	}

	return &Formatter{format: normalized, out: out, localizer: localizer}
}

// PrintTable prints data as a table in the specified format.
func (f *Formatter) PrintTable(t table.Writer) error {
	t.SetOutputMirror(f.out)

	switch f.format {
	case "json":
		_, _ = fmt.Fprintln(f.out, "{}")

	case "raw":
		t.Style().Format.Header = text.FormatDefault
		t.Style().Format.Footer = text.FormatDefault
		t.Render()

	case "md", "markdown":
		_, _ = fmt.Fprintln(f.out, t.RenderMarkdown())

	case "html":
		_, _ = fmt.Fprintln(f.out, t.RenderHTML())

	case "table":
		fallthrough

	default:
		t.Render()
	}

	return nil
}

// PrintJSON prints data as JSON.
func (f *Formatter) PrintJSON(data any) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf(
			"%s: %w",
			localizeFormatterError(f.localizer, "error.json_marshal", "failed to marshal JSON"),
			err,
		)
	}
	_, _ = fmt.Fprintln(f.out, string(jsonData))

	return nil
}

// localizeFormatterError resolves formatter errors
// without coupling it to the full application object.
func localizeFormatterError(localizer *flags.Localizer, key, fallback string) string {
	if localizer == nil {
		return fallback
	}

	return localizer.Localize(key, fallback, nil)
}

// PrintRaw prints raw data (for rules).
func (f *Formatter) PrintRaw(data map[string]string) error {
	if f.format == "json" {
		return f.PrintJSON(data)
	}

	t := table.NewWriter()
	t.SetStyle(table.StyleRounded)
	t.AppendHeader(table.Row{"Rule", "Value"})

	for k, v := range data {
		t.AppendRow(table.Row{k, v})
	}

	return f.PrintTable(t)
}

// ShouldUseJSON returns true if the format is JSON.
func (f *Formatter) ShouldUseJSON() bool {
	return f.format == "json"
}

// GetFormat returns the normalized format string.
func (f *Formatter) GetFormat() string {
	return f.format
}

// IsTableFormat returns true if the format is table (default).
func (f *Formatter) IsTableFormat() bool {
	return f.format == "table" || f.format == ""
}

// PrintSectionHeader prints a section header for md/html formats to separate tables.
func (f *Formatter) PrintSectionHeader(title string) {
	switch f.format {
	case "md", "markdown":
		_, _ = fmt.Fprintf(f.out, "\n## %s\n\n", title)

	case "html":
		_, _ = fmt.Fprintf(f.out, "<h2>%s</h2>\n", title)
	}
}
