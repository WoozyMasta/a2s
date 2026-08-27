// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/woozymasta/flags"
)

// Formatter handles output formatting in different formats.
type Formatter struct {
	out           io.Writer
	localizer     *flags.Localizer
	format        string
	terminalWidth int
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

	return &Formatter{
		format:        normalized,
		out:           out,
		localizer:     localizer,
		terminalWidth: detectTableTerminalWidth(normalized, out),
	}
}

// detectTableTerminalWidth returns the terminal width for interactive tables.
// Non-table formats and non-TTY writers intentionally return zero
// so their output remains independent of the current terminal.
func detectTableTerminalWidth(format string, out io.Writer) int {
	if format != "table" {
		return 0
	}

	file, ok := out.(*os.File)
	if !ok || !flags.DetectFileTTY(file) {
		return 0
	}

	width, _ := flags.DetectTerminalSize()
	if width <= 0 {
		return 0
	}

	return width
}

// NewTable creates a rounded table and applies responsive widths
// for an interactive table output.
// Other output formats keep their natural layout.
func (f *Formatter) NewTable(
	header table.Row,
	rows []table.Row,
	hints ...columnWidthHint,
) table.Writer {
	t := table.NewWriter()
	t.SetStyle(table.StyleRounded)
	t.AppendHeader(header)
	t.AppendRows(rows)

	if f.IsTableFormat() && f.terminalWidth > 0 {
		layout, err := planTableLayout(f.terminalWidth, header, rows, hints)
		if err == nil {
			configs := make([]table.ColumnConfig, len(layout.Widths))
			for columnIndex, width := range layout.Widths {
				configs[columnIndex] = table.ColumnConfig{
					Number:           columnIndex + 1,
					WidthMax:         width,
					WidthMaxEnforcer: text.WrapText,
				}
			}

			t.SetColumnConfigs(configs)
			t.Style().Size.WidthMax = f.terminalWidth
		}
	}

	return t
}

// PrintTable prints data as a table in the specified format.
func (f *Formatter) PrintTable(t table.Writer) error {
	t.SetOutputMirror(f.out)

	switch f.format {
	case "json":
		_, _ = fmt.Fprintln(f.out, "{}")

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
