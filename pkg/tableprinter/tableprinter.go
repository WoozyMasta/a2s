package tableprinter

import (
	"fmt"
	"sort"
	"strings"
)

// print tables with column alignment
type TablePrinter struct {
	headers   []string
	rows      [][]string
	colWidths []int
	delimiter string
}

// create new table printer
func NewTablePrinter(headers []string, delimiter string) *TablePrinter {
	colWidths := make([]int, len(headers))
	for i, header := range headers {
		colWidths[i] = len(header)
	}

	return &TablePrinter{
		headers:   headers,
		colWidths: colWidths,
		delimiter: delimiter,
	}
}

// add a single row to table
func (tp *TablePrinter) AddRow(row []string) error {
	if len(row) != len(tp.headers) {
		return fmt.Errorf("row length does not match headers length")
	}

	for i, cell := range row {
		if len(cell) > tp.colWidths[i] {
			tp.colWidths[i] = len(cell)
		}
	}
	tp.rows = append(tp.rows, row)

	return nil
}

// add multiple rows at once
func (tp *TablePrinter) AddRows(rows [][]string) error {
	for _, row := range rows {
		if err := tp.AddRow(row); err != nil {
			return err
		}
	}

	return nil
}

// print table
func (tp *TablePrinter) Print() {
	tp.printTable(tp.rows)
}

// print table sorted by column
func (tp *TablePrinter) PrintSorted(col int) {
	if col < 0 || col >= len(tp.headers) {
		fmt.Printf("Invalid column index: %d\n", col)
		return
	}

	// copy and sort data
	rows := make([][]string, len(tp.rows))
	copy(rows, tp.rows)
	sort.Slice(rows, func(i, j int) bool {
		return rows[i][col] < rows[j][col]
	})

	// print sorted table
	tp.printTable(rows)
}

// internal printer for table header and rows
func (tp *TablePrinter) printTable(rows [][]string) {
	if len(tp.headers) == 0 {
		fmt.Printf("\nNot enough data to display ...\n")
		return
	}

	format := ""
	for _, width := range tp.colWidths {
		format += fmt.Sprintf("%%-%ds  ", width)
	}
	format = format[:len(format)-2] + "\n"

	delimiter := ""
	for _, width := range tp.colWidths {
		delimiter += strings.Repeat(tp.delimiter, width+2)
	}

	// print header and line delimiter
	fmt.Printf("\n"+format, convertToAny(tp.headers)...)
	fmt.Println(delimiter)

	// print table rows
	for _, row := range rows {
		for i, val := range row {
			row[i] = escapeSpecialChars(val)
		}
		fmt.Printf(format, convertToAny(row)...)
	}

	fmt.Println(delimiter)
}
