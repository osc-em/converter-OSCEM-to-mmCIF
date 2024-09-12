package parser

import (
	"converter/converterUtils"
	"encoding/json"
	"fmt"
	"log"
	"os"
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
// arrayNestingLevel is a boolean that controls that elements array in json doea not have another array element inside. Multiple nesting is not supported
// return a mapResult
func visit(mapInitial map[string]any, newJsonKey string, mapResult map[string]any, mapResultUnits map[string]any, propertiesLen int, propertyI int, currentLevel int, level string, levelIndex int, save bool, savedLevel bool, arrayNestingLevel int) (map[string]any, map[string]any) {
	keys := converterUtils.GetKeys(mapInitial)
	for _, k := range keys {

		if !save && level != "" && level != k {
			switch nestedMap := mapInitial[k].(type) {
			case map[string]any:
				visit(nestedMap, "", mapResult, mapResultUnits, propertiesLen, propertyI, currentLevel+1, level, levelIndex, false, savedLevel, arrayNestingLevel)
			default:
				continue
			}
		} else if level != "" && level == k {
			switch nestedMap := mapInitial[k].(type) {
			case map[string]any:
				// here we will enter the correct level
				save = true
				levelIndex = currentLevel
				visit(nestedMap, "", mapResult, mapResultUnits, propertiesLen, propertyI, currentLevel+1, level, levelIndex, save, savedLevel, arrayNestingLevel)
			default:
				log.Printf("Metadata Level '%s' is not a nested json structure", level)
				return mapResult, mapResultUnits
			}
		} else if level != "" && levelIndex == currentLevel && !savedLevel {
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
						mapResult[jsonFlat] = fmt.Sprint(value)
						mapResultUnits[jsonFlat] = fmt.Sprint(unit)
					} else {
						if propertyI == 0 {
							values := make([]string, propertiesLen)
							values[0] = fmt.Sprint(value)
							mapResult[jsonFlat] = values
							units := make([]string, propertiesLen)
							units[0] = fmt.Sprint(unit)
							mapResultUnits[jsonFlat] = units
						} else {
							values := mapResult[jsonFlat]
							units := mapResultUnits[jsonFlat]
							if v, ok := values.([]string); ok {
								v[propertyI] = fmt.Sprint(value)
								mapResult[jsonFlat] = v
							}
							if u, ok := units.([]string); ok {
								u[propertyI] = fmt.Sprint(unit)
								mapResultUnits[jsonFlat] = u
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
					// else {
					// 	log.Fatal("Weird case ") // should be covered by initially unmarshalling json. coming here would mean json was broken but that should have been checked
					// }
				}
			default: // the value is just a value
				if arrayNestingLevel == currentLevel {
					arrayNestingLevel = -1 //reset
				}
				if propertiesLen == 1 {
					// if it's an array that only contains one element, save instead as a value
					mapResult[jsonFlat] = fmt.Sprint(mapInitial[k])
				} else {
					if propertyI == 0 {
						values := make([]string, propertiesLen)
						values[0] = fmt.Sprint(mapInitial[k])
						mapResult[jsonFlat] = values
					} else {
						values := mapResult[jsonFlat]
						if v, ok := values.([]string); ok {
							v[propertyI] = fmt.Sprint(mapInitial[k])
							mapResult[jsonFlat] = v
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
func FromJson(jsonPath string, mapAll *map[string]any, mapUnits *map[string]any, level string) error {
	// Read JSON file
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return fmt.Errorf("error while reading the JSON file: %w", err)
	}

	// Unmarshal JSON
	var jsonContent map[string]any
	if err := json.Unmarshal(data, &jsonContent); err != nil {
		return fmt.Errorf("error while unmarshaling JSON: %w", err)
	}

	// Initialize flattened maps
	jsonFlat := make(map[string]any)
	jsonUnits := make(map[string]any)

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
