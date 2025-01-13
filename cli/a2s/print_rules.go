package main

import (
	"fmt"

	"github.com/woozymasta/a2s/internal/tableprinter"
	"github.com/woozymasta/a2s/pkg/a2s"
)

func printRules(client *a2s.Client, json, raw bool) {
	if raw {
		printRawRules(client, json)
	} else {
		printParsedRules(client, json)
	}
}

func printRawRules(client *a2s.Client, json bool) {
	rules, err := client.GetRules()
	if err != nil {
		fatalf("Failed to get rules: %s", err)
	}

	if json {
		printJSON(rules)
	}

	table := tableprinter.NewTablePrinter([]string{"Rule", "Value"}, "=")

	for k, v := range rules {
		if err := table.AddRow([]string{k, v}); err != nil {
			fatalf("Create rules table (Raw): %s", err)
		}
	}

	table.Print()
	fmt.Printf("A2S_RULES response for %s\n", client.Address)
}

func printParsedRules(client *a2s.Client, json bool) {
	rules, err := client.GetParsedRules()
	if err != nil {
		fatalf("Failed to get rules: %s", err)
	}

	if json {
		printJSON(rules)
	}

	table := tableprinter.NewTablePrinter([]string{"Rule", "Value"}, "=")

	for k, v := range rules {
		if err := table.AddRow([]string{k, fmt.Sprint(v)}); err != nil {
			fatalf("Create rules table (Parsed): %s", err)
		}
	}

	table.PrintSorted(0)
	fmt.Printf("A2S_RULES response for %s\n", client.Address)
}
