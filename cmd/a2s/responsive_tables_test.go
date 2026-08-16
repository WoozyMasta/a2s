package main

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/woozymasta/a2s/pkg/a2s"
	"github.com/woozymasta/a2s/pkg/a3sb"
)

func TestResponsiveTablesFitSupportedTerminalWidths(t *testing.T) {
	for _, terminalWidth := range []int{60, 80, 86, 100, 120, 160} {
		t.Run(fmt.Sprintf("width_%d", terminalWidth), func(t *testing.T) {
			tableCases := []struct {
				name   string
				render func(*bytes.Buffer, *Formatter) error
			}{
				{
					name: "info",
					render: func(output *bytes.Buffer, formatter *Formatter) error {
						app := NewApplication(output, output)
						return renderInfoTable(
							app,
							&a2s.Info{
								Format: a2s.InfoFormat(a2s.ResponseInfo),
								Name:   strings.Repeat("Long server name ", 12),
								Map:    strings.Repeat("custom_map_", 8),
							},
							a2s.QueryMeta{Duration: 57 * time.Millisecond},
							"127.0.0.1:27015",
							"table",
							formatter,
						)
					},
				},
				{
					name: "generic rules",
					render: func(output *bytes.Buffer, formatter *Formatter) error {
						app := NewApplication(output, output)
						return renderRulesTable(app, map[string]any{
							"hostname":                               strings.Repeat("rule value ", 10),
							"short":                                  "value",
							strings.Repeat("very_long_rule_key_", 8): "outlier",
						}, "127.0.0.1:27015", formatter)
					},
				},
				{
					name: "a3sb rules",
					render: func(output *bytes.Buffer, formatter *Formatter) error {
						app := NewApplication(output, output)
						rules := &a3sb.Rules{
							Version:     2,
							Island:      "deerisle",
							Description: strings.Repeat("Long DayZ server description ", 20),
						}
						for index := 0; index < 20; index++ {
							rules.Mods = append(rules.Mods, a3sb.Mod{
								Name: fmt.Sprintf("Mod %02d with a descriptive name", index+1),
								ID:   uint64(1000000000 + index),
							})
						}

						return renderA3SBRules(app, rules, formatter, "127.0.0.1:27015")
					},
				},
				{
					name: "players",
					render: func(output *bytes.Buffer, formatter *Formatter) error {
						app := NewApplication(output, output)
						players := make([]a2s.Player, 0, 20)
						for index := 0; index < 20; index++ {
							players = append(players, a2s.Player{
								Name:     fmt.Sprintf("Player %02d with an unusually long name", index+1),
								Score:    int32(index * 10),
								Index:    byte(index + 1),
								Duration: time.Duration(index+1) * time.Minute,
							})
						}

						return renderPlayersTable(app, players, "127.0.0.1:27015", formatter)
					},
				},
			}

			for _, tableCase := range tableCases {
				t.Run(tableCase.name, func(t *testing.T) {
					var output bytes.Buffer
					formatter := NewFormatter("table", &output)
					formatter.terminalWidth = terminalWidth

					if err := tableCase.render(&output, formatter); err != nil {
						t.Fatalf("render() error = %v", err)
					}

					assertRenderedTableFits(t, output.String(), terminalWidth)
				})
			}
		})
	}
}

func assertRenderedTableFits(t *testing.T, output string, terminalWidth int) {
	t.Helper()

	tableLines := 0
	for _, line := range strings.Split(output, "\n") {
		if !strings.HasPrefix(line, "╭") &&
			!strings.HasPrefix(line, "│") &&
			!strings.HasPrefix(line, "├") &&
			!strings.HasPrefix(line, "╰") {
			continue
		}

		tableLines++
		if width := text.StringWidthWithoutEscSequences(line); width > terminalWidth {
			t.Fatalf("table line width = %d, want <= %d: %q", width, terminalWidth, line)
		}
	}

	if tableLines == 0 {
		t.Fatal("output contains no rounded table lines")
	}
}

func TestA3SBLinkTablesUseCompactIDsOnlyWhenNeeded(t *testing.T) {
	rules := &a3sb.Rules{
		Version: 2,
		DLC: []a3sb.DLCInfo{
			{Name: strings.Repeat("Long DLC name ", 4), ID: 1151700},
		},
		CreatorDLC: []a3sb.DLCInfo{
			{Name: strings.Repeat("Long Creator DLC name ", 3), ID: 1175380},
		},
		Mods: []a3sb.Mod{
			{Name: strings.Repeat("Long mod name ", 3), ID: 123456789},
		},
	}

	for _, test := range []struct {
		name        string
		width       int
		wantCompact bool
	}{
		{name: "constrained", width: 86, wantCompact: true},
		{name: "wide", width: 160, wantCompact: false},
		{name: "non tty", width: 0, wantCompact: false},
	} {
		t.Run(test.name, func(t *testing.T) {
			var output bytes.Buffer
			formatter := NewFormatter("table", &output)
			formatter.terminalWidth = test.width
			app := NewApplication(&output, &output)

			if err := renderA3SBRules(app, rules, formatter, "127.0.0.1:27015"); err != nil {
				t.Fatalf("renderA3SBRules() error = %v", err)
			}

			text := output.String()
			compact := strings.Contains(text, "WORKSHOP ID") && strings.Contains(text, "APP ID")
			if compact != test.wantCompact {
				t.Fatalf("compact = %t, want %t; output = %q", compact, test.wantCompact, text)
			}

			fullModURL := "https://steamcommunity.com/sharedfiles/filedetails/?id=123456789"
			if test.wantCompact && strings.Contains(text, fullModURL) {
				t.Fatalf("constrained output contains full mod URL: %q", text)
			}
			if !test.wantCompact && !strings.Contains(text, fullModURL) {
				t.Fatalf("wide output does not contain full mod URL: %q", text)
			}
		})
	}
}
