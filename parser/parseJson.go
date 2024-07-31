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

func visit(mymap map[string]any, my_json_path string, my_long_properties map[string]any, lenVal int, valI int) map[string]any {
	keys := converterUtils.GetKeys(mymap)
	for _, k := range keys {
		jsonFlat := keyConcat(my_json_path, k)
		switch nestedMap := mymap[k].(type) {
		case map[string]any:
			visit(nestedMap, jsonFlat, my_long_properties, lenVal, valI)
		case []any:
			for i, jsonSlice := range nestedMap {
				if js, ok := jsonSlice.(map[string]any); ok {
					visit(js, jsonFlat, my_long_properties, len(nestedMap), i)
				} else {
					log.Fatal("Weird case ")
				}
			}
		default:
			if lenVal == 1 {
				my_long_properties[jsonFlat] = fmt.Sprint(mymap[k])
			} else {
				if valI == 0 {
					values := make([]string, lenVal)
					values[0] = fmt.Sprint(mymap[k])
					my_long_properties[jsonFlat] = values
				} else {
					values := my_long_properties[jsonFlat]
					if v, ok := values.([]string); ok {
						v[valI] = fmt.Sprint(mymap[k])
						my_long_properties[jsonFlat] = v
					}
				}
			}
		}
	}
	return my_long_properties
}

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
