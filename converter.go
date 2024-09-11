package main

import (
	"converter/parser"
	"flag"
	"fmt"
	"log"
	"os"
)

// GetValues returns slice of strings with values in maps. In those maps both Key and Value are strings.
func getValues[K string, V string](m map[string]string) []string {
	values := make([]string, 0)
	for _, v := range m {
		values = append(values, v)
	}
	return values
}

func main() {

	//var json string
	appendToMmCif := flag.Bool("append", true, "append metadata to existing mmCIF")
	mmCIFInputPath := flag.String("mmCIFfile", "", "path to existing mmCIF file with atoms information")
	conversionFile := flag.String("conversions", "", "path to a CSV file with conversions table")
	dictFile := flag.String("dic", "", "path to PDBx dictionary")
	mmCIFOutputPath := flag.String("output", "", "path for a new mmCIF file")
	metadataLevelNameInJson := flag.String("level", "scientificMetadata", "Name of JSON key, for which element the conversion will take place. It must be at first level")

	// Parse JSON files
	json := flag.String("json", "", "JSON file containing metadata")
	flag.Parse()

	// might be string or array of string depending on the size of json array
	mapJson := make(map[string]any, 0)
	unitsOSCEM := make(map[string]any, 0)

	parser.FromJson(*json, &mapJson, &unitsOSCEM, *metadataLevelNameInJson)

	// read  conversion table by column:
	namesOSCEM, err := parser.ConversionTableReadColumn(*conversionFile, "OSCEM")
	if err != nil {
		log.Fatal(err)
		return
	}
	namesPDBx, err := parser.ConversionTableReadColumn(*conversionFile, "in PDBx/mmCIF")
	if err != nil {
		log.Fatal(err)
		return
	}

	// create a map containing OSCEM - PDBx naming mappings
	mapper := make(map[string]string, 0)
	for i := range namesOSCEM {
		// skip values that have notation in OSCEM but not in PDBx
		if namesPDBx[i] != "" {
			mapper[namesOSCEM[i]] = namesPDBx[i]
		}
	}

	// parse PDBx dictionary to retrieve order of data items and units
	dataItems, err := parser.PDBxDict(*dictFile, getValues(mapper))
	if err != nil {
		log.Fatal("Error while reading PDBx dictionary: ", err)
		return

	}
	dataItemsPerCategory := parser.AssignPDBxCategories((dataItems))

	// create mmCIF text and write it to a file
	mmCIFText := parser.ToMmCIF(mapper, dataItemsPerCategory, mapJson, unitsOSCEM, *appendToMmCif, *mmCIFInputPath)

	// now write to cif file
	// Open the file, create it if it doesn't exist, and truncate it if it does
	file, err := os.OpenFile(*mmCIFOutputPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatal("Error opening file: ", err)
		return
	}
	defer file.Close() // Ensure the file is closed after the operation

	// Write the string to the file
	_, err = file.WriteString(mmCIFText)
	if err != nil {
		log.Fatal("Error writing to file: ", err)
		return
	}
	fmt.Println("String successfully written to the file.")

}

// go run converter.go -append=false -conversions /Users/sofya/Documents/openem/LS_Metadata_reader/conversion/conversions.csv -dic ./data/mmcif_pdbx_v50.dic -output /Users/sofya/Documents/openem/converter-JSON-to-mmCIF/results/output.cif -instrument data/data_instrument.json -sample data/data_sample.json

// go run converter.go -append=true -mmCIFfile /Users/sofya/Documents/openem/converter-JSON-to-mmCIF/data/K3DAK4_full__real_space_refined_000.cif -conversions /Users/sofya/Documents/openem/LS_Metadata_reader/conversion/conversions.csv -dic ./data/mmcif_pdbx_v50.dic -output /Users/sofya/Documents/openem/converter-JSON-to-mmCIF/results/outputAppended.cif -instrument data/data_instrument.json -sample data/data_sample.json
