package parser

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"
)

// ConversionTableReadColumn return a slice of column values in a table.
// It takes a path to a CSV file and column name as an argument.
func ConversionTableReadColumn(path string, column string) ([]string, error) {

	columnValues := make([]string, 0)
	file, err := os.Open(path)
	if err != nil {
		return columnValues, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return columnValues, err
	}

	for j := range records[0] {
		// Since conversions table is often handled from Microsoft Excel and saved in UTF-8, it adds a byte-order mark (BOM) at the beginning of file, which is invisible.
		// for first column in header this needs to be removed from the name of column!
		if j == 0 {
			records[0][j] = strings.Trim(records[0][j], string('\uFEFF'))
		}
	}

	for _, r := range records[1:] {
		columnValues = append(columnValues, r[0])
	}
	fmt.Println("Column values:", columnValues)
	return columnValues, nil
}
