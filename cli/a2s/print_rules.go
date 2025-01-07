package main

import (
	"fmt"

	"github.com/woozymasta/a2s/pkg/tableprinter"
)

func printRules(rules map[string]string, address string, json bool) {
	if json {
		printJson(rules)
	} else {
		table := makeRules(rules)
		table.Print()
		fmt.Printf("A2S_RULES response for %s\n", address)
	}
}

func makeRules(rules map[string]string) *tableprinter.TablePrinter {
	table := tableprinter.NewTablePrinter([]string{"Rule", "Value"}, "=")

	for k, v := range rules {
		table.AddRow([]string{k, v})
	}

	return table
}

func printParsedRules(rules map[string]any, address string, json bool) {
	if json {
		printJson(rules)
	} else {
		table := makeParsedRules(rules)
		table.PrintSorted(0)
		fmt.Printf("A2S_RULES response for %s\n", address)
	}
}

func makeParsedRules(rules map[string]any) *tableprinter.TablePrinter {
	table := tableprinter.NewTablePrinter([]string{"Rule", "Value"}, "=")

	for k, v := range rules {
		table.AddRow([]string{k, fmt.Sprint(v)})
	}

	return table
}
