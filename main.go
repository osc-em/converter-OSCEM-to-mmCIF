package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/osc-em/converter-OSCEM-to-mmCIF/parser"
)

func main() {
	appendToMmCif := flag.Bool("append", true, "append metadata to existing mmCIF")
	mmCIFInputPath := flag.String("mmCIFfile", "", "path to existing mmCIF file with atoms information")
	conversionFile := flag.String("conversions", "", "path to a CSV file with conversions table")
	dictFile := flag.String("dic", "", "path to PDBx dictionary")
	mmCIFOutputPath := flag.String("output", "", "path for a new mmCIF file")
	metadataLevelNameInJson := flag.String("level", "scientificMetadata", "Name of JSON key, for which element the conversion will take place. It must be at first level")
	// Parse JSON files
	jsonPath := flag.String("json", "", "JSON file containing metadata")

	flag.Parse()

	// Read JSON file
	data, err := os.ReadFile(*jsonPath)
	if err != nil {
		errorText := fmt.Errorf("error while reading the JSON file: %w", err)
		log.Fatal(errorText)
		return
	}

	// Unmarshal JSON
	var jsonContent map[string]any
	if err := json.Unmarshal(data, &jsonContent); err != nil {
		errorText := fmt.Errorf("error while unmarshaling JSON: %w", err)
		log.Fatal(errorText)
		return
	}
	var mmCIFText string
	if *appendToMmCif {
		mmCIFText, err = parser.PDBconvertFromPath(jsonContent, *metadataLevelNameInJson, *conversionFile, *dictFile, *mmCIFInputPath)
		if err != nil {
			log.Fatal(err.Error())
		}
	} else {
		mmCIFText, err = parser.EMDBconvert(jsonContent, *metadataLevelNameInJson, *conversionFile, *dictFile)
		if err != nil {
			log.Fatal(err.Error())
		}
	}
	parser.WriteCif(mmCIFText, *mmCIFOutputPath)
}
