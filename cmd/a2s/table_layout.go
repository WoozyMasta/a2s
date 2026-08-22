// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

// columnWidthHint provides optional semantic limits for one table column.
type columnWidthHint struct {
	Min    int // preferred lower bound for the column width
	Max    int // semantic upper bound for the column width
	Weight int // relative allocation weight for later width distribution
}

// columnStats contains display-width measurements for one table column.
type columnStats struct {
	Samples   []int // display widths of row cells in this column
	Header    int
	Mean      float64
	P90       int
	Max       int
	Preferred int
}

// tableLayout contains content-derived widths before terminal allocation.
type tableLayout struct {
	NaturalWidths   []int
	PreferredWidths []int
	Stats           []columnStats
	Widths          []int
	TerminalWidth   int
	ContentBudget   int
	FitsNatural     bool
	FitsPreferred   bool
}

// displayWidth returns the widest visual line in a table cell.
func displayWidth(value any) int {
	maxWidth := 0
	for line := range strings.SplitSeq(fmt.Sprint(value), "\n") {
		width := text.StringWidthWithoutEscSequences(line)
		if width > maxWidth {
			maxWidth = width
		}
	}

	return maxWidth
}

// analyzeTableLayout measures natural and preferred widths for table columns.
// Missing cells are treated as empty; rows with extra cells are invalid.
func analyzeTableLayout(
	header table.Row,
	rows []table.Row,
	hints []columnWidthHint,
) (tableLayout, error) {
	columnCount := len(header)
	layout := tableLayout{
		NaturalWidths:   make([]int, columnCount),
		PreferredWidths: make([]int, columnCount),
		Stats:           make([]columnStats, columnCount),
	}

	widths := make([][]int, columnCount)
	for columnIndex, value := range header {
		layout.Stats[columnIndex].Header = displayWidth(value)
	}

	for rowIndex, row := range rows {
		if len(row) > columnCount {
			return tableLayout{}, fmt.Errorf(
				"table row %d has %d columns, header has %d",
				rowIndex,
				len(row),
				columnCount,
			)
		}

		for columnIndex := range header {
			var value any = ""
			if columnIndex < len(row) {
				value = row[columnIndex]
			}
			widths[columnIndex] = append(widths[columnIndex], displayWidth(value))
		}
	}

	for columnIndex := range header {
		stats := &layout.Stats[columnIndex]
		samples := widths[columnIndex]
		stats.Samples = samples
		stats.Max = maxWidth(samples)
		stats.Mean = meanWidth(samples)
		stats.P90 = percentileWidth(samples, 90)

		natural := max(stats.Header, stats.Max)
		preferred := max(stats.Header, stats.P90)
		natural, preferred = applyWidthHint(natural, preferred, hintAt(hints, columnIndex))

		layout.NaturalWidths[columnIndex] = natural
		layout.PreferredWidths[columnIndex] = preferred
		stats.Preferred = preferred
	}

	return layout, nil
}

// planTableLayout analyzes table content and allocates widths for a terminal.
func planTableLayout(
	terminalWidth int,
	header table.Row,
	rows []table.Row,
	hints []columnWidthHint,
) (tableLayout, error) {
	layout, err := analyzeTableLayout(header, rows, hints)
	if err != nil {
		return tableLayout{}, err
	}

	layout.TerminalWidth = terminalWidth
	layout.ContentBudget = roundedTableContentBudget(terminalWidth, len(header))
	layout.Widths = append([]int(nil), layout.NaturalWidths...)

	naturalTotal := sumWidths(layout.NaturalWidths)
	preferredTotal := sumWidths(layout.PreferredWidths)
	if terminalWidth <= 0 {
		layout.FitsNatural = true
		layout.FitsPreferred = true
		return layout, nil
	}

	layout.FitsNatural = naturalTotal <= layout.ContentBudget
	layout.FitsPreferred = preferredTotal <= layout.ContentBudget
	if layout.FitsNatural {
		return layout, nil
	}

	layout.Widths = minimumWidths(hints, len(header), layout.ContentBudget)
	growWidths(layout.Widths, layout.Stats, layout.PreferredWidths, hints, layout.ContentBudget)
	growWidths(layout.Widths, layout.Stats, layout.NaturalWidths, hints, layout.ContentBudget)

	return layout, nil
}

// roundedTableContentBudget returns the approximate cell-content budget
// for go-pretty's rounded style:
// borders, separators, and one space of padding
// on both sides of every column are excluded from the terminal width.
func roundedTableContentBudget(terminalWidth, columnCount int) int {
	if terminalWidth <= 0 || columnCount <= 0 {
		return 0
	}

	return max(0, terminalWidth-(3*columnCount+1))
}

