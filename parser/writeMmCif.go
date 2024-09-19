package parser

import (
	"bufio"
	"converter/converterUtils"
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func getKeyByValue(value string, m map[string]string) (string, error) {
	for k, v := range m {
		if v == value {
			return k, nil
		}
	}
	return "", fmt.Errorf("value %v is not in the conversion table", value)
}

// is element e in the slice s?
func sliceContains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// given a slice of PDBx items get the length of a longest data item name in it (because the category is the same)
func getLongestPDBxItem(s []converterUtils.PDBxItem) int {
	var l int
	for i := range s {
		if len(s[i].Name) > l {
			l = len(s[i].Name)
		}
	}
	return l + len(s[0].CategoryID) + 1
}

func validateDateIsRFC3339(date string) string {
	t, err := time.Parse(time.RFC3339, date)
	if err != nil {
		fmt.Println(err)
		log.Printf("Date seems to be in the wrong format! We expect RFC3339 format, provided date %s is not.", date)
		return ""
	}
	return t.Format(time.DateOnly)
}

func validateRange(value string, dataItem converterUtils.PDBxItem, unitOSCEM string, nameOSCEM string) (bool, error) {
	rMin := dataItem.RangeMin
	rMax := dataItem.RangeMax
	unitPDBx := dataItem.Unit
	namePDBx := dataItem.CategoryID + "." + dataItem.Name
	var unitsSame bool
	var errorMessage string
	var unitsError error
	if unitOSCEM == "" && unitPDBx == "" {
		// both OSCEM and PDBx have no units definition
		unitsSame = true
	} else if unitOSCEM == "" && unitPDBx != "" {
		errorMessage = fmt.Sprintf("No units defined for %s in OSCEM! Analogous property %s in PDBx has %s units. Value will still be used in mmCIF file!", nameOSCEM, namePDBx, unitPDBx)
		unitsSame = false
		unitsError = errors.New(errorMessage)
		return true, unitsError
	} else if unitOSCEM != "" && unitPDBx == "" {
		errorMessage = fmt.Sprintf("No units defined for %s in PDBx! Analogous property %s in OSCEM has %s units. Value will still be used in mmCIF file!", namePDBx, nameOSCEM, unitOSCEM)
		unitsSame = false
		unitsError = errors.New(errorMessage)
		return true, unitsError
	} else {
		explicitUnitOSCEM, ok := converterUtils.UnitsName[unitOSCEM]
		if !ok {
			errorMessage = fmt.Sprintf("No explicit unit name is specified for property %s in OSCEM, only a short name %s. Value will still be used in mmCIF file!", nameOSCEM, unitOSCEM)
			unitsSame = false
			unitsError = errors.New(errorMessage)
			return true, unitsError
		} else {
			unitsError = nil
		}
		unitsSame = explicitUnitOSCEM == unitPDBx
	}
	// FIXME: when units are settled, do conversion when possible!
	if unitsSame {
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			errorMessage = fmt.Sprintf("JSON value %s not numeric, but supposed to be", value)
			return false, errors.New(errorMessage)
		}
		rMin, err := strconv.ParseFloat(rMin, 64)
		if err != nil {
			rMin = math.NaN()
		}
		rMax, err := strconv.ParseFloat(rMax, 64)
		if err != nil {
			rMax = math.NaN()
		}
		if math.IsNaN(rMin) && math.IsNaN(float64(rMax)) {
			return true, unitsError
		} else if math.IsNaN(rMin) {
			return float64(v) <= rMax, unitsError
		} else if math.IsNaN(rMax) {
			return float64(v) >= rMin, unitsError
		} else {
			return float64(v) >= rMin && float64(v) <= rMax, unitsError
		}
	} else {
		errorMessage = fmt.Sprintf("Units for analogous properties %s in OSCEM and %s in PDBx  don't match! Implement a converter from %s in OSCEM to %s expected by PDBx. Value will still be used in mmCIF file!", nameOSCEM, namePDBx, unitOSCEM, unitPDBx)
		unitsError = errors.New(errorMessage)
		return true, unitsError
	}

}
func validateEnum(value string, dataItem converterUtils.PDBxItem) string {
	enumFromEMDB := dataItem.EnumValues
	enumFromPDBx := dataItem.PDBxEnumValues
	namePDBx := dataItem.CategoryID + "." + dataItem.Name
	if value == "true" {
		return "YES"
	} else if value == "false" {
		return "NO"
	}

	if namePDBx == "_em_imaging.microscope_model" {
		reTitan := regexp.MustCompile(`(?i)titan`)
		if reTitan.MatchString(value) {
			return "TFS KRIOS"
		}
	} else if namePDBx == "_em_imaging.mode" {
		if value == "BrightField" {
			return "BRIGHT FIELD"
		}
	} else if namePDBx == "_em_imaging.electron_source" {
		if value == "FieldEmission" {
			return "FIELD EMISSION GUN"
		}
	}
	for i := range enumFromEMDB {
		if strings.EqualFold(enumFromEMDB[i], value) {
			value = enumFromEMDB[i]
			return value
		}
	}
	// scan through both enums
	for i := range enumFromPDBx {
		if strings.EqualFold(enumFromPDBx[i], value) {
			value = enumFromPDBx[i]
			return value
		}
	}
	// add additional matching mechanism for grid material by checmical element name/ regular expression
	if namePDBx == "_em_sample_support.grid_material" {

		reGraphene := regexp.MustCompile(`(?i)graphene`)
		reSilicon := regexp.MustCompile(`(?i)silicon`)
		if reGraphene.MatchString(value) {
			return "GRAPHENE OXIDE"
		} else if reSilicon.MatchString(value) {
			return "SILICON NITRIDE"
		}
		switch value {
		case "Cu":
			return "COPPER"
		case "Cu/Pd":
			return "COPPER/PALLADIUM"
		case "Cu/Rh":
			return "COPPER/RHODIUM"
		case "Au":
			return "GOLD"
		case "Ni":
			return "NICKEL"
		case "Ni/Ti":
			return "NICKEL/TITANIUM"
		case "Pt":
			return "PLATINUM"
		case "W":
			return "TUNGSTEN"
		case "Ti":
			return "TITANIUM"
		case "Mo":
			return "MOLYBDENUM"
		}

	} else if namePDBx == "_em_image_recording.film_or_detector_model" {
		// add additional matching mechanism for "Falcon" detector model by regular expression; other don't seem feasible

		reFalconI := regexp.MustCompile(`(?i)falcon[\s_]*?(1|I)`)
		reFalconII := regexp.MustCompile(`(?i)falcon[\s_]*?(2|II)`)
		reFalconIII := regexp.MustCompile(`(?i)falcon[\s_]*?(3|III)`)
		reFalconIV := regexp.MustCompile(`(?i)falcon[\s_]*?(4|IV)`)
		switch {
		case reFalconIV.MatchString(value):
			return "FEI FALCON IV (4k x 4k)"
		case reFalconIII.MatchString(value):
			return "FEI FALCON III (4k x 4k)"
		case reFalconII.MatchString(value):
			return "FEI FALCON II (4k x 4k)"
		case reFalconI.MatchString(value):
			return "FEI FALCON I (4k x 4k)"
		}
	}

	log.Printf("value %v is not in enum %s!", value, namePDBx)
	// if not found in enum list and it's a funding organisation, put a certain string
	if namePDBx == "_pdbx_audit_support.funding_organization" {
		return "Other government"
	}

	return strings.ToUpper(value)
}

