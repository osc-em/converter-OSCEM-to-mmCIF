package main

import (
	"converter/parser"
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
	jsonInstr := os.Args[1]
	jsonSample := os.Args[2]

	conversionFile := os.Args[3]
	dictFile := os.Args[4]

	mmCIFpath := os.Args[5]
	//unitsPath := os.Args[6]

	// read all input json files and create one map from them all
	mapInstr := parser.FromJson(jsonInstr)
	mapSample := parser.FromJson(jsonSample)

	// create one mapping for all the JSON contents: key - value
	mapJson := make(map[string]any, len(mapInstr)+len(mapSample))
	for i := range mapInstr {
		mapJson[i] = mapInstr[i]
	}
	for i := range mapSample {
		mapJson[i] = mapSample[i]
	}

	// read the conversion table by column:
	namesOSCEM, err := parser.ConversionTableReadColumn(conversionFile, "OSCEM")
	if err != nil {
		log.Fatal(err)
		return
	}
	namesPDBx, err := parser.ConversionTableReadColumn(conversionFile, "in PDBx/mmCIF")
	if err != nil {
		log.Fatal(err)
		return
	}
	units, err := parser.ConversionTableReadColumn(conversionFile, "unitsExplicit")
	if err != nil {
		log.Fatal(err)
		return
	}
	mapper := make(map[string]string, 0)
	unitsOSCEM := make(map[string]string, 0)
	// create a map containing OSCEM - PDBx naming mappings
	for i := range namesOSCEM {
		// skip values that have notation in OSCEM but not in PDBx
		if namesPDBx[i] != "" {
			mapper[namesOSCEM[i]] = namesPDBx[i]
			unitsOSCEM[namesOSCEM[i]] = units[i]
		}
	}

	// parse PDBx dictionary to retrieve order of data items and units
	dataItems := parser.PDBxDict(dictFile, getValues(mapper))

	// use only a map of dataItems that will be needed my mapper
	//var PDBxdataItems = make(map[string][]converterUtils.PDBxItem)

	// for k, v := range dataItems {
	// 	for _, mV := range mapper {
	// 		if "_"+k == strings.Split(mV, ".")[0] {
	// 			PDBxdataItems[k] = v
	// 			break
	// 		}
	// 	}
	// }

	// create mmCIF text
	mmCIFlines := parser.ToMmCIF(mapper, dataItems, mapJson, unitsOSCEM)

	// now write to cif file

	// Open the file, create it if it doesn't exist, and truncate it if it does
	file, err := os.OpenFile(mmCIFpath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatal("Error opening file: ", err)
		return
	}
	defer file.Close() // Ensure the file is closed after the operation

	// Write the string to the file
	_, err = file.WriteString(mmCIFlines)
	if err != nil {
		log.Fatal("Error writing to file: ", err)
		return
	}

	fmt.Println("String successfully written to the file.")

	// // Open the file, create it if it doesn't exist, and truncate it if it does
	// fileUnits, err := os.OpenFile(unitsPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	// if err != nil {
	// 	log.Fatal("Error opening file: ", err)
	// 	return
	// }
	// defer file.Close() // Ensure the file is closed after the operation

	// //var unitsString = ""
	// var unitsString strings.Builder
	// unitsString.WriteString("")
	// for k, v := range unitsOSCEM {
	// 	fmt.Fprintf(&unitsString, "%s,%s\n", k, v)
	// }
	// // Write the string to the file
	// _, err = fileUnits.WriteString(unitsString.String())
	// if err != nil {
	// 	log.Fatal("Error writing to file: ", err)
	// 	return
	// }

}

// go run converter.go data/data_instrument.json data/data_sample.json /Users/sofya/Documents/openem/LS_Metadata_reader/conversion/conversions.csv ./data/mmcif_pdbx_v50.dic /Users/sofya/Documents/openem/converter-JSON-to-mmCIF/results/output.cif /Users/sofya/Documents/openem/converter-JSON-to-mmCIF/results/units.csv
