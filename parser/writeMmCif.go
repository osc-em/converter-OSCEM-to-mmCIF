package parser

import (
	"bufio"
	"converter/converterUtils"
	"fmt"
	"log"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// for a property category.dataItem extract the category name i.e. string before dot
func itemCategory(item string) string {
	return strings.Split(item, ".")[0]
}

// find all data items that have the same category as item
func findItemByCategory(item string, mapper map[string]string) []string {
	itemsInCategory := make([]string, 0)
	for _, v := range mapper {
		if itemCategory(v) == item {
			itemsInCategory = append(itemsInCategory, v)
		}
	}
	return itemsInCategory
}

func getKeyByValue(value string, m map[string]string) string {
	for k, v := range m {
		if v == value {
			return k
		}
	}
	return ""
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

func validateRange(value string, rMin string, rMax string, unitOSCEM string, unitPDBx string, name string, name2 string) bool {
	if unitOSCEM == unitPDBx {
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			log.Fatal("JSON value not numeric, but supposed to be", err)
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
			return true
		} else if math.IsNaN(rMin) {
			return float64(v) <= rMax
		} else if math.IsNaN(rMax) {
			return float64(v) >= rMin
		} else {

			return float64(v) >= rMin && float64(v) <= rMax
		}
	} else {
		log.Printf("Units in %s vs %s  don't match! Implement a converter from %s to %s units\n", name, name2, unitOSCEM, unitPDBx)
		return true
	}
}
func validateEnum(value string, enumFromPDBx []string, dataItem string) string {
	if value == "true" {
		return "YES"
	} else if value == "false" {
		return "NO"
	}
	for i := range enumFromPDBx {
		if dataItem == "microscope_model" {
			reTitan := regexp.MustCompile(`(?i)titan`)
			if reTitan.MatchString(value) {
				return "TFS KRIOS"
			} else {
				return strings.ToUpper(value)
			}
		} else if strings.EqualFold(enumFromPDBx[i], value) {
			return enumFromPDBx[i]
		} else {
			log.Println("value is not in enum!", value) //make better logging
			return value
		}
	}
	return strings.ToUpper(value)
}

func checkValue(dataItem converterUtils.PDBxItem, value string, jsonKey string, unitsOSCEM string) string {

	//now based on the found struct implement range matching, units matching and enum matching
	if dataItem.ValueType == "int" || dataItem.ValueType == "float" {
		validateRange := validateRange(value, dataItem.RangeMin, dataItem.RangeMax, unitsOSCEM, dataItem.Unit, jsonKey, dataItem.Name)
		if !validateRange {
			log.Printf("Value %s of property %s is not in range of [ %s, %s ]!\n", value, jsonKey, dataItem.RangeMin, dataItem.RangeMax)
		}
	} else if dataItem.ValueType == "yyyy-mm-dd" {
		value = validateDateIsRFC3339(value)
	} else if len(dataItem.EnumValues) > 0 {
		value = validateEnum(value, dataItem.EnumValues, dataItem.Name)
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
func ToMmCIF(nameMapper map[string]string, PDBxItems map[string][]converterUtils.PDBxItem, valuesMap map[string]any, OSCEMunits map[string]any, appendToMmCif bool, mmCIFpath string) string {
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
				key = getKeyByValue(catDI[i].CategoryID+"."+catDI[i].Name, nameMapper)
				_, ok := valuesMap[key]
				if ok {
					break
				}
			}
			switch jsonValueType := valuesMap[key].(type) {
			case []string: // loop notation
				str.WriteString("loop_\n")
				for _, dataItem := range catDI {
					// check the length of all first and throw an error in case that they have different length?? can that be? e.g two authors and for one the property Phone is not there?
					jsonKey := getKeyByValue(dataItem.CategoryID+"."+dataItem.Name, nameMapper)
					if valuesMap[jsonKey] == nil {
						continue // not required and not provided in OSCEM
					}
					fmt.Fprintf(&str, "%s\n", dataItem.CategoryID+"."+dataItem.Name)
				}
				for v := range jsonValueType {
					for _, dataItem := range catDI {
						jsonKey := getKeyByValue(dataItem.CategoryID+"."+dataItem.Name, nameMapper)

						if valuesMap[jsonKey] == nil {
							continue // key was not required and not provided in OSCEM
						}
						if correctSlice, ok := valuesMap[jsonKey].([]string); ok {

							switch unit := OSCEMunits[jsonKey].(type) {
							case string:
								valueString := checkValue(dataItem, correctSlice[v], jsonKey, unit)
								fmt.Fprintf(&str, "%s", valueString)
							default:
								log.Printf("%s property is defined to have no units in OSCEM", jsonKey)
							}
						}
					}
					str.WriteString("\n")
				}
				str.WriteString("#\n")

			case string:
				l := getLongestPDBxItem(catDI) + 5
				for _, dataItem := range catDI {
					jsonKey := getKeyByValue(dataItem.CategoryID+"."+dataItem.Name, nameMapper)
					if valuesMap[jsonKey] == nil {
						continue // not required in mmCIF
					}
					formatString := fmt.Sprintf("%%-%ds", l)
					fmt.Fprintf(&str, formatString, dataItem.CategoryID+"."+dataItem.Name)
					if jsonValue, ok := valuesMap[jsonKey].(string); ok {
						switch unit := OSCEMunits[jsonKey].(type) {
						case string:
							valueString := checkValue(dataItem, jsonValue, jsonKey, unit) //OSCEMunits[jsonKey])
							fmt.Fprintf(&str, "%s", valueString)
						default:
							log.Printf("%s property is defined to have no units in OSCEM is not a string", jsonKey)
						}

					}
					str.WriteString("\n")
				}
				str.WriteString("#\n")
			default:
				fmt.Printf("Problem appeared while unmarshalling JSON: %s for key %s\n", jsonValueType, key)
			}

		}
		mmCifLines, ok := mmCIFCategories[category]
		if ok {
			str.WriteString(mmCifLines)
		}
	}
	return str.String()

}
