package main

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/woozymasta/a2s/pkg/a2s"
	"github.com/woozymasta/a2s/pkg/a3sb"
	"github.com/woozymasta/a2s/pkg/appid"
)

// gameToAppID converts a game name string to AppID.
// Returns 0 if the game name is not recognized.
func gameToAppID(game string) uint64 {
	switch strings.ToLower(game) {
	case "arma3", "arma":
		return appid.Arma3

	case "dayz":
		return appid.DayZ

	default:
		return 0
	}
}

// executeRules selects the standard or automatic A3SB rules parser and renders its output.
// Automatic mode uses one A2S_RULES request and does not require A2S_INFO merely to choose a parser.
func executeRules(cmd *RulesCommand) {
	if cmd.Args.Host == "" {
		fatal("Host must be provided")
	}

	client := createClient(cmd.Args.Host, cmd.Args.Port, cmd.Timeout, cmd.Buffer)
	defer closeClient(client)

	formatter := NewFormatter(cmd.Format)
	ctx := context.Background()

	if cmd.Game != "" {
		appID := gameToAppID(cmd.Game)
		if appID == 0 {
			fatalf("Unknown game: %s. Supported games: arma3, dayz", cmd.Game)
		}
		if cmd.Raw {
			executeRulesStandard(ctx, client, true, formatter)
			return
		}

		executeRulesA3SB(ctx, client, appID, formatter)
		return
	}

	if cmd.Raw {
		executeRulesStandard(ctx, client, cmd.Raw, formatter)
		return
	}

	executeRulesA3SB(ctx, client, 0, formatter)
}

// executeRulesStandard retrieves and renders ordinary A2S_RULES values.
func executeRulesStandard(ctx context.Context, client *a2s.Client, raw bool, formatter *Formatter) {
	var rules map[string]any
	var err error

	if raw {
		rawRules, rawErr := client.GetRules(ctx)
		err = rawErr
		rules = make(map[string]any, len(rawRules))
		for key, value := range rawRules {
			rules[key] = value
		}
	} else {
		rules, err = client.GetParsedRules(ctx)
	}

	if err != nil {
		fatalf("Failed to get rules: %s", err)
	}

	printRules(rules, client, formatter)
}

// printRules renders an already fetched rules map using the normal A2S output shape.
// Values remain typed for JSON and are formatted as text in tables.
func printRules(rules map[string]any, client *a2s.Client, formatter *Formatter) {
	if formatter.ShouldUseJSON() {
		formatter.PrintJSON(rules)
		return
	}

	t := table.NewWriter()
	if formatter.IsTableFormat() {
		t.SetOutputMirror(os.Stdout)
	}
	t.SetStyle(table.StyleRounded)
	t.AppendHeader(table.Row{"Rule", "Value"})

	// Sort keys so table and text output is deterministic.
	keys := make([]string, 0, len(rules))
	for k := range rules {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		t.AppendRow(table.Row{k, rules[k]})
	}

	formatter.PrintTable(t)
	if formatter.IsTableFormat() {
		fmt.Printf("A2S_RULES response for %s\n", client.Addr())
	}
}

// executeRulesA3SB retrieves and renders Arma 3/DayZ server-browser rules.
func executeRulesA3SB(ctx context.Context, client *a2s.Client, appID uint64, formatter *Formatter) {
	a3sbClient := &a3sb.Client{Client: client}

	rules, err := a3sbClient.GetRules(ctx, appID)
	if err != nil {
		fatalf("Failed to get server rules: %s", err)
	}

	if rules.Version == 0 {
		parsed := a2s.ParseRuleValues(rules.ExtraRules)
		printRules(parsed, client, formatter)
		return
	}

	if formatter.ShouldUseJSON() {
		formatter.PrintJSON(rules)
		return
	}

	// Print Island/Description info (DayZ specific)
	if rules.Island != "" {
		formatter.PrintSectionHeader("Server Information")
		t := table.NewWriter()
		if formatter.IsTableFormat() {
			t.SetOutputMirror(os.Stdout)
		}
		t.SetStyle(table.StyleRounded)
		t.AppendHeader(table.Row{"Option", "Value"})

		if rules.Description != "" {
			t.AppendRow(table.Row{"Description:", rules.Description})
		}

		t.AppendRows([]table.Row{
			{"Allowed build:", fmt.Sprintf("%d", rules.AllowedBuild)},
			{"Client port:", fmt.Sprintf("%d", rules.ClientPort)},
			{"Dedicated:", fmt.Sprintf("%t", rules.Dedicated)},
			{"Island:", rules.Island},
			{"Language:", rules.Language.String()},
			{"Platform:", rules.Platform},
			{"Required build:", fmt.Sprintf("%d", rules.RequiredBuild)},
			{"Required version:", fmt.Sprintf("%d", rules.RequiredVersion)},
			{"TimeLeft:", fmt.Sprintf("%d", rules.TimeLeft)},
		})

		formatter.PrintTable(t)
	}

	// Print Difficulty (Arma3 specific)
	if rules.Difficulty != nil {
		formatter.PrintSectionHeader("Difficulty Settings")
		t := table.NewWriter()
		if formatter.IsTableFormat() {
			t.SetOutputMirror(os.Stdout)
		}
		t.SetStyle(table.StyleRounded)
		t.AppendHeader(table.Row{"Option", "Value"})
		t.AppendRows([]table.Row{
			{"Difficulty Level:", fmt.Sprintf("%d", rules.Difficulty.Level)},
			{"AI Level:", fmt.Sprintf("%d", rules.Difficulty.AILevel)},
			{"Advanced Flight:", fmt.Sprintf("%t", rules.Difficulty.AdvanceFlight)},
			{"Third Person:", fmt.Sprintf("%t", rules.Difficulty.ThirdPerson)},
			{"Crosshair:", fmt.Sprintf("%t", rules.Difficulty.Crosshair)},
		})
		formatter.PrintTable(t)
	}

	// Print DLC
	if len(rules.DLC) > 0 {
		formatter.PrintSectionHeader("DLC")
		t := table.NewWriter()
		if formatter.IsTableFormat() {
			t.SetOutputMirror(os.Stdout)
		}
		t.SetStyle(table.StyleRounded)
		t.AppendHeader(table.Row{"#", "DLC Name", "DLC URL"})

		for i, dlc := range rules.DLC {
			t.AppendRow(table.Row{
				fmt.Sprintf("%d", i+1),
				dlc.Name,
				fmt.Sprintf("https://store.steampowered.com/app/%d", dlc.ID),
			})
		}

		formatter.PrintTable(t)
	}

	// Print Creator DLC
	if len(rules.CreatorDLC) > 0 {
		formatter.PrintSectionHeader("Creator DLC")
		t := table.NewWriter()
		if formatter.IsTableFormat() {
			t.SetOutputMirror(os.Stdout)
		}
		t.SetStyle(table.StyleRounded)
		t.AppendHeader(table.Row{"#", "Creator DLC Name", "Creator DLC URL"})

		for i, dlc := range rules.CreatorDLC {
			t.AppendRow(table.Row{
				fmt.Sprintf("%d", i+1),
				dlc.Name,
				fmt.Sprintf("https://store.steampowered.com/app/%d", dlc.ID),
			})
		}

		formatter.PrintTable(t)
	}

	// Print Mods
	if len(rules.Mods) > 0 {
		formatter.PrintSectionHeader("Mods")
		t := table.NewWriter()
		if formatter.IsTableFormat() {
			t.SetOutputMirror(os.Stdout)
		}
		t.SetStyle(table.StyleRounded)
		t.AppendHeader(table.Row{"#", "Mod Name", "Mod URL"})

		for i, mod := range rules.Mods {
			t.AppendRow(table.Row{
				fmt.Sprintf("%d", i+1),
				mod.Name,
				fmt.Sprintf("https://steamcommunity.com/sharedfiles/filedetails/?id=%d", mod.ID),
			})
		}

		formatter.PrintTable(t)
	}

	// Only print footer message for table format
	if formatter.IsTableFormat() {
		fmt.Printf("A2S_RULES response for %s\n", client.Addr())
	}
}
