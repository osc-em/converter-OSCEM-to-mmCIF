package parser

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/osc-em/converter-OSCEM-to-mmCIF/converterUtils"
)

func relevantId(PDBxItems map[string][]converterUtils.PDBxItem, dataItem converterUtils.PDBxItem) bool {
	accessValues := PDBxItems[dataItem.CategoryID]
	var parentValue string
	for i := range len(accessValues) {
		for r := range len(accessValues[i].ChildName) {
			if accessValues[i].ChildName[r] == dataItem.CategoryID+"."+dataItem.Name {
				//if accessValues[i].CategoryID == dataItem.CategoryID && accessValues[i].Name == dataItem.Name && len(accessValues[i].ParentName) != 0 {
				parentValue = accessValues[i].ParentName[r]
			}
		}
	}
	if parentValue != "" {
		catValues, ok := PDBxItems[strings.Split(parentValue, ".")[0]]
		if ok {

			for r := range len(catValues) {
				if catValues[r].Name == strings.Split(parentValue, ".")[1] {
					return true
				}
			}
		}
	}
	return false
}

func getKeyByValue(value string, m map[string]string) (string, error) {
	for k, v := range m {
		if v == value {
			return k, nil
		}
	}
	return "", fmt.Errorf("value %v is not in the conversion table", value)
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
	if namePDBx == "_em_imaging.illumination_mode" {
		reFloodBeam := regexp.MustCompile(`(?i)parallel`)
		if reFloodBeam.MatchString(value) {
			return "FLOOD BEAM"
		}
	}
	if namePDBx == "_pdbx_contact_author.role" {
		reRole := regexp.MustCompile(`(?i)(principal investigator|group leader|pi)`)
		if reRole.MatchString(value) {
			return "principal investigator/group leader"
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
	// // if no match and enum contains option for OTHER, choose it
	// if converterUtils.SliceContains(enumFromEMDB, "OTHER") || converterUtils.SliceContains(enumFromPDBx, "OTHER"){
	// 	return "OTHER"
	// }

	return strings.ToUpper(value)
}

func checkValue(dataItem converterUtils.PDBxItem, value string, jsonKey string, unitsOSCEM string) string {

	//now based on the found struct implement range matching, units matching and enum matching
	if dataItem.ValueType == "int" || dataItem.ValueType == "float" {
		namePDBx := dataItem.CategoryID + "." + dataItem.Name
		// in OSCEM defocus is negative value and overfocus is positive. In PDBx it's vice versa if string starts with  minus, ut it off, othersise add a - prefix
		if namePDBx == "_em_imaging.nominal_defocus_min" || namePDBx == "_em_imaging.calibrated_defocus_min" || namePDBx == "_em_imaging.nominal_defocus_max" || namePDBx == "_em_imaging.calibrated_defocus_max" {
			// change the sign negative to positiove and vice versa
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

func getOrderCategories(parsedCategories []string, mmCIFCategories []string) []string {
	var order []string
	allCategories := append(parsedCategories, mmCIFCategories...)
	sort.Slice(allCategories, func(i, j int) bool {
		return allCategories[i] < allCategories[j]
	})
	// sort based on the pre-defined (administrative, polymer related entities, ligand (non-polymer) related instances, and structure level description)
	for _, category := range converterUtils.PDBxCategoriesOrder {
		category = "_" + category
		if converterUtils.SliceContains(allCategories, category) {
			order = append(order, category)
		}
	}
	// add the rest not atom-related in some order (it can be random)
	for _, c := range allCategories {
		if !converterUtils.SliceContains(order, c) && !converterUtils.SliceContains(converterUtils.PDBxCategoriesOrderAtom, c[1:]) && !(len(c) > 5 && c[0:5] == "data_") {
			order = append(order, c)
		}

	}
	// add atoms categories
	for _, category := range converterUtils.PDBxCategoriesOrderAtom {
		category = "_" + category
		if converterUtils.SliceContains(allCategories, category) {
			order = append(order, category)
		}
	}
	// append the rest of "unparsed" categories that were inside their own "data_" containers
	for i := range mmCIFCategories {
		if mmCIFCategories[i][0:5] == "data_" {
			order = append(order, mmCIFCategories[i])
		}
	}
	return order
}

func parseMmCIF(dictFile *os.File) (string, map[string]string, error) {
	scanner := bufio.NewScanner(dictFile)

	var longestData uint32 = 0
	var dataToStr = make(map[string]string, 0)
	var mmCIFfieldsMain map[string]string // should always point to the main model
	var longestNameData string
	// initialize first category
	scanner.Scan()
	firstName := scanner.Text()
	// first run
	longestNameData = firstName
	mmCIFfields, bigString, l, new_data := LoopDataEntry(scanner, firstName)
	mmCIFfieldsMain = mmCIFfields
	dataToStr[firstName] = bigString
	longestData = l

	for new_data != "" {
		mmCIFfields, s, l, d := LoopDataEntry(scanner, new_data)
		dataToStr[new_data] = s
		if l > longestData {
			mmCIFfieldsMain = mmCIFfields
			longestNameData = new_data
		}
		new_data = d
	}

	for k, v := range dataToStr {
		if k != longestNameData {
			// add the data_prefix back
			mmCIFfieldsMain[k] = k + "\n" + v
		}
	}

	return longestNameData, mmCIFfieldsMain, nil
}

func LoopDataEntry(scanner *bufio.Scanner, category string) (map[string]string, string, uint32, string) { // return the categories, same as long string, the length and next category
	var l uint32 = 0
	var str strings.Builder
	var asText strings.Builder
	inCategoryFlag := true
	var mmCIFfields = make(map[string]string, 0)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data_") {
			category = line
			return mmCIFfields, asText.String(), l, category
		}
		l++
		asText.WriteString(line + "\n")
		if strings.HasPrefix(line, "#") || len(strings.Fields(line)) == 0 {
			// break between categories is denoted by # in PDB-related software, Phenix uses an empty line.
			// category ends, appends to the map
			if category != "" {
				mmCIFfields[category] = str.String()
				str.Reset()
				inCategoryFlag = true // record the next category name
			}
		} else {
			if !strings.HasPrefix(line, "loop_") {

				if inCategoryFlag {
					category = strings.Split(strings.Fields(line)[0], ".")[0]
					inCategoryFlag = false
				}
			}
			str.WriteString(line + "\n")

		}
	}
	return mmCIFfields, asText.String(), l, ""
}

// CreteMetadataCif creates and mmCIF file content as a string. This function only converts metadata in the required format.
// Meant for EMDB depositions where no other existing mmCIF is available
func CreteMetadataCif(nameMapper map[string]string, PDBxItems map[string][]converterUtils.PDBxItem, valuesMap map[string][]string, OSCEMunits map[string][]string) (string, error) {
	return createCifText("data_myID", map[string]string{}, nameMapper, PDBxItems, valuesMap, OSCEMunits)
}

// Given an mmCIF file create a new one with added scientific Metadata.
// Meant for PDB depositions
func SupplementCoordinatesFromFile(nameMapper map[string]string, PDBxItems map[string][]converterUtils.PDBxItem, valuesMap map[string][]string, OSCEMunits map[string][]string, mmCIFpath *os.File) (string, error) {
	var dataID string
	var mmCIFvalues map[string]string

	name, categories, err := parseMmCIF(mmCIFpath)
	if err != nil {
		return "", err
	}
	dataID = name
	mmCIFvalues = categories
	return createCifText(dataID, mmCIFvalues, nameMapper, PDBxItems, valuesMap, OSCEMunits)
}

// Given an mmCIF file path, open it and create a new one with added scientific Metadata.
// Meant for PDB depositions
func SupplementCoordinatesFromPath(nameMapper map[string]string, PDBxItems map[string][]converterUtils.PDBxItem, valuesMap map[string][]string, OSCEMunits map[string][]string, mmCIFpath string) (string, error) {
	var dataID string
	var mmCIFvalues map[string]string
	dictFile, err := os.Open(mmCIFpath)
	if err != nil {
		errorString := fmt.Sprintf("mmCIF file %s does not exist!", mmCIFpath)
		return "", errors.New(errorString)
	}
	defer dictFile.Close()

	name, categories, err := parseMmCIF(dictFile)
	if err != nil {
		return "", err
	}
	dataID = name
	mmCIFvalues = categories
	return createCifText(dataID, mmCIFvalues, nameMapper, PDBxItems, valuesMap, OSCEMunits)
}
func createCifText(dataName string, mmCIFCategories map[string]string, nameMapper map[string]string, PDBxItems map[string][]converterUtils.PDBxItem, valuesMap map[string][]string, OSCEMunits map[string][]string) (string, error) {

	// keeps track of values from JSON that have already been mapped to the PDBx properties
	var str strings.Builder
	str.WriteString(dataName + "\n#\n") //write the data Identifier in the header

	parsedCategories := make([]string, 0)
	for k := range PDBxItems {
		parsedCategories = append(parsedCategories, k)
	}
	allCategories := getOrderCategories(parsedCategories, converterUtils.GetKeys(mmCIFCategories))

	for _, category := range allCategories {
		var duplicatedFlag bool = false
		catDI, ok := PDBxItems[category]
		if ok {
			_, ok := mmCIFCategories[category]
			if ok {
				duplicatedFlag = true
				log.Printf("Category %s exists both in metadata from JSON files and in existing mmCIF file! Data in mmCIF will be substituted by data from JSON", category)
			}
			//
			var size int
			// need to loop here through all data items in category, as it reflects the order of data items in PDBx, it still might not exist in json
			// loop until we find first key that exists in json
			for i := range catDI {
				k, err := getKeyByValue(catDI[i].CategoryID+"."+catDI[i].Name, nameMapper)
				if err != nil {
					//occurs when this PDBx category not in the conversions table - contains the "id" and used to link data items -> use something else for counting
					continue
				}
				//check if that key is present in json file and extract it's size
				_, ok := valuesMap[k]
				if ok {
					size = len(valuesMap[k])
					break
				}
			}

			switch {
			case size > 1:
				// loop notation
				str.WriteString("loop_\n")
				var isRelevantID bool
				for _, dataItem := range catDI {
					// check the length of all first and throw an error in case that they have different length?? can that be? e.g two authors and for one the property Phone is not there?
					jsonKey, err := getKeyByValue(dataItem.CategoryID+"."+dataItem.Name, nameMapper)
					if err != nil {
						// it is _id property -> check if we need it go through all data items andd see if it's a parent somewhere!
						isRelevantID = relevantId(PDBxItems, dataItem)
						if isRelevantID {
							fmt.Fprintf(&str, "%s\n", dataItem.CategoryID+"."+dataItem.Name)
						}
					} else if valuesMap[jsonKey] == nil {
						continue // not required and not provided in OSCEM
					} else {
						fmt.Fprintf(&str, "%s\n", dataItem.CategoryID+"."+dataItem.Name)
					}
				}
				for v := range size {
					for _, dataItem := range catDI {
						jsonKey, err := getKeyByValue(dataItem.CategoryID+"."+dataItem.Name, nameMapper)
						if err != nil {
							// errorString := fmt.Sprintf("Value %s for PDBx is not in the names conversion!", dataItem.CategoryID+"."+dataItem.Name)
							// return "", errors.New(errorString)

							if isRelevantID {
								fmt.Fprintf(&str, "%v ", v+1)
							}

						} else if valuesMap[jsonKey] == nil {
							continue // key was not required and not provided in OSCEM
						} else if correctSlice, ok := valuesMap[jsonKey]; ok {
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
				var isRelevantID bool
				for _, dataItem := range catDI {
					jsonKey, err := getKeyByValue(dataItem.CategoryID+"."+dataItem.Name, nameMapper)
					if err != nil {
						// it is _id property -> check if we need it go through all data items andd see if it's a parent somewhere!
						isRelevantID = relevantId(PDBxItems, dataItem)
						if isRelevantID {
							formatString := fmt.Sprintf("%%-%ds", l)
							fmt.Fprintf(&str, formatString, dataItem.CategoryID+"."+dataItem.Name)

							fmt.Fprintf(&str, "%v", 1)
							str.WriteString("\n")
							continue
						}
					}

					if valuesMap[jsonKey] == nil {
						continue // not required in mmCIF
					}
					if jsonValue, ok := valuesMap[jsonKey]; ok {

						formatString := fmt.Sprintf("%%-%ds", l)
						fmt.Fprintf(&str, formatString, dataItem.CategoryID+"."+dataItem.Name)

						if unit, ok := OSCEMunits[jsonKey]; ok {
							valueString := checkValue(dataItem, jsonValue[0], jsonKey, unit[0]) // the 0th element, because it's the case where only one value is present
							fmt.Fprintf(&str, "%s", valueString)
						} else {
							// values that have no units definition in OSCEM
							valueString := checkValue(dataItem, jsonValue[0], jsonKey, "")
							fmt.Fprintf(&str, "%s", valueString)

						}
					}
					str.WriteString("\n")
				}
				str.WriteString("#\n")
			default:
				// Based on conversion table, the correspondance in naming between OSCEM and PDBx exist. But for this PDBx data category not OSCEM propoerties are used in this JSON file.
				continue
			}
		}
		if !duplicatedFlag {
			// this category is not present both in mmCIF and in a new metadata.
			// We won't duplicate it from mmCIF since it was taken from new metadata!
			mmCifLines, ok := mmCIFCategories[category]
			if ok {
				str.WriteString(mmCifLines)
				str.WriteString("#\n")
			}
		}
	}
	return str.String(), nil
}
