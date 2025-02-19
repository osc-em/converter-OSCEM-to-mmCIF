package parser

import (
	"fmt"
	"log"

	"github.com/osc-em/converter-OSCEM-to-mmCIF/converterUtils"
)

func keyConcat(s1, s2 string) string {
	if s1 != "" {
		return s1 + "." + s2
	} else {
		return s2
	}
}

// visit recursively unwinds the nested JSON file.
// map initial is the result of a simple unmarshalling on a json file
// newJsonKey is the key for a map at the current level
// mapResult is a map that accumulates a resulting dictionary
// propertiesLen variable controls how many items are kept in the json value array
// propertyI is the current index of iterated slice
// currentLevel is the level of nesting
// level is an initial name where metadata is stored that needs to be converted
// levelIndex is the nesting level of a propoerty that we should keep track of
// save is a boolean variable that controls if the correct metadata level was entered already
// savedLevel a boolean variable that controls if that level was saved once, should turn off after that
// arrayNestingLevel is a boolean that controls that elements array in json does not have another array element inside. Multiple nesting is not supported
// return a mapResult
func visit(mapInitial map[string]any, newJsonKey string, mapResult map[string][]string, mapResultUnits map[string][]string, propertiesLen int, propertyI int, currentLevel int, level string, levelIndex int, save bool, savedLevel bool, arrayNestingLevel int) (map[string][]string, map[string][]string) {
	keys := converterUtils.GetKeys(mapInitial)
	for _, k := range keys {
		// first we need to go down to the correct json level, that contains the property for metadata
		if !save && level != "" && level != k {
			// go down as the key is not found on this level
			switch nestedMap := mapInitial[k].(type) {
			case map[string]any:
				visit(nestedMap, "", mapResult, mapResultUnits, propertiesLen, propertyI, currentLevel+1, level, levelIndex, false, savedLevel, arrayNestingLevel)
			default:
				continue
			}
		} else if level != "" && level == k {
			// here we will enter the correct level
			switch nestedMap := mapInitial[k].(type) {
			case map[string]any:
				save = true
				levelIndex = currentLevel
				visit(nestedMap, "", mapResult, mapResultUnits, propertiesLen, propertyI, currentLevel+1, level, levelIndex, save, savedLevel, arrayNestingLevel)
			default:
				log.Printf("Metadata Level '%s' is not a nested json structure", level)
				return mapResult, mapResultUnits
			}
		} else if level != "" && levelIndex == currentLevel && !savedLevel {
			// here are the keys that are at the same level as the metadata, but we saved the metadata already
			continue
		} else {
			jsonFlat := keyConcat(newJsonKey, k)
			switch nestedMap := mapInitial[k].(type) {
			case map[string]any: //value is a nested json structure
				if arrayNestingLevel == currentLevel {
					arrayNestingLevel = -1 //reset
				}
				value, okV := nestedMap["value"]
				unit, okU := nestedMap["unit"]
				if okV && okU {
					if propertiesLen == 1 {
						mapResult[jsonFlat] = []string{fmt.Sprint(value)}
						mapResultUnits[jsonFlat] = []string{fmt.Sprint(unit)}
					} else {
						// if it's a first one, it needs initialization
						if propertyI == 0 {
							mapResult[jsonFlat] = make([]string, propertiesLen)
							mapResult[jsonFlat][0] = fmt.Sprint(value)
							mapResultUnits[jsonFlat] = make([]string, propertiesLen)
							mapResultUnits[jsonFlat][0] = fmt.Sprint(unit)
						} else {
							_, okMV := mapResult[jsonFlat]
							if okMV {
								mapResult[jsonFlat][propertyI] = fmt.Sprint(value)
								mapResultUnits[jsonFlat][propertyI] = fmt.Sprint(unit)
							} else {
								mapResult[jsonFlat] = make([]string, propertiesLen)
								mapResult[jsonFlat][propertyI] = fmt.Sprint(value)
								mapResultUnits[jsonFlat] = make([]string, propertiesLen)
								mapResultUnits[jsonFlat][propertyI] = fmt.Sprint(unit)
							}
						}
					}

				} else {
					visit(nestedMap, jsonFlat, mapResult, mapResultUnits, propertiesLen, propertyI, currentLevel+1, level, levelIndex, save, savedLevel, arrayNestingLevel)
				}
			case []any: // value is array
				if arrayNestingLevel >= 0 && arrayNestingLevel < currentLevel {
					log.Printf("MULTIPLE NESTING NOT SUPPORTED YET!")
					break
				}
				for i, jsonSlice := range nestedMap {
					if js, ok := jsonSlice.(map[string]any); ok {
						visit(js, jsonFlat, mapResult, mapResultUnits, len(nestedMap), i, currentLevel+1, level, levelIndex, save, savedLevel, currentLevel)
					}
				}
			default: // the value is just a value (as value/unit pairs fall into a map category)
				if arrayNestingLevel == currentLevel {
					arrayNestingLevel = -1 //reset
				}
				if propertiesLen == 1 {
					// if it's an array that only contains one element, save instead as a value
					mapResult[jsonFlat] = []string{fmt.Sprint(mapInitial[k])}
				} else {
					if propertyI == 0 {
						mapResult[jsonFlat] = make([]string, propertiesLen)
						mapResult[jsonFlat][0] = fmt.Sprint(mapInitial[k])
					} else {
						_, okMV := mapResult[jsonFlat]
						if okMV {
							mapResult[jsonFlat][propertyI] = fmt.Sprint(mapInitial[k])
						} else {
							mapResult[jsonFlat] = make([]string, propertiesLen)
							mapResult[jsonFlat][propertyI] = fmt.Sprint(mapInitial[k])
						}
					}
				}
			}

			savedLevel = true
		}
	}
	return mapResult, mapResultUnits
}

// FromJson function creates a map from a JSON file. Nested data is flattend and keys are connected by a dot.
// In cases where JSON has multiple values, final map's value will be a slice.

func FromJson(jsonContent map[string]any, mapAll *map[string][]string, mapUnits *map[string][]string, level string) error {
	// Initialize flattened maps
	jsonFlat := make(map[string][]string)
	jsonUnits := make(map[string][]string)
	// Process JSON content and populate the maps
	jsonFlat, jsonUnits = visit(jsonContent, "", jsonFlat, jsonUnits, 1, 0, 0, level, 0, false, false, -1)
	// Update the provided maps
	for k, v := range jsonFlat {
		(*mapAll)[k] = v
	}
	for k, v := range jsonUnits {
		(*mapUnits)[k] = v
	}

	return nil
}
