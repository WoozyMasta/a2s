// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

func TestNewFormatterDetectsNoTerminalWidthForNonTTY(t *testing.T) {
	var output bytes.Buffer

	formatter := NewFormatter("table", &output)

	if formatter.terminalWidth != 0 {
		t.Fatalf("terminal width = %d, want 0 for non-TTY output", formatter.terminalWidth)
	}
}

func TestNewFormatterDoesNotDetectTerminalWidthForNonTableFormats(t *testing.T) {
	formats := []string{"json", "raw", "md", "markdown", "html"}

	for _, format := range formats {
		t.Run(format, func(t *testing.T) {
			var output bytes.Buffer

			formatter := NewFormatter(format, &output)

			if formatter.terminalWidth != 0 {
				t.Fatalf("terminal width = %d, want 0 for %q output", formatter.terminalWidth, format)
			}
		})
	}
}

func TestDetectTableTerminalWidthRejectsNonTTYWriter(t *testing.T) {
	var output bytes.Buffer

	if width := detectTableTerminalWidth("table", &output); width != 0 {
		t.Fatalf("terminal width = %d, want 0 for non-TTY writer", width)
	}
}

func TestFormatterNewTableAppliesResponsiveWidths(t *testing.T) {
	var output bytes.Buffer
	formatter := NewFormatter("table", &output)
	formatter.terminalWidth = 37

	rows := make([]table.Row, 100)
	for index := range rows {
		rows[index] = table.Row{"short", strings.Repeat("v", 20)}
	}
	rows[len(rows)-1][0] = strings.Repeat("x", 80)

	if err := formatter.PrintTable(formatter.NewTable(
		table.Row{"Key", "Value"},
		rows,
	)); err != nil {
		t.Fatalf("PrintTable() error = %v", err)
	}

	for _, line := range strings.Split(strings.TrimSuffix(output.String(), "\n"), "\n") {
		if width := text.StringWidthWithoutEscSequences(line); width > 37 {
			t.Fatalf("rendered line width = %d, want <= 37: %q", width, line)
		}
	}
}

func TestFormatterNewTableKeepsNonTableFormatsNatural(t *testing.T) {
	var output bytes.Buffer
	formatter := NewFormatter("markdown", &output)
	formatter.terminalWidth = 20
	longValue := strings.Repeat("value", 10)

	if err := formatter.PrintTable(formatter.NewTable(
		table.Row{"Key", "Value"},
		[]table.Row{{"key", longValue}},
	)); err != nil {
		t.Fatalf("PrintTable() error = %v", err)
	}

	if !strings.Contains(output.String(), longValue) {
		t.Fatalf("markdown output does not contain the complete value: %q", output.String())
	}
}
