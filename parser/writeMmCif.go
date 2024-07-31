package parser

import (
	"converter/converterUtils"
	"fmt"
	"strings"
)

func itemParent(item string) string {
	return strings.Split(item, ".")[0]
}

func findElemByItem(item string, mapper map[string]string, toMmCIF bool) []string {
	elements := make([]string, len(mapper))
	if toMmCIF {
		i := 0
		for e := range mapper {
			elements[i] = mapper[e]
			i++
		}
	} else {
		elements = converterUtils.GetKeys(mapper)
	}
	var simElem []string
	for i := range elements {
		if itemParent(elements[i]) == item {
			simElem = append(simElem, elements[i])
		}
	}
	return simElem
}

func ToMmCIF(nameMapper map[string]string, PDBxItems map[string][]string, valuesMap map[string]any) string {
	// values from JSON are mapped to the mmcif properties
	mappedVal := make([]string, 0, len(nameMapper))
	var str strings.Builder
	str.WriteString("data_myID\n#\n")
	for jsonName := range valuesMap {
		PDBxName := nameMapper[jsonName]
		if PDBxName == "" {
			continue // because translation is iterating on json, it still contains elements that don't exist in mmcif
		}
		var orderedElements = PDBxItems[itemParent(PDBxName)[1:]]
		if converterUtils.Contains(mappedVal, PDBxName) {
			continue
		}
		elements := findElemByItem(itemParent(PDBxName), nameMapper, true)
		var valueString string
		switch valSlice := valuesMap[jsonName].(type) {
		case []string: // loop notation
			if valSlice == nil {
				continue // not required in mmCIF
			}
			str.WriteString("loop_\n")

			// list category names
			// instead of all find elements use the ordered list from parsing the dictionary

			for _, oE := range orderedElements {
				for _, e := range elements {
					if oE == strings.Split(e, ".")[1] {
						fmt.Fprintf(&str, "%s\n", e)
						mappedVal = append(mappedVal, e)
					}
				}
			}
			// write the values
			for i := range len(valSlice) {
				for _, oE := range orderedElements {
					for _, e := range elements {
						if oE == strings.Split(e, ".")[1] {
							jsonKey := converterUtils.GetKeyByValue(e, nameMapper)

							if slice, ok := valuesMap[jsonKey].([]string); ok {
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
			l := converterUtils.GetLongest(elements) + 5
			for _, oE := range orderedElements {
				for _, e := range elements {
					if oE == strings.Split(e, ".")[1] {

						jsonKey := converterUtils.GetKeyByValue(e, nameMapper)
						if valuesMap[jsonKey] == nil {
							continue // not required in mmCIF
						}
						formatString := fmt.Sprintf("%%-%ds", l)
						fmt.Fprintf(&str, formatString, e)
						if jsonValue, ok := valuesMap[jsonKey].(string); ok {

							if strings.Contains(jsonValue, " ") {
								valueString = fmt.Sprintf("'%s'\n", jsonValue)
							} else {
								valueString = fmt.Sprintf("%s\n", jsonValue)
							}
							fmt.Fprintf(&str, "%s", valueString)
						}

						mappedVal = append(mappedVal, e)
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
