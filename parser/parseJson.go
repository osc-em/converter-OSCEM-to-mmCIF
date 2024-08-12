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

func visit(mymap map[string]any, myJsonPath string, myLongProperties map[string]any, lenVal int, valI int) map[string]any {
	keys := converterUtils.GetKeys(mymap)
	for _, k := range keys {
		jsonFlat := keyConcat(myJsonPath, k)
		switch nestedMap := mymap[k].(type) {
		case map[string]any:
			visit(nestedMap, jsonFlat, myLongProperties, lenVal, valI)
		case []any:
			for i, jsonSlice := range nestedMap {
				if js, ok := jsonSlice.(map[string]any); ok {
					visit(js, jsonFlat, myLongProperties, len(nestedMap), i)
				} else {
					log.Fatal("Weird case ")
				}
			}
		default:
			if lenVal == 1 {
				myLongProperties[jsonFlat] = fmt.Sprint(mymap[k])
			} else {
				if valI == 0 {
					values := make([]string, lenVal)
					values[0] = fmt.Sprint(mymap[k])
					myLongProperties[jsonFlat] = values
				} else {
					values := myLongProperties[jsonFlat]
					if v, ok := values.([]string); ok {
						v[valI] = fmt.Sprint(mymap[k])
						myLongProperties[jsonFlat] = v
					}
				}
			}
		}
	}
	return myLongProperties
}

// FromJson function creates a map from a JSON file. Nested data is flattend and keys are connected by a dot.
// In cases where JSON has multiple values, final map's value will be a slice.
func FromJson(jsonPath string) map[string]any {
	// read json file and keep it in form of a map with keys consisting.of.joined.elements
	var data []byte
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		log.Fatal("Error while reading the Json ", err)
	}
	// unmarshal json
	var jsonContent map[string]any
	json.Unmarshal([]byte(data), &jsonContent)

	jsonFlat := make(map[string]any)
	jsonFlat = visit(jsonContent, "", jsonFlat, 1, 0)

	return jsonFlat
}
