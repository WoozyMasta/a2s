// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package main

import (
	"strings"
	"testing"

	"github.com/jedib0t/go-pretty/v6/table"
)

func TestDisplayWidth(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  int
	}{
		{name: "ascii", value: "server", want: 6},
		{name: "cyrillic", value: "Сервер", want: 6},
		{name: "wide rune", value: "服务器", want: 6},
		{name: "ansi", value: "\x1b[31mred\x1b[0m", want: 3},
		{name: "multiline", value: "short\nlongest", want: 7},
		{name: "empty", value: "", want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := displayWidth(tt.value); got != tt.want {
				t.Fatalf("displayWidth(%q) = %d, want %d", tt.value, got, tt.want)
			}
		})
	}
}

func TestAnalyzeTableLayoutUsesRobustColumnWidths(t *testing.T) {
	rows := make([]table.Row, 100)
	for index := range rows {
		rows[index] = table.Row{"shortkey", "medium value"}
	}
	rows[len(rows)-1][0] = strings.Repeat("x", 80)

	layout, err := analyzeTableLayout(table.Row{"Rule", "Value"}, rows, nil)
	if err != nil {
		t.Fatalf("analyzeTableLayout() error = %v", err)
	}

	stats := layout.Stats[0]
	if stats.Max != 80 {
		t.Fatalf("max width = %d, want 80", stats.Max)
	}
	if stats.P90 != 8 {
		t.Fatalf("p90 width = %d, want 8", stats.P90)
	}
	if stats.Preferred != 8 {
		t.Fatalf("preferred width = %d, want 8", stats.Preferred)
	}
	if stats.Mean <= 8 || stats.Mean >= 9 {
		t.Fatalf("mean width = %f, want a value between 8 and 9", stats.Mean)
	}
	if layout.NaturalWidths[0] != 80 {
		t.Fatalf("natural width = %d, want 80", layout.NaturalWidths[0])
	}

	if layout.PreferredWidths[1] != len("medium value") {
		t.Fatalf("value preferred width = %d, want %d", layout.PreferredWidths[1], len("medium value"))
	}
}

func TestAnalyzeTableLayoutIncludesHeaderAndMissingCells(t *testing.T) {
	layout, err := analyzeTableLayout(
		table.Row{"Long Header", "Value"},
		[]table.Row{{"x"}},
		nil,
	)
	if err != nil {
		t.Fatalf("analyzeTableLayout() error = %v", err)
	}

	if layout.Stats[0].Header != len("Long Header") {
		t.Fatalf("header width = %d, want %d", layout.Stats[0].Header, len("Long Header"))
	}
	if layout.PreferredWidths[0] != len("Long Header") {
		t.Fatalf("preferred width = %d, want %d", layout.PreferredWidths[0], len("Long Header"))
	}
	if layout.Stats[1].Max != 0 {
		t.Fatalf("missing cell max width = %d, want 0", layout.Stats[1].Max)
	}
}

func TestAnalyzeTableLayoutRejectsExtraCells(t *testing.T) {
	_, err := analyzeTableLayout(
		table.Row{"Rule"},
		[]table.Row{{"key", "value"}},
		nil,
	)
	if err == nil {
		t.Fatal("analyzeTableLayout() error = nil, want an error")
	}
}

func TestAnalyzeTableLayoutAppliesWidthHints(t *testing.T) {
	layout, err := analyzeTableLayout(
		table.Row{"Key"},
		[]table.Row{{strings.Repeat("x", 20)}},
		[]columnWidthHint{{Min: 6, Max: 10}},
	)
	if err != nil {
		t.Fatalf("analyzeTableLayout() error = %v", err)
	}

	if layout.NaturalWidths[0] != 10 {
		t.Fatalf("natural width = %d, want 10", layout.NaturalWidths[0])
	}
	if layout.PreferredWidths[0] != 10 {
		t.Fatalf("preferred width = %d, want 10", layout.PreferredWidths[0])
	}
}

func TestPercentileWidth(t *testing.T) {
	widths := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	if got := percentileWidth(widths, 90); got != 9 {
		t.Fatalf("percentileWidth(90) = %d, want 9", got)
	}
	if got := percentileWidth(widths, 0); got != 1 {
		t.Fatalf("percentileWidth(0) = %d, want 1", got)
	}
	if got := percentileWidth(widths, 100); got != 10 {
		t.Fatalf("percentileWidth(100) = %d, want 10", got)
	}
}

