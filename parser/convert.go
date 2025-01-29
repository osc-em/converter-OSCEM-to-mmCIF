package parser

import (
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/osc-em/converter-OSCEM-to-mmCIF/converterUtils"
)

// GetValues returns slice of strings with values in maps. In those maps both Key and Value are strings.
func getValues[K string, V string](m map[string]string) []string {
	values := make([]string, 0)
	for _, v := range m {
		values = append(values, v)
	}
	return values
}
func PDBconvertFromFile(scientificMetadata map[string]any, metadataLevelNameInJson string, conversionFile string, dictFile string, mmCIFInput *os.File) (string, error) {
	mapper, PDBxdictvalues, jsonMeta, jsonUnits := parseInputs(scientificMetadata, metadataLevelNameInJson, conversionFile, dictFile)
	mmCIFText, err := SupplementCoordinatesFromFile(mapper, PDBxdictvalues, jsonMeta, jsonUnits, mmCIFInput)
	if err != nil {
		return "", err
	}
	return mmCIFText, nil
}
func PDBconvertFromPath(scientificMetadata map[string]any, metadataLevelNameInJson string, conversionFile string, dictFile string, mmCIFInputPath string) (string, error) {
	mapper, PDBxdictvalues, jsonMeta, jsonUnits := parseInputs(scientificMetadata, metadataLevelNameInJson, conversionFile, dictFile)
	fileName := strings.Split(mmCIFInputPath, ".")
	extension := fileName[len(fileName)-1]
	switch extension {
	case "gz":
		gzippedFile, err := os.Open(mmCIFInputPath)
		if err != nil {
			return "", err
		}
		defer gzippedFile.Close()
		gzipReader, err := gzip.NewReader(gzippedFile)
		if err != nil {
			return "", err
		}
		defer gzipReader.Close()

		// create a temporary cif file that will be used for a deposition
		uncompressed, err := os.CreateTemp("", "metadata.cif")
		if err != nil {
			return "", err
		}

		fileUncompressed := uncompressed.Name()
		_, err = io.Copy(uncompressed, gzipReader)
		if err != nil {
			return "", err
		}
		mmCIFText, err := SupplementCoordinatesFromPath(mapper, PDBxdictvalues, jsonMeta, jsonUnits, fileUncompressed)
		if err != nil {
			return "", err
		}
		defer os.Remove(fileUncompressed)
		return mmCIFText, nil
	case "bz2":
		return "", errors.New("not implemented")
	default:
		mmCIFText, err := SupplementCoordinatesFromPath(mapper, PDBxdictvalues, jsonMeta, jsonUnits, mmCIFInputPath)
		if err != nil {
			return "", err
		}
		return mmCIFText, nil
	}
}

func EMDBconvert(scientificMetadata map[string]any, metadataLevelNameInJson string, conversionFile string, dictFile string) (string, error) {
	mapper, PDBxdictvalues, jsonMeta, jsonUnits := parseInputs(scientificMetadata, metadataLevelNameInJson, conversionFile, dictFile)
	mmCIFText, err := CreteMetadataCif(mapper, PDBxdictvalues, jsonMeta, jsonUnits)
	if err != nil {
		return "", err
	}
	return mmCIFText, nil
}

func WriteCif(mmCIFText string, mmCIFOutputPath string) error {
	// Open the file, create it if it doesn't exist, and truncate it if it does
	file, err := os.OpenFile(mmCIFOutputPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0777)
	if err != nil {
		return err
	}
	defer file.Close() // Ensure the file is closed after the operation

	// Write the string to the file
	_, err = file.WriteString(mmCIFText)
	if err != nil {
		return err
	}
	fmt.Println("String successfully written to the file.")
	return nil
}

func parseInputs(scientificMetadata map[string]any, metadataLevelNameInJson string, conversionFile string, dictFile string) (map[string]string, map[string][]converterUtils.PDBxItem, map[string][]string, map[string][]string) {
	var dataItemsPerCategory map[string][]converterUtils.PDBxItem
	mapper := make(map[string]string, 0)
	// might be string or array of string depending on the size of json array
	mapJson := make(map[string][]string, 0)
	unitsOSCEM := make(map[string][]string, 0)

	FromJson(scientificMetadata, &mapJson, &unitsOSCEM, metadataLevelNameInJson)
	// read  conversion table by column:
	namesOSCEM, err := ConversionTableReadColumn(conversionFile, "OSCEM")
	if err != nil {
		log.Fatal(err)
		return mapper, dataItemsPerCategory, mapJson, unitsOSCEM
	}
	namesPDBx, err := ConversionTableReadColumn(conversionFile, "in PDBx/mmCIF")
	if err != nil {
		log.Fatal(err)
		return mapper, dataItemsPerCategory, mapJson, unitsOSCEM
	}
	// create a map containing OSCEM - PDBx naming mappings

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
		return mapper, dataItemsPerCategory, mapJson, unitsOSCEM

	}
	dataItemsPerCategory = AssignPDBxCategories((dataItems))
	return mapper, dataItemsPerCategory, mapJson, unitsOSCEM

}
