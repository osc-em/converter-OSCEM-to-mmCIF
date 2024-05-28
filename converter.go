package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
)

func get_keys(mymap map[string]any) []string {
	keys := make([]string, len(mymap))
	i := 0
	for k := range mymap {
		keys[i] = k
		i++
	}
	return keys
}
func key_concat(sArray []string, sep string) string {
	s := ""
	for i := range sArray {
		if sArray[i] != "" {
			s += sArray[i]
			if i != 0 {
				s += sep
			}
		}
	}
	return s
}
func visit(mymap map[string]any, my_json_path string, my_long_properties map[string]any, lenVal int, valI int) map[string]any {
	keys := get_keys(mymap)
	for _, k := range keys {
		key := []string{my_json_path, k}
		jsonFlat := key_concat(key, ".")
		if nestedMap, ok := mymap[k].(map[string]any); ok {
			visit(nestedMap, jsonFlat, my_long_properties, lenVal, valI)
		} else if nestedSl, ok := mymap[k].([]any); ok {
			for i, jsonSlice := range nestedSl {
				if js, ok := jsonSlice.(map[string]any); ok {
					visit(js, jsonFlat, my_long_properties, len(nestedSl), i)
				} else {
					log.Fatal("Weirrd case ")
				}
			}
		} else {
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

func readJson(jsonPath string) map[string]any {
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

func formatMapper(path string, toMmCIF bool) map[string]string {
	// read mapper file and based on the task (JSON -> mmCIF or mmCIF -> JSON) create a map with variable names
	file, err := os.Open(path)
	if err != nil {
		log.Fatal("Error while reading the file ", err)
	}
	defer file.Close()
	reader := csv.NewReader(file)
	reader.Comma = '\t'
	records, err := reader.ReadAll()

	if err != nil {
		log.Fatal("Error reading records", err)
	}
	// get the length of json map
	l := 0
	for i := range len(records) {
		if records[i][1] != "" {
			l++
		}
	}
	mapper := make(map[string]string, l)
	if toMmCIF {
		// json properties are mapped to the mmcif properties, missing ones are skipped
		for _, eachrecord := range records {
			if eachrecord[1] != "" {
				mapper[eachrecord[0]] = eachrecord[1]
			}
		}
	} else {
		for _, eachrecord := range records {
			if eachrecord[1] != "" {
				mapper[eachrecord[1]] = eachrecord[0]
			}
		}
	}
	return mapper
}

func valueMapper(nameMapper map[string]string, valuesMap map[string]string) []string {
	// values are mapped to the mmcif properties
	mappedVal := make([]string, len(nameMapper))
	i := 0
	for k := range valuesMap {
		if _, ok := nameMapper[k]; ok {
			//if v, ok2 := valuesMap[k].([]string); ok2 {
			mappedVal[i] = nameMapper[k] + "   " + valuesMap[k]
			i++
			//}
		}
	}
	return mappedVal
}

func main() {
	jsonInstr := os.Args[1]
	jsonSample := os.Args[2]
	jsonToMmCif := os.Args[3]

	mapInstr := readJson(jsonInstr)
	mapSample := readJson(jsonSample)

	mapJson := make(map[string]any, len(mapInstr)+len(mapSample))
	for i := range mapInstr {
		mapJson[i] = mapInstr[i]
	}
	for i := range mapSample {
		mapJson[i] = mapSample[i]
	}

	mapper := formatMapper(jsonToMmCif, true)

	// fix me
	//mmCIFlines := valueMapper(mapper, mapJson)

	// now write to cif file
}

// go run converter.go ../OSCEM_Schemas/Instrument/test_data_valid.json ../OSCEM_Schemas/Sample/Sample_valid.json /Users/sofya/Documents/openem/converter-JSON-to-mmCIF/mapper.tsv
