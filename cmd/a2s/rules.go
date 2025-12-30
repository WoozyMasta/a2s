package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/woozymasta/a2s/pkg/a2s"
	"github.com/woozymasta/a2s/pkg/a3sb"
	"github.com/woozymasta/steam/utils/appid"
)

// gameToAppID converts a game name string to AppID.
// Returns 0 if the game name is not recognized.
func gameToAppID(game string) uint64 {
	switch strings.ToLower(game) {
	case "arma3", "arma":
		return appid.Arma3.Uint64()
	case "dayz":
		return appid.DayZ.Uint64()
	default:
		return 0
	}
}

// isA3SBGame checks if the given AppID corresponds to Arma3 or DayZ.
func isA3SBGame(id uint64) bool {
	return id == appid.Arma3.Uint64() || id == appid.DayZ.Uint64() || id == appid.DayZExp.Uint64()
}

func executeRules(cmd *RulesCommand) {
	if cmd.Args.Host == "" {
		fatal("Host must be provided")
	}

	client := createClient(cmd.Args.Host, cmd.Args.Port, cmd.Timeout, cmd.Buffer)
	defer client.Close()

	formatter := NewFormatter(cmd.Format)

	// Determine if we should use a3sb parser
	useA3SB := false
	var appID uint64

	// Convert game string to AppID if specified
	if cmd.RulesOptions.Game != "" {
		appID = gameToAppID(cmd.RulesOptions.Game)
		if appID == 0 {
			fatalf("Unknown game: %s. Supported games: arma3, dayz", cmd.RulesOptions.Game)
		}
		useA3SB = true
	} else if !cmd.RulesOptions.SkipInfo && !cmd.RulesOptions.Raw {
		// If game not specified and skip-info is not set, try to detect from server info
		info, err := client.GetInfo()
		if err == nil {
			appID = info.ID
			if isA3SBGame(appID) {
				useA3SB = true
			}
		}
	}

	if useA3SB && !cmd.RulesOptions.Raw {
		executeRulesA3SB(client, appID, formatter)
	} else {
		executeRulesStandard(client, cmd.RulesOptions.Raw, formatter)
	}
}

func executeRulesStandard(client *a2s.Client, raw bool, formatter *Formatter) {
	var rules map[string]string
	var err error

	if raw {
		rules, err = client.GetRules()
	} else {
		parsedRules, err2 := client.GetParsedRules()
		if err2 != nil {
			fatalf("Failed to get rules: %s", err2)
		}

		rules = make(map[string]string)
		for k, v := range parsedRules {
			rules[k] = fmt.Sprint(v)
		}
	}

	if err != nil {
		fatalf("Failed to get rules: %s", err)
	}

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

	// Sort keys for better output
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
		fmt.Printf("A2S_RULES response for %s\n", client.Address)
	}
}

func executeRulesA3SB(client *a2s.Client, appID uint64, formatter *Formatter) {
	a3sbClient := &a3sb.Client{Client: client}

	rules, err := a3sbClient.GetRules(appID)
	if err != nil {
		fatalf("Failed to get server rules: %s", err)
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
		fmt.Printf("A2S_RULES response for %s\n", client.Address)
	}
}
