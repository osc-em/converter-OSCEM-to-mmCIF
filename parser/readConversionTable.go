package parser

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"
)

func stringJoiner(stringsArray []string) string {
	var joinedString string
	for _, s := range stringsArray {
		if s != "" {
			if joinedString != "" {
				joinedString += "."
			}
			joinedString += strings.TrimSpace(s) // trim all leading and trailing whitespaces
		}
	}
	return joinedString
}

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

	// in the table header extract the index of the column
	var fieldsStart int
	var fieldsEnd int
	var colExists bool

	for j := range records[0] {
		// Since conversions table is often handled from Microsoft Excel and saved in UTF-8, it adds a byte-order mark (BOM) at the beginning of file, which is invisible.
		// for first column in header this needs to be removed from the name of column!
		if j == 0 {
			records[0][j] = strings.Trim(records[0][j], string('\uFEFF'))
		}
		if records[0][j] == column {
			fieldsStart = j
			colExists = true
			break
		}
	}
	for j := range records[0][fieldsStart+1:] {
		if records[0][fieldsStart+1+j] != "" {
			fieldsEnd = fieldsStart + 1 + j
			break
		}
	}

	if !colExists {

		return columnValues, fmt.Errorf("column %s does not exist in table %s", column, path)
	}

	for _, r := range records[1:] {
		columnValues = append(columnValues, stringJoiner(r[fieldsStart:fieldsEnd]))
	}
	return columnValues, nil
}
