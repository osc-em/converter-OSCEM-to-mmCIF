package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
)

func getKeys[K string, V any](m map[string]V) []string {
	keys := make([]string, 0)
	for k := range m {
		keys = append(keys, k)
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
	mapper := make(map[string]string, 0)
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

func detailLines(line string, details bool) bool {
	if strings.HasPrefix(line, ";") {
		if details {
			details = false
		} else {
			details = true
		}
	}
	return details
}

func parseDict(dictFile *os.File) (map[string][]string, map[string]string) {

	reSaveDataItem := regexp.MustCompile(`save_[a-zA-Z0-9]+[a-zA-Z0-9]+`)
	reSaveDataItemChild := regexp.MustCompile(`save__([a-zA-Z1-9_.]+)`)
	reUnits := regexp.MustCompile(`_item_units.code`)

	scanner := bufio.NewScanner(dictFile)

	var dataItems = make(map[string][]string)
	var units = make(map[string]string)
	var dataItem string
	var details bool

	i := 0

	var category string
	var itemsCategory []string

	for scanner.Scan() {
		i++
		// ignore multi-line comment/detail lines
		details = detailLines(scanner.Text(), details)
		if details {
			continue
		}

		// grab the save__ elements
		matchDataItem := reSaveDataItem.MatchString(scanner.Text())
		if matchDataItem {
			dataItem = strings.Split(scanner.Text(), "save_")[1]
			itemsCategory = make([]string, 0)
		}
		// once dataItem was grabbed scan for category properties within it:
		matchCategory := reSaveDataItemChild.MatchString(scanner.Text())
		if matchCategory {
			category = strings.Split(scanner.Text(), ".")[1]
			itemsCategory = append(itemsCategory, category)
			dataItems[dataItem] = itemsCategory
		}
		// once category was grabbed, scan if this category has a specific units defintion
		matchUnits := reUnits.MatchString(scanner.Text())
		if matchUnits {
			units[dataItem+"."+category] = strings.Fields(scanner.Text())[1]
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return dataItems, units
}

func valueMapper(nameMapper map[string]string, PDBxItems map[string][]string, valuesMap map[string]any) string {
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
		if contains(mappedVal, PDBxName) {
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
							jsonKey := getKeyByValue(e, nameMapper)

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
			l := getLongest(elements) + 5
			for _, oE := range orderedElements {
				for _, e := range elements {
					if oE == strings.Split(e, ".")[1] {

						jsonKey := getKeyByValue(e, nameMapper)
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

func main() {
	jsonInstr := os.Args[1]
	jsonSample := os.Args[2]
	jsonToMmCif := os.Args[3]
	dictFile, err := os.Open(os.Args[4])
	if err != nil {
		log.Fatal(err)
	}
	defer dictFile.Close()

	mmCIFpath := os.Args[5]
	unitsPath := os.Args[6]

	// read all input json files and create one map from them all
	mapInstr := readJson(jsonInstr)
	mapSample := readJson(jsonSample)

	// create one mapping for all the JSON contents: key - value
	mapJson := make(map[string]any, len(mapInstr)+len(mapSample))
	for i := range mapInstr {
		mapJson[i] = mapInstr[i]
	}
	for i := range mapSample {
		mapJson[i] = mapSample[i]
	}

	// create a map containing OSCEM - PDBx naming mappings
	mapper := formatMapper(jsonToMmCif, true)

	// parse PDBx dictionary to retrieve order of data items and units
	dataItems, units := parseDict(dictFile)

	// use only a map of dataItems that will be needed my mapper
	var PDBxdataItems = make(map[string][]string)
	for k, v := range dataItems {
		for _, mV := range mapper {
			if "_"+k == strings.Split(mV, ".")[0] {
				PDBxdataItems[k] = v
				break
			}
		}
	}

	// create mmCIF text
	mmCIFlines := valueMapper(mapper, PDBxdataItems, mapJson)

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

	// Open the file, create it if it doesn't exist, and truncate it if it does
	fileUnits, err := os.OpenFile(unitsPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatal("Error opening file: ", err)
		return
	}
	defer file.Close() // Ensure the file is closed after the operation

	//var unitsString = ""
	var unitsString strings.Builder
	unitsString.WriteString("")
	for k, v := range units {
		fmt.Fprintf(&unitsString, "%s,%s\n", k, v)
	}
	// Write the string to the file
	_, err = fileUnits.WriteString(unitsString.String())
	if err != nil {
		log.Fatal("Error writing to file: ", err)
		return
	}
}

// go run converter.go ../OSCEM_Schemas/Instrument/test_data_valid.json ../OSCEM_Schemas/Sample/Sample_valid.json /Users/sofya/Documents/openem/converter-JSON-to-mmCIF/data/mapper.tsv ./data/mmcif_pdbx_v50.dic /Users/sofya/Documents/openem/converter-JSON-to-mmCIF/results/output.cif /Users/sofya/Documents/openem/converter-JSON-to-mmCIF/results/units.csv
