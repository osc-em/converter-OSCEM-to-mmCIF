package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
)

func getKeys[K string, V any](m map[string]V) []string {
	keys := make([]string, len(m))
	i := 0
	for k := range m {
		keys[i] = k
		i++
	}
	return keys
}

func keyConcat(s1, s2 string) string {
	if s1 != "" {
		return s1 + "." + s2
	} else {
		return s2
	}
}

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
		elements = getKeys(mapper)
	}
	var simElem []string
	for i := range elements {
		if itemParent(elements[i]) == item {
			simElem = append(simElem, elements[i])
		}
	}
	return simElem
}
func visit(mymap map[string]any, my_json_path string, my_long_properties map[string]any, lenVal int, valI int) map[string]any {
	keys := getKeys(mymap)
	for _, k := range keys {
		jsonFlat := keyConcat(my_json_path, k)
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
func getKeyByValue(value string, m map[string]string) string {
	for k, v := range m {
		if v == value {
			return k
		}
	}
	return ""
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func getLongest(s []string) int {
	var r int
	for _, a := range s {
		if len(a) > r {
			r = len(a)
		}
	}
	return r
}
func valueMapper(nameMapper map[string]string, valuesMap map[string]any) string {
	// values are mapped to the mmcif properties
	mappedVal := make([]string, 0, len(nameMapper))
	var str strings.Builder
	str.WriteString("data_someID?\n#\n")
	for jsonName := range valuesMap {
		PDBxName := nameMapper[jsonName]
		if PDBxName == "" {
			continue // because translation is iterating on json, it still contains elements that don't exist in mmcif
		}
		if contains(mappedVal, PDBxName) {
			continue
		}
		elements := findElemByItem(itemParent(PDBxName), nameMapper, true)

		if valSlice, ok := valuesMap[jsonName].([]string); ok {
			if valSlice == nil {
				continue // not required in mmCIF
			}
			str.WriteString("loop_\n")

			for _, e := range elements {
				fmt.Fprintf(&str, "%s\n", e)
				mappedVal = append(mappedVal, e)
			}
			for i := range len(valSlice) {
				for _, e := range elements {
					jsonKey := getKeyByValue(e, nameMapper)
					if slice, ok := valuesMap[jsonKey].([]string); ok {
						fmt.Fprintf(&str, "%s ", slice[i])
					} else {
						log.Printf("This element has no multiple values! Possibly %s element is not required in the JSON schema", jsonKey)
					}
				}
				str.WriteString("\n")
			}
			str.WriteString("#\n")
		} else {
			l := getLongest(elements) + 5
			for _, e := range elements {
				jsonKey := getKeyByValue(e, nameMapper)
				if valuesMap[jsonKey] == nil {
					continue // not required in mmCIF
				}
				formatString := fmt.Sprintf("%%-%ds", l)
				fmt.Fprintf(&str, formatString, e)

				fmt.Fprintf(&str, "%s\n", valuesMap[jsonKey])
				mappedVal = append(mappedVal, e)
			}
			str.WriteString("#\n")
		}
	}
	return str.String()
}

func main() {
	jsonInstr := os.Args[1]
	jsonSample := os.Args[2]
	jsonToMmCif := os.Args[3]
	mmCIFpath := os.Args[4]

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

	mmCIFlines := valueMapper(mapper, mapJson)
	//fmt.Println(mmCIFlines)

	// now write to cif file

	// Open the file, create it if it doesn't exist, and truncate it if it does
	file, err := os.OpenFile(mmCIFpath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatal("Error opening file: ", err)
		return
	}
	defer file.Close() // Ensure the file is closed after the operation

	// Write the string to the file
	_, err = file.WriteString(mmCIFlines)
	if err != nil {
		log.Fatal("Error writing to file: ", err)
		return
	}

	fmt.Println("String successfully written to the file.")
}

// go run converter.go ../OSCEM_Schemas/Instrument/test_data_valid.json ../OSCEM_Schemas/Sample/Sample_valid.json /Users/sofya/Documents/openem/converter-JSON-to-mmCIF/mapper.tsv /Users/sofya/Documents/openem/converter-JSON-to-mmCIF/output.txt
