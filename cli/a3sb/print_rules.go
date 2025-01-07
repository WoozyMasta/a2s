package main

import (
	"fmt"

	"github.com/woozymasta/a2s/pkg/a3sb"
	"github.com/woozymasta/a2s/pkg/tableprinter"
)

const (
	steamAppURL string = "https://store.steampowered.com/app/"
	steamModURL string = "https://steamcommunity.com/sharedfiles/filedetails/?id="
)

func printRules(rules *a3sb.Rules, address string, json bool) {
	if json {
		printJson(rules)
		return
	}

	if rules.Island != "" {
		table := makeRules(rules)
		table.Print()
	}

	if len(rules.DLC) > 0 {
		table := makeRulesDLC(rules)
		table.Print()
	}

	if len(rules.CreatorDLC) > 0 {
		table := makeRulesCreatorDLC(rules)
		table.Print()
	}

	if len(rules.Mods) > 0 {
		table := makeRulesMods(rules)
		table.Print()
	}

	fmt.Printf("A2S_RULES response for %s\n", address)
}

func makeRules(rules *a3sb.Rules) *tableprinter.TablePrinter {
	table := tableprinter.NewTablePrinter([]string{"Option", "Value"}, "=")

	if rules.Description != "" {
		table.AddRow([]string{"Description:", rules.Description})
	}

	table.AddRows([][]string{
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

	return table
}

func makeRulesDLC(rules *a3sb.Rules) *tableprinter.TablePrinter {
	table := tableprinter.NewTablePrinter([]string{"  #", "DLC Name", "DLC URL"}, "=")
	for i, dlc := range rules.DLC {
		table.AddRow([]string{fmt.Sprintf("%3d", i+1), dlc.Name, fmt.Sprintf("%s%d", steamAppURL, dlc.ID)})
	}

	return table
}

func makeRulesCreatorDLC(rules *a3sb.Rules) *tableprinter.TablePrinter {
	table := tableprinter.NewTablePrinter([]string{"  #", "Creator DLC Name", "Creator DLC URL"}, "=")
	for i, dlc := range rules.CreatorDLC {
		table.AddRow([]string{fmt.Sprintf("%3d", i+1), dlc.Name, fmt.Sprintf("%s%d", steamAppURL, dlc.ID)})
	}

	return table
}

func makeRulesMods(rules *a3sb.Rules) *tableprinter.TablePrinter {
	table := tableprinter.NewTablePrinter([]string{"  #", "Mod Name", "Mod URL"}, "=")
	for i, mod := range rules.Mods {
		table.AddRow([]string{fmt.Sprintf("%3d", i+1), mod.Name, fmt.Sprintf("%s%d", steamModURL, mod.ID)})
	}

	return table
}