func checkValue(dataItem converterUtils.PDBxItem, value string, jsonKey string, unitsOSCEM string) string {

	//now based on the found struct implement range matching, units matching and enum matching
	if dataItem.ValueType == "int" || dataItem.ValueType == "float" {
		namePDBx := dataItem.CategoryID + "." + dataItem.Name
		// in OSCEM defocus is negative value and overfocus is positive. In PDBx it's vice versa if string starts with  minus, ut it off, othersise add a - prefix
		if namePDBx == "_em_imaging.nominal_defocus_min" || namePDBx == "_em_imaging.calibrated_defocus_min" || namePDBx == "_em_imaging.nominal_defocus_max" || namePDBx == "_em_imaging.calibrated_defocus_max" {
			if value[0] == 45 {
				value = value[1:]
			} else {
				value = "-" + value
			}
		}
		validatedRange, err := validateRange(value, dataItem, unitsOSCEM, jsonKey)
		if err != nil {
			errorNumeric := fmt.Sprintf("JSON value %s not numeric, but supposed to be", value)
			if err.Error() == errorNumeric {
				return "?"
			} else {
				// FIXME when units convertion is implemented, handle this error
				log.Println(err.Error())
			}
		}
		if !validatedRange {
			log.Printf("Value %s of property %s is not in range of [ %s, %s ]!\n", value, jsonKey, dataItem.RangeMin, dataItem.RangeMax)
		}
	} else if dataItem.ValueType == "yyyy-mm-dd" {
		value = validateDateIsRFC3339(value)
	} else if len(dataItem.EnumValues) > 0 || len(dataItem.PDBxEnumValues) > 0 {
		value = validateEnum(value, dataItem)
	}

	if strings.Contains(value, " ") {
		value = fmt.Sprintf("'%s' ", value) // if name contains whitespaces enclose it in single quotes
	} else {
		value = fmt.Sprintf("%s ", value) // take value as is
	}
	return value
}

