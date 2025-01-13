package tableprinter

import (
	"fmt"
	"sort"
	"strings"
)

// TablePrinter is used to create and print tables with aligned columns.
// It maintains headers, rows, column widths, and a delimiter for separating columns.
type TablePrinter struct {
	delimiter string
	headers   []string
	rows      [][]string
	colWidths []int
}

// NewTablePrinter creates and initializes a new TablePrinter with the provided headers and delimiter.
//
// Parameters:
//   - headers: A slice of strings representing the column headers.
//   - delimiter: A string used to separate columns when printing the table.
//
// Returns:
//   - A pointer to an initialized TablePrinter.
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

// AddRow adds a single row to the table.
//
// Parameters:
//   - row: A slice of strings representing the row data.
//
// Returns:
//   - An error if the length of the row does not match the number of headers.
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

// AddRows adds multiple rows to the table at once.
//
// Parameters:
//   - rows: A slice of slices of strings, where each inner slice represents a row.
//
// Returns:
//   - An error if any of the rows have a length that does not match the number of headers.
func (tp *TablePrinter) AddRows(rows [][]string) error {
	for _, row := range rows {
		if err := tp.AddRow(row); err != nil {
			return err
		}
	}

	return nil
}

// Print displays the table in the console with proper column alignment.
//
// It prints the headers, a delimiter line, all added rows, and a closing delimiter line.
func (tp *TablePrinter) Print() {
	tp.printTable(tp.rows)
}

// PrintSorted displays the table sorted by the specified column index.
//
// Parameters:
//   - col: The zero-based index of the column to sort by.
//
// If the column index is out of range, it prints an error message and does not display the table.
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

// printTable is an internal method that handles the actual printing of the table.
//
// Parameters:
//   - rows: A slice of slices of strings representing the rows to be printed.
func (tp *TablePrinter) printTable(rows [][]string) {
	if len(tp.headers) == 0 {
		fmt.Printf("\nNot enough data to display ...\n")
		return
	}

	// Construct the format string based on column widths
	format := ""
	for _, width := range tp.colWidths {
		format += fmt.Sprintf("%%-%ds  ", width)
	}
	format = format[:len(format)-2] + "\n" // Remove the last two spaces and add a newline

	// Create the delimiter line
	delimiter := ""
	for _, width := range tp.colWidths {
		delimiter += strings.Repeat(tp.delimiter, width+2)
	}

	// Print header and delimiter
	fmt.Printf("\n"+format, convertToAny(tp.headers)...)
	fmt.Println(delimiter)

	// Print each row with escaped special characters
	for _, row := range rows {
		for i, val := range row {
			row[i] = escapeSpecialChars(val)
		}
		fmt.Printf(format, convertToAny(row)...)
	}

	fmt.Println(delimiter)
}
