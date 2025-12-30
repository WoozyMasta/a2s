package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

// Formatter handles output formatting in different formats.
type Formatter struct {
	format string
}

// NewFormatter creates a new formatter with the specified format.
func NewFormatter(format string) *Formatter {
	normalized := format
	if normalized == "" {
		normalized = "table"
	}
	normalized = strings.ToLower(normalized)

	return &Formatter{format: normalized}
}

// PrintTable prints data as a table in the specified format.
func (f *Formatter) PrintTable(t table.Writer) {
	switch f.format {
	case "json":
		fmt.Println("{}")
	case "raw":
		t.Style().Format.Header = text.FormatDefault
		t.Style().Format.Footer = text.FormatDefault
		t.Render()
	case "md", "markdown":
		fmt.Println(t.RenderMarkdown())
	case "html":
		fmt.Println(t.RenderHTML())
	case "table":
		fallthrough
	default:
		t.Render()
	}
}

// PrintJSON prints data as JSON.
func (f *Formatter) PrintJSON(data any) {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fatalf("Failed to marshal JSON: %v", err)
	}
	fmt.Println(string(jsonData))
}

// PrintRaw prints raw data (for rules).
func (f *Formatter) PrintRaw(data map[string]string) {
	if f.format == "json" {
		f.PrintJSON(data)
		return
	}

	t := table.NewWriter()
	t.SetStyle(table.StyleRounded)

	if f.format != "md" && f.format != "markdown" && f.format != "html" {
		t.SetOutputMirror(os.Stdout)
	}
	t.AppendHeader(table.Row{"Rule", "Value"})

	for k, v := range data {
		t.AppendRow(table.Row{k, v})
	}

	f.PrintTable(t)
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
		fmt.Printf("\n## %s\n\n", title)
	case "html":
		fmt.Printf("<h2>%s</h2>\n", title)
	}
}