// minimumWidths creates a best-effort starting allocation for every column.
func minimumWidths(hints []columnWidthHint, columnCount, budget int) []int {
	widths := make([]int, columnCount)
	if columnCount == 0 {
		return widths
	}

	for columnIndex := range widths {
		widths[columnIndex] = 1
		hint := hintAt(hints, columnIndex)
		if hint.Min > widths[columnIndex] {
			widths[columnIndex] = hint.Min
		}
		if hint.Max > 0 && widths[columnIndex] > hint.Max {
			widths[columnIndex] = hint.Max
		}
	}

	if sumWidths(widths) <= budget {
		return widths
	}

	for columnIndex := range widths {
		widths[columnIndex] = 1
	}

	return widths
}

// growWidths distributes available cells toward a set of target widths.
func growWidths(
	widths []int,
	stats []columnStats,
	targets []int,
	hints []columnWidthHint,
	budget int,
) {
	for sumWidths(widths) < budget {
		bestColumn := -1
		bestScore := growthScore{}

		for columnIndex := range widths {
			if widths[columnIndex] >= targets[columnIndex] {
				continue
			}

			score := scoreColumnGrowth(
				widths[columnIndex],
				stats[columnIndex],
				hintAt(hints, columnIndex).Weight,
			)
			if bestColumn == -1 || score.betterThan(bestScore, columnIndex, bestColumn) {
				bestColumn = columnIndex
				bestScore = score
			}
		}

		if bestColumn == -1 {
			return
		}

		widths[bestColumn]++
	}
}

// growthScore describes the benefit of giving one more cell to a column.
type growthScore struct {
	Benefit  int
	Overflow int
	P90      int
	Mean     float64
}

// scoreColumnGrowth counts cells that still benefit from additional width.
func scoreColumnGrowth(current int, stats columnStats, weight int) growthScore {
	if weight <= 0 {
		weight = 1
	}

	benefit := 0
	if stats.Header > current {
		benefit = 2
	}

	overflow := 0
	for _, sample := range stats.Samples {
		if sample > current {
			benefit++
			overflow += sample - current
		}
	}

	return growthScore{
		Benefit:  benefit * weight,
		Overflow: overflow,
		P90:      stats.P90,
		Mean:     stats.Mean,
	}
}

// betterThan compares growth scores and keeps allocation deterministic.
func (score growthScore) betterThan(other growthScore, index, otherIndex int) bool {
	if score.Benefit != other.Benefit {
		return score.Benefit > other.Benefit
	}
	if score.Overflow != other.Overflow {
		return score.Overflow > other.Overflow
	}
	if score.P90 != other.P90 {
		return score.P90 > other.P90
	}
	if score.Mean != other.Mean {
		return score.Mean > other.Mean
	}

	return index < otherIndex
}

// sumWidths returns the total width of all columns.
func sumWidths(widths []int) int {
	total := 0
	for _, width := range widths {
		total += width
	}

	return total
}

// hintAt returns a column hint or its zero-value default.
func hintAt(hints []columnWidthHint, columnIndex int) columnWidthHint {
	if columnIndex >= len(hints) {
		return columnWidthHint{}
	}

	return hints[columnIndex]
}

// applyWidthHint applies optional semantic bounds
// without producing invalid negative or zero widths
// when a caller supplies incomplete hints.
func applyWidthHint(natural, preferred int, hint columnWidthHint) (int, int) {
	minWidth := max(0, hint.Min)
	maxWidth := hint.Max
	if maxWidth > 0 && minWidth > maxWidth {
		minWidth = maxWidth
	}

	if maxWidth > 0 {
		natural = min(natural, maxWidth)
		preferred = min(preferred, maxWidth)
	}

	natural = max(natural, minWidth)
	preferred = max(preferred, minWidth)
	if preferred > natural {
		natural = preferred
	}

	return natural, preferred
}

// maxWidth returns the largest sample width or zero for an empty column.
func maxWidth(widths []int) int {
	maxValue := 0
	for _, width := range widths {
		if width > maxValue {
			maxValue = width
		}
	}

	return maxValue
}

// meanWidth returns the arithmetic mean of sample widths.
func meanWidth(widths []int) float64 {
	if len(widths) == 0 {
		return 0
	}

	var total int
	for _, width := range widths {
		total += width
	}

	return float64(total) / float64(len(widths))
}

// percentileWidth returns a nearest-rank percentile of sample widths.
func percentileWidth(widths []int, percentile int) int {
	if len(widths) == 0 {
		return 0
	}

	sorted := append([]int(nil), widths...)
	sort.Ints(sorted)

	percentile = max(1, min(percentile, 100))
	rank := (percentile*len(sorted) + 99) / 100
	return sorted[rank-1]
}
