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

// func OSCEMConversionTable(path string) (map[string]string, map[string]string) {
// 	file, err := os.Open(path)
// 	if err != nil {
// 		log.Fatal("Error while reading the file ", err)
// 	}
// 	defer file.Close()

// 	reader := csv.NewReader(file)
// 	records, err := reader.ReadAll()
// 	if err != nil {
// 		log.Fatal("Error reading records", err)
// 	}

// 	mapperProperty := make(map[string]string, 0)
// 	mapperUnits := make(map[string]string, 0)
// 	// in the table header extract column numbers which are under "OSCEM", "PDBx" and "units"
// 	var oscemFields int
// 	var indexPDB int
// 	var indexUnits int
// 	for j := range records[0][1:] {
// 		if records[0][1+j] != "" {
// 			oscemFields = 1 + j
// 			break
// 		}
// 	}
// 	for j := range records[0] {
// 		if records[0][j] == "in PDBx/mmCIF" {
// 			indexPDB = j
// 			break
// 		}
// 	}

// 	for j := range records[0] {
// 		if records[0][j] == "unitsExplicit" {
// 			indexUnits = j
// 			break
// 		}
// 	}
// 	for _, r := range records[1:] {
// 		if r[indexPDB] != "" {
// 			mapperProperty[stringJoiner(r[0:oscemFields])] = r[indexPDB]
// 			mapperUnits[stringJoiner(r[0:oscemFields])] = r[indexUnits]
// 		}
// 	}
// 	return mapperProperty, mapperUnits
// }

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