func getOrderCategories(parsedCategories []string) []string {
	var order []string
	var processed []string
	// sort based on the pre-defined (administrative, polymer related entities, ligand (non-polymer) related instances, and structure level description)
	for _, category := range converterUtils.PDBxCategoriesOrder {
		category = "_" + category
		if sliceContains(parsedCategories, category) {
			order = append(order, category)
			processed = append(processed, category)
		}
	}
	// add the rest not atom-related in some order (it can be random)
	for _, c := range parsedCategories {
		if !sliceContains(processed, c) && !sliceContains(converterUtils.PDBxCategoriesOrderAtom, c) {
			order = append(order, c)
		}
	}
	// add atoms categories
	for _, category := range converterUtils.PDBxCategoriesOrderAtom {
		category = "_" + category
		order = append(order, category)
	}
	return order
}

func parseMmCIF(path string) (string, map[string]string) {
	dictFile, err := os.Open(path)
	if err != nil {
		log.Fatal("Error while reading the file ", err)
	}
	defer dictFile.Close()
	scanner := bufio.NewScanner(dictFile)

	var dataName string
	var category string
	var mmCIFfields = make(map[string]string, 0)

	var str strings.Builder

	l := 0
	inCategoryFlag := true
	for scanner.Scan() {
		// first line is the name
		if l == 0 {
			dataName = scanner.Text()
		}
		// break between categories is denoted by # in PDB-related software, Phenix uses an empty line.
		if strings.HasPrefix(scanner.Text(), "#") || len(strings.Fields(scanner.Text())) == 0 {

			//category ends, appends to the map
			if category != "" {
				mmCIFfields[category] = str.String()
				str.Reset()
				inCategoryFlag = true // record the next category name
			}
		} else {
			if !strings.HasPrefix(scanner.Text(), "loop_") {

				if inCategoryFlag {
					category = strings.Split(strings.Fields(scanner.Text())[0], ".")[0]
					inCategoryFlag = false
				}
			}
			str.WriteString(scanner.Text() + "\n")

		}
		l++
	}
	return dataName, mmCIFfields
}
func ToMmCIF(nameMapper map[string]string, PDBxItems map[string][]converterUtils.PDBxItem, valuesMap map[string][]string, OSCEMunits map[string][]string, appendToMmCif bool, mmCIFpath string) (string, error) {
	var dataName string
	var mmCIFCategories map[string]string
	if appendToMmCif {
		dataName, mmCIFCategories = parseMmCIF(mmCIFpath)
	} else {
		dataName = "data_myID"
	}

	// keeps track of values from JSON that have already been mapped to the PDBx properties
	var str strings.Builder
	str.WriteString(dataName + "\n#\n") //write the data Identifier in the header

	parsedCategories := make([]string, 0)
	for k := range PDBxItems {
		parsedCategories = append(parsedCategories, k)
	}
	allCategories := getOrderCategories(parsedCategories)
	for _, category := range allCategories {
		catDI, ok := PDBxItems[category]
		if ok {
			_, ok := mmCIFCategories[category]
			if ok {
				log.Printf("Category %s exists both in metadata from JSON files and in existing mmCIF file! Data in mmCIF will be substituted by data from JSON", category)
			}
			// this category exists in PDBx items (as extracted only to relevant OSCEM names) --> extract this input
			var key string
			// need to loop here through all data items in category, as it reflects the order of data items in PDBx, it still might not exist in json
			// loop until we find first key that exists in json
			for i := range catDI {
				k, err := getKeyByValue(catDI[i].CategoryID+"."+catDI[i].Name, nameMapper)
				key = k
				if err != nil {
					return "", err
				}
				_, ok := valuesMap[key]
				if ok {
					break
				}
			}
			size := len(valuesMap[key])
			switch {
			case size > 1:
				// loop notation
				str.WriteString("loop_\n")
				for _, dataItem := range catDI {
					// check the length of all first and throw an error in case that they have different length?? can that be? e.g two authors and for one the property Phone is not there?
					jsonKey, err := getKeyByValue(dataItem.CategoryID+"."+dataItem.Name, nameMapper)
					if err != nil {
						return "", err
					}
					if valuesMap[jsonKey] == nil {
						continue // not required and not provided in OSCEM
					}
					fmt.Fprintf(&str, "%s\n", dataItem.CategoryID+"."+dataItem.Name)
				}
				for v := range valuesMap[key] {
					for _, dataItem := range catDI {
						jsonKey, err := getKeyByValue(dataItem.CategoryID+"."+dataItem.Name, nameMapper)
						if err != nil {
							return "", err
						}

						if valuesMap[jsonKey] == nil {
							continue // key was not required and not provided in OSCEM
						}
						if correctSlice, ok := valuesMap[jsonKey]; ok {
							if unit, ok := OSCEMunits[jsonKey]; ok {
								valueString := checkValue(dataItem, correctSlice[v], jsonKey, unit[v])
								fmt.Fprintf(&str, "%s", valueString)
							} else {
								valueString := checkValue(dataItem, correctSlice[v], jsonKey, "")
								fmt.Fprintf(&str, "%s", valueString)

							}

						}
					}
					str.WriteString("\n")
				}
				str.WriteString("#\n")
			case size == 1:
				l := getLongestPDBxItem(catDI) + 5
				for _, dataItem := range catDI {
					jsonKey, err := getKeyByValue(dataItem.CategoryID+"."+dataItem.Name, nameMapper)
					if err != nil {
						return "", err
					}

					if valuesMap[jsonKey] == nil {
						continue // not required in mmCIF
					}
					formatString := fmt.Sprintf("%%-%ds", l)
					fmt.Fprintf(&str, formatString, dataItem.CategoryID+"."+dataItem.Name)
					if jsonValue, ok := valuesMap[jsonKey]; ok {
						if unit, ok := OSCEMunits[jsonKey]; ok {
							valueString := checkValue(dataItem, jsonValue[0], jsonKey, unit[0]) // the 0th element, because it's the case where only one value is present
							fmt.Fprintf(&str, "%s", valueString)
						} else {
							valueString := checkValue(dataItem, jsonValue[0], jsonKey, "")
							fmt.Fprintf(&str, "%s", valueString)

						}
					}
					str.WriteString("\n")
				}
				str.WriteString("#\n")
			default:
				// this key does not exist in the json
				continue
			}
		}
		mmCifLines, ok := mmCIFCategories[category]
		if ok {
			str.WriteString(mmCifLines)
		}
	}
	return str.String(), nil

}
