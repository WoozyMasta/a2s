/*
Package tableprinter provides utilities for printing well-formatted tables with aligned columns in the console.

It allows you to create tables with headers, add single or multiple rows, and print them with customizable delimiters.
Additionally, it supports sorting the table based on a specified column.

Example usage:

	package main

	import (
		"log"

		"github.com/woozymasta/a2s/pkg/tableprinter"
	)

	func main() {
		headers := []string{"Name", "Age", "Occupation"}
		tp := tableprinter.NewTablePrinter(headers, "|")

		err := tp.AddRow([]string{"Alice", "30", "Engineer"})
		if err != nil {
			log.Fatalf("Error adding row: %v", err)
		}

		err = tp.AddRows([][]string{
			{"Bob", "25", "Designer"},
			{"Charlie", "35", "Teacher"},
		})
		if err != nil {
			log.Fatalf("Error adding rows: %v", err)
		}

		// Print the table as is
		tp.Print()

		// Print the table sorted by the second column (Age)
		tp.PrintSorted(1)
	}
*/
package tableprinter
