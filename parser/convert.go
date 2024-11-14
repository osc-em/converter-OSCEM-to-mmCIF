package parser

import (
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

func Convert(scientificMetadata map[string]any, metadataLevelNameInJson string, conversionFile string, dictFile string, appendToMmCif bool, mmCIFInputPath string, mmCIFOutputPath string) {
	// might be string or array of string depending on the size of json array
	mapJson := make(map[string][]string, 0)
	unitsOSCEM := make(map[string][]string, 0)

	FromJson(scientificMetadata, &mapJson, &unitsOSCEM, metadataLevelNameInJson)
	// read  conversion table by column:
	namesOSCEM, err := ConversionTableReadColumn(conversionFile, "OSCEM")
	if err != nil {
		log.Fatal(err)
		return
	}
	namesPDBx, err := ConversionTableReadColumn(conversionFile, "in PDBx/mmCIF")
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
	dataItems, err := PDBxDict(dictFile, getValues(mapper))
	if err != nil {
		log.Fatal("Error while reading PDBx dictionary: ", err)
		return

	}
	dataItemsPerCategory := AssignPDBxCategories((dataItems))
	// create mmCIF text and write it to a file
	mmCIFText, err := ToMmCIF(mapper, dataItemsPerCategory, mapJson, unitsOSCEM, appendToMmCif, mmCIFInputPath)
	if err != nil {
		fmt.Println("Couldn't create text in mmCIF format!", err)
	}

	// now write to cif file
	// Open the file, create it if it doesn't exist, and truncate it if it does
	file, err := os.OpenFile(mmCIFOutputPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0777)
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
