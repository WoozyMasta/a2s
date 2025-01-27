package main

import (
	"fmt"

	"github.com/woozymasta/a2s/internal/tableprinter"
	"github.com/woozymasta/a2s/pkg/a2s"
	"github.com/woozymasta/a2s/pkg/a3sb"
)

const (
	steamAppURL string = "https://store.steampowered.com/app/"
	steamModURL string = "https://steamcommunity.com/sharedfiles/filedetails/?id="
)

// processing A2S_RULES request
func printRules(clientA2S *a2s.Client, appID uint64, json bool) {
	client := &a3sb.Client{Client: clientA2S}

	rules, err := client.GetRules(appID)
	if err != nil {
		fatalf("Failed to get server rules: %s", err)
	}

	if json {
		printJSON(rules)
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

	fmt.Printf("A2S_RULES response for %s\n", client.Address)
}

func makeRules(rules *a3sb.Rules) *tableprinter.TablePrinter {
	table := tableprinter.NewTablePrinter([]string{"Option", "Value"}, "=")

	if rules.Description != "" {
		if err := table.AddRow([]string{"Description:", rules.Description}); err != nil {
			fatalf("Create rules table (Description): %s", err)
		}
	}

	if err := table.AddRows([][]string{
		{"Allowed build:", fmt.Sprintf("%d", rules.AllowedBuild)},
		{"Client port:", fmt.Sprintf("%d", rules.ClientPort)},
		{"Dedicated:", fmt.Sprintf("%t", rules.Dedicated)},
		{"Island:", rules.Island},
		{"Language:", rules.Language.String()},
		{"Platform:", rules.Platform},
		{"Required build:", fmt.Sprintf("%d", rules.RequiredBuild)},
		{"Required version:", fmt.Sprintf("%d", rules.RequiredVersion)},
		{"TimeLeft:", fmt.Sprintf("%d", rules.TimeLeft)},
	}); err != nil {
		fatalf("Create rules table: %s", err)
	}

	return table
}

func makeRulesDLC(rules *a3sb.Rules) *tableprinter.TablePrinter {
	table := tableprinter.NewTablePrinter([]string{"  #", "DLC Name", "DLC URL"}, "=")
	for i, dlc := range rules.DLC {
		if err := table.AddRow([]string{fmt.Sprintf("%3d", i+1), dlc.Name, fmt.Sprintf("%s%d", steamAppURL, dlc.ID)}); err != nil {
			fatalf("Create rules table (DLC): %s", err)
		}
	}

	return table
}

func makeRulesCreatorDLC(rules *a3sb.Rules) *tableprinter.TablePrinter {
	table := tableprinter.NewTablePrinter([]string{"  #", "Creator DLC Name", "Creator DLC URL"}, "=")
	for i, dlc := range rules.CreatorDLC {
		if err := table.AddRow([]string{fmt.Sprintf("%3d", i+1), dlc.Name, fmt.Sprintf("%s%d", steamAppURL, dlc.ID)}); err != nil {
			fatalf("Create rules table (Creator DLC): %s", err)
		}
	}

	return table
}

func makeRulesMods(rules *a3sb.Rules) *tableprinter.TablePrinter {
	table := tableprinter.NewTablePrinter([]string{"  #", "Mod Name", "Mod URL"}, "=")
	for i, mod := range rules.Mods {
		if err := table.AddRow([]string{fmt.Sprintf("%3d", i+1), mod.Name, fmt.Sprintf("%s%d", steamModURL, mod.ID)}); err != nil {
			fatalf("Create rules table (Mods): %s", err)
		}
	}

	return table
}