func TestPlanTableLayoutUsesNaturalWidthsWhenTheyFit(t *testing.T) {
	layout, err := planTableLayout(
		86,
		table.Row{"Rule", "Value"},
		[]table.Row{{"key", "value"}},
		nil,
	)
	if err != nil {
		t.Fatalf("planTableLayout() error = %v", err)
	}

	if !layout.FitsNatural || !layout.FitsPreferred {
		t.Fatalf("fit flags = natural:%t preferred:%t, want both true", layout.FitsNatural, layout.FitsPreferred)
	}
	if got, want := layout.ContentBudget, 79; got != want {
		t.Fatalf("content budget = %d, want %d", got, want)
	}
	if got, want := layout.Widths, []int{4, 5}; !sameInts(got, want) {
		t.Fatalf("widths = %v, want %v", got, want)
	}
}

func TestPlanTableLayoutGrowsTowardPreferredThenNatural(t *testing.T) {
	rows := make([]table.Row, 100)
	for index := range rows {
		rows[index] = table.Row{"short", strings.Repeat("v", 20)}
	}
	rows[len(rows)-1][0] = strings.Repeat("x", 80)

	layout, err := planTableLayout(37, table.Row{"Key", "Value"}, rows, nil)
	if err != nil {
		t.Fatalf("planTableLayout() error = %v", err)
	}

	if layout.FitsNatural {
		t.Fatal("FitsNatural = true, want false")
	}
	if !layout.FitsPreferred {
		t.Fatal("FitsPreferred = false, want true")
	}
	if got, want := layout.Widths, []int{10, 20}; !sameInts(got, want) {
		t.Fatalf("widths = %v, want %v", got, want)
	}
}

func TestPlanTableLayoutAllocatesConstrainedWidthByMarginalBenefit(t *testing.T) {
	rows := make([]table.Row, 100)
	for index := range rows {
		rows[index] = table.Row{"shortkey", strings.Repeat("v", 30)}
	}
	rows[len(rows)-1][0] = strings.Repeat("x", 80)

	layout, err := planTableLayout(45, table.Row{"Key", "Value"}, rows, nil)
	if err != nil {
		t.Fatalf("planTableLayout() error = %v", err)
	}

	if got, want := layout.Widths, []int{8, 30}; !sameInts(got, want) {
		t.Fatalf("widths = %v, want %v", got, want)
	}
}

func TestPlanTableLayoutUsesWeightAsTieBreaker(t *testing.T) {
	rows := []table.Row{{"12345", "12345"}}
	layout, err := planTableLayout(
		10,
		table.Row{"A", "B"},
		rows,
		[]columnWidthHint{{Weight: 2}, {Weight: 1}},
	)
	if err != nil {
		t.Fatalf("planTableLayout() error = %v", err)
	}

	if got, want := layout.Widths, []int{2, 1}; !sameInts(got, want) {
		t.Fatalf("widths = %v, want %v", got, want)
	}
}

func TestPlanTableLayoutKeepsWidthsPositiveWhenTerminalIsTooNarrow(t *testing.T) {
	layout, err := planTableLayout(5, table.Row{"Key", "Value"}, nil, nil)
	if err != nil {
		t.Fatalf("planTableLayout() error = %v", err)
	}

	if got, want := layout.Widths, []int{1, 1}; !sameInts(got, want) {
		t.Fatalf("widths = %v, want %v", got, want)
	}
}

func TestPlanTableLayoutAppliesMinimumAndMaximumHints(t *testing.T) {
	layout, err := planTableLayout(
		15,
		table.Row{"Key", "Value"},
		[]table.Row{{strings.Repeat("x", 20), "value"}},
		[]columnWidthHint{{Min: 4, Max: 6}},
	)
	if err != nil {
		t.Fatalf("planTableLayout() error = %v", err)
	}

	if got, want := layout.NaturalWidths[0], 6; got != want {
		t.Fatalf("natural width = %d, want %d", got, want)
	}
	if layout.Widths[0] < 4 || layout.Widths[0] > 6 {
		t.Fatalf("allocated width = %d, want it in [4, 6]", layout.Widths[0])
	}
}

func sameInts(left, right []int) bool {
	if len(left) != len(right) {
		return false
	}

	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}

	return true
}
