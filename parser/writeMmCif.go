package parser

import (
	"converter/converterUtils"
	"fmt"
	"log"
	"math"
	"regexp"
	"strconv"
	"strings"
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

// given a slice of strings get the length of a longest in it
func getLongest(s []string) int {
	var r int
	for _, a := range s {
		if len(a) > r {
			r = len(a)
		}
	}
	return r
}

func valueInRange(value string, rMin float64, rMax float64, unitOSCEM string, unitPDBx string, name string, name2 string) bool {
	if unitOSCEM == unitPDBx {
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			log.Fatal("JSON value not numeric, but supposed to be", err)
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
func valueInEnum(value string, enumFromPDBx []string, dataItem string) string {
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

func ToMmCIF(nameMapper map[string]string, PDBxItems map[string][]converterUtils.PDBxItem, valuesMap map[string]any, OSCEMunits map[string]string) string {
	// keeps track of values from JSON that have already been mapped to the PDBx properties
	mappedVal := make([]string, 0, len(nameMapper))
	var str strings.Builder
	str.WriteString("data_myID\n#\n")
	for jsonName := range valuesMap {
		PDBxName := nameMapper[jsonName]
		if PDBxName == "_em_imaging.tilt_angle_increment" {
			fmt.Println("_em_imaging.tilt_angle_increment is not filled in properly in PDBx and might not really exist!")
			continue
		}
		if PDBxName == "" {
			continue // because translation is iterating on json, it still contains elements that don't exist in mmcif
		}
		//extract correct order for PDBx category
		var orderedItems = PDBxItems[itemCategory(PDBxName)]
		//fmt.Println(mappedVal, itemCategory(PDBxName))
		if sliceContains(mappedVal, itemCategory(PDBxName)) {
			continue
		}

		inSameCategory := findItemByCategory(itemCategory(PDBxName), nameMapper)

		var valueString string
		switch valSlice := valuesMap[jsonName].(type) {
		case []string: // loop notation
			if valSlice == nil {
				continue // not required in mmCIF
			}
			str.WriteString("loop_\n")

			// list category names
			// instead of all find elements use the ordered list from parsing the dictionary

			for _, oE := range orderedItems {
				for _, e := range inSameCategory {
					if oE.Name == strings.Split(e, ".")[1] {
						fmt.Fprintf(&str, "%s\n", e)
						mappedVal = append(mappedVal, oE.CategoryID)
					}
				}
			}
			// write the values
			for i := range len(valSlice) {
				for _, oE := range orderedItems {
					for _, e := range inSameCategory {
						if oE.Name == strings.Split(e, ".")[1] {

							jsonKey := getKeyByValue(e, nameMapper)
							if slice, ok := valuesMap[jsonKey].([]string); ok {
								//now based on the found struct implement range matching, units matching and enum matching
								if oE.ValueType == "int" || oE.ValueType == "float" {
									valueInRange := valueInRange(slice[i], oE.RangeMin, oE.RangeMax, OSCEMunits[jsonKey], oE.Unit, jsonKey, oE.Name)
									if !valueInRange {
										log.Printf("Value %s of property %s is not in range of [ %f, %f ]!\n", slice[i], jsonKey, oE.RangeMin, oE.RangeMax)
									}
								} else if len(oE.EnumValues) > 0 {
									slice[i] = valueInEnum(slice[i], oE.EnumValues, oE.Name)
								}

								if strings.Contains(slice[i], " ") {
									valueString = fmt.Sprintf("'%s' ", slice[i]) // if name contains whitespaces enclose it in single quotes
								} else {
									valueString = fmt.Sprintf("%s ", slice[i]) // take value as is
								}
							} else { // if name is present in both OSCEM and PDBx but no value is available set it as "omitted"
								valueString = ". "
							}
							fmt.Fprintf(&str, "%s", valueString)
						}
					}
				}
				str.WriteString("\n")
			}
			str.WriteString("#\n")
		case string: // simple list of categories
			l := getLongest(inSameCategory) + 5
			for _, oE := range orderedItems {
				for _, e := range inSameCategory {
					if oE.Name == strings.Split(e, ".")[1] {
						jsonKey := getKeyByValue(e, nameMapper)
						if valuesMap[jsonKey] == nil {
							continue // not required in mmCIF
						}
						formatString := fmt.Sprintf("%%-%ds", l)
						fmt.Fprintf(&str, formatString, e)
						if jsonValue, ok := valuesMap[jsonKey].(string); ok {
							if oE.ValueType == "int" || oE.ValueType == "float" {
								valueInRange := valueInRange(jsonValue, oE.RangeMin, oE.RangeMax, OSCEMunits[jsonKey], oE.Unit, jsonKey, oE.Name)
								if !valueInRange {
									log.Printf("Value %s of property %s is not in range of [ %f, %f ]!\n", jsonValue, jsonKey, oE.RangeMin, oE.RangeMax)
								}
							} else if len(oE.EnumValues) > 0 {
								jsonValue = valueInEnum(jsonValue, oE.EnumValues, oE.Name)
							}
							if strings.Contains(jsonValue, " ") {
								valueString = fmt.Sprintf("'%s'\n", jsonValue)
							} else {
								valueString = fmt.Sprintf("%s\n", jsonValue)
							}
							fmt.Fprintf(&str, "%s", valueString)
						}

						mappedVal = append(mappedVal, oE.CategoryID)
					}
				}
			}
			str.WriteString("#\n")
		default:
			fmt.Println("Problem appeared while unmarshalling JSON")
		}
	}
	return str.String()
}