package parser

import (
	"converter/converterUtils"
	"encoding/csv"
	"log"
	"os"
)

func ConversionTable(path string) (map[string]string, map[string]string) {

	file, err := os.Open(path)
	if err != nil {
		log.Fatal("Error while reading the file ", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatal("Error reading records", err)
	}
	// get the length of json map
	l := 0
	for i := range len(records) {
		if records[i][1] != "" {
			l++
		}
	}

	mapperProperty := make(map[string]string, 0)
	mapperUnits := make(map[string]string, 0)
	// in the table header extract column numbers which are under "OSCEM", "PDBx" and "units"
	var oscemFields int
	var indexPDB int
	var indexUnits int
	for j := range records[0][1:] {
		if records[0][1+j] != "" {
			oscemFields = 1 + j
			break
		}
	}
	for j := range records[0] {
		if records[0][j] == "in PDBx/mmCIF" {
			indexPDB = j
			break
		}
	}

	for j := range records[0] {
		if records[0][j] == "unitsExplicit" {
			indexUnits = j
			break
		}
	}
	for _, r := range records[1:] {
		if r[indexPDB] != "" {
			mapperProperty[converterUtils.StringJoiner(r[0:oscemFields])] = r[indexPDB]
			mapperUnits[converterUtils.StringJoiner(r[0:oscemFields])] = r[indexUnits]
		}
	}
	return mapperProperty, mapperUnits
}
