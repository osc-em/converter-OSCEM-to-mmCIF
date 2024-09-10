package parser

import (
	"converter/converterUtils"
	"encoding/json"
	"fmt"
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
// return a mapResult
func visit(mapInitial map[string]any, newJsonKey string, mapResult map[string]any, mapResultUnits map[string]any, propertiesLen int, propertyI int) (map[string]any, map[string]any) {
	keys := converterUtils.GetKeys(mapInitial)
	for _, k := range keys {
		jsonFlat := keyConcat(newJsonKey, k)
		switch nestedMap := mapInitial[k].(type) {
		case map[string]any: //value is a nested json structure
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
				visit(nestedMap, jsonFlat, mapResult, mapResultUnits, propertiesLen, propertyI)
			}
		case []any: // value is array
			for i, jsonSlice := range nestedMap {
				if js, ok := jsonSlice.(map[string]any); ok {
					visit(js, jsonFlat, mapResult, mapResultUnits, len(nestedMap), i)
				}
				// else {
				// 	log.Fatal("Weird case ") // should be covered by initially unmarshalling json. coming here would mean json was broken but that should have been checked
				// }
			}
		default: // the value is just a value
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
	}
	return mapResult, mapResultUnits
}

// FromJson function creates a map from a JSON file. Nested data is flattend and keys are connected by a dot.
// In cases where JSON has multiple values, final map's value will be a slice.
func FromJson(jsonPath string, mapAll *map[string]any, mapUnits *map[string]any) error {
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
	jsonFlat, jsonUnits = visit(jsonContent, "", jsonFlat, jsonUnits, 1, 0)

	// Update the provided maps
	for k, v := range jsonFlat {
		(*mapAll)[k] = v
	}
	for k, v := range jsonUnits {
		(*mapUnits)[k] = v
	}

	return nil
}
