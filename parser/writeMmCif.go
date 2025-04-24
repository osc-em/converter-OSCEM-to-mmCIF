package parser

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	cU "github.com/osc-em/converter-OSCEM-to-mmCIF/converterUtils"
	"golang.org/x/exp/slices"
)

func relevantId(PDBxItems map[string][]cU.PDBxItem, dataItem cU.PDBxItem) bool {
	accessValues := PDBxItems[dataItem.CategoryID]
	var parentValue string
	for i := range len(accessValues) {
		for r := range len(accessValues[i].ChildName) {
			if accessValues[i].ChildName[r] == dataItem.CategoryID+"."+dataItem.Name {
				//if accessValues[i].CategoryID == dataItem.CategoryID && accessValues[i].Name == dataItem.Name && len(accessValues[i].ParentName) != 0 {
				parentValue = accessValues[i].ParentName[r]
			}
		}
	}
	if parentValue != "" {
		catValues, ok := PDBxItems[strings.Split(parentValue, ".")[0]]
		if ok {

			for r := range len(catValues) {
				if catValues[r].Name == strings.Split(parentValue, ".")[1] {
					return true
				}
			}
		}
	}
	return false
}

func getKeyByValue(value string, m map[string]string) (string, error) {
	for k, v := range m {
		if v == value {
			return k, nil
		}
	}
	return "", fmt.Errorf("value %v is not in the conversion table", value)
}

// given a slice of PDBx items get the length of a longest data item name in it (because the category is the same)
func getLongestPDBxItem(s []cU.PDBxItem, s2keys []string) int {
	var l int
	for i := range s {
		if len(s[i].Name) > l {
			l = len(s[i].Name)
		}
	}
	for i := range s2keys {
		if len(s2keys[i]) > l {
			l = len(s2keys[i])
		}
	}
	return l + len(s[0].CategoryID) + 1
}

func mmCIFStringToPDBxItem(s string) (map[string][]string, []string, int) {
	cifDI := make(map[string][]string, 0)
	var keys []string
	cifDIlen := 0
	lines := strings.Split(s, "\n")
	if strings.Contains(s, "loop_") {

		indexData := 0
		for _, line := range lines {

			fields := make([]string, 0)
			sqFields := strings.Split(line, "'")
			dqFields := (strings.Split(line, "\""))
			if len(sqFields) != 1 {
				for _, f := range sqFields {
					if f == "" {
						continue
					} else if string(f[0]) == " " || string(f[len(f)-1]) == " " {
						fields = append(fields, strings.Fields(f)...)
					} else {
						fields = append(fields, f)
					}
				}
			} else if len(dqFields) != 1 {
				for _, f := range dqFields {
					if f == "" {
						continue
					} else if string(f[0]) == " " || string(f[len(f)-1]) == " " {
						fields = append(fields, strings.Fields(f)...)
					} else {
						fields = append(fields, f)
					}
				}
			} else {
				fields = strings.Split(line, " ")
			}
			if len(fields) == 1 {
				if fields[0] == "loop_" {
					continue
				} else if fields[0] == "" {
					continue
				}
				keys = append(keys, fields[0])
			} else if len(fields) > 1 && indexData == 0 {
				// data entries start, all keys were collected, assign them to map
				for i := range keys {
					cifDI[keys[i]] = []string{fields[i]}
				}
				indexData++
			} else if len(fields) > 1 {
				for i := range keys {
					cifDI[keys[i]] = append(cifDI[keys[i]], fields[i])
				}
				indexData++
			}
		}
		cifDIlen = indexData
	} else {
		cifDIlen = 1
		var k, v string
		for _, line := range lines {
			if line == "" {
				continue
			}
			fields := strings.Fields(line)
			if len(fields) != 2 {
				k, v = fields[0], strings.Join(fields[1:], " ")
			} else {
				k, v = fields[0], fields[1]
			}
			keys = append(keys, k)
			cifDI[k] = []string{strings.Replace(v, "'", "", -1)}
		}
	}
	return cifDI, keys, cifDIlen
}

func getOrderCategories(parsedCategories []string, mmCIFCategories []string) []string {
	var order []string
	allCategories := append(parsedCategories, mmCIFCategories...)
	slices.Sort(allCategories)
	// sort based on the pre-defined (administrative, polymer related entities, ligand (non-polymer) related instances, and structure level description)
	for _, category := range cU.PDBxCategoriesOrder {
		category = "_" + category
		if cU.SliceContains(allCategories, category) {
			order = append(order, category)
		}
	}
	// add the rest not atom-related in some order (it can be random)
	for _, c := range allCategories {
		if !cU.SliceContains(order, c) && !cU.SliceContains(cU.PDBxCategoriesOrderAtom, c[1:]) && !(len(c) > 5 && c[0:5] == "data_") {
			order = append(order, c)
		}

	}
	// add atoms categories
	for _, category := range cU.PDBxCategoriesOrderAtom {
		category = "_" + category
		if cU.SliceContains(allCategories, category) {
			order = append(order, category)
		}
	}
	// append the rest of "unparsed" categories that were inside their own "data_" containers
	for i := range mmCIFCategories {
		if len(mmCIFCategories[i]) > 5 {
			if mmCIFCategories[i][0:5] == "data_" {
				order = append(order, mmCIFCategories[i])
			}
		}
	}
	return order
}

func parseMmCIF(dictR io.Reader) (string, map[string]string, error) {
	scanner := bufio.NewScanner(dictR)

	var longestData uint32 = 0
	var dataToStr = make(map[string]string, 0)
	var mmCIFfieldsMain map[string]string // should always point to the main model
	var longestNameData string
	// initialize first category
	scanner.Scan()
	firstName := scanner.Text()
	// first run
	longestNameData = firstName
	mmCIFfields, bigString, l, new_data := LoopDataEntry(scanner, firstName)
	mmCIFfieldsMain = mmCIFfields
	dataToStr[firstName] = bigString
	longestData = l

	for new_data != "" {
		mmCIFfields, s, l, d := LoopDataEntry(scanner, new_data)
		dataToStr[new_data] = s
		if l > longestData {
			mmCIFfieldsMain = mmCIFfields
			longestNameData = new_data
		}
		new_data = d
	}

	for k, v := range dataToStr {
		if k != longestNameData {
			// add the data_prefix back
			mmCIFfieldsMain[k] = k + "\n" + v
		}
	}

	return longestNameData, mmCIFfieldsMain, nil
}

// return the categories, same as long string, the length and next category
func LoopDataEntry(scanner *bufio.Scanner, category string) (map[string]string, string, uint32, string) {
	var l uint32 = 0
	var str strings.Builder
	var asText strings.Builder
	inCategoryFlag := true
	var mmCIFfields = make(map[string]string, 0)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data_") {
			category = line
			return mmCIFfields, asText.String(), l, category
		}
		l++
		asText.WriteString(line + "\n")
		if strings.HasPrefix(line, "#") || len(strings.Fields(line)) == 0 {
			// break between categories is denoted by # in PDB-related software, Phenix uses an empty line.
			// category ends, appends to the map
			if category != "" {
				mmCIFfields[category] = str.String()
				str.Reset()
				inCategoryFlag = true // record the next category name
			}
		} else {
			if !strings.HasPrefix(line, "loop_") {

				if inCategoryFlag {
					category = strings.Split(strings.Fields(line)[0], ".")[0]
					inCategoryFlag = false
				}
			}
			str.WriteString(line + "\n")

		}
	}
	return mmCIFfields, asText.String(), l, ""
}

// CreteMetadataCif creates and mmCIF file content as a string. This function only converts metadata in the required format.
// Meant for EMDB depositions where no other existing mmCIF is available
func CreteMetadataCif(nameMapper map[string]string, PDBxItems map[string][]cU.PDBxItem, valuesMap map[string][]string, OSCEMunits map[string][]string) (string, error) {
	return createCifText("data_myID", map[string]string{}, nameMapper, PDBxItems, valuesMap, OSCEMunits)
}

// Given an mmCIF file create a new one with added scientific Metadata.
// Meant for PDB depositions
func SupplementCoordinates(nameMapper map[string]string, PDBxItems map[string][]cU.PDBxItem, valuesMap map[string][]string, OSCEMunits map[string][]string, mmCIFpath io.Reader) (string, error) {
	var dataID string
	var mmCIFvalues map[string]string

	name, categories, err := parseMmCIF(mmCIFpath)
	if err != nil {
		return "", err
	}
	dataID = name
	mmCIFvalues = categories
	return createCifText(dataID, mmCIFvalues, nameMapper, PDBxItems, valuesMap, OSCEMunits)
}

// Given an mmCIF file path, open it and create a new one with added scientific Metadata.
// Meant for PDB depositions
func SupplementCoordinatesFromPath(nameMapper map[string]string, PDBxItems map[string][]cU.PDBxItem, valuesMap map[string][]string, OSCEMunits map[string][]string, mmCIFpath string) (string, error) {
	dictR, err := os.Open(mmCIFpath)
	if err != nil {
		errorString := fmt.Sprintf("mmCIF file %s does not exist!", mmCIFpath)
		return "", errors.New(errorString)
	}
	defer dictR.Close()

	return SupplementCoordinates(nameMapper, PDBxItems, valuesMap, OSCEMunits, dictR)
}

func createCifText(dataName string, mmCIFCategories map[string]string, nameMapper map[string]string, PDBxItems map[string][]cU.PDBxItem, valuesMap map[string][]string, OSCEMunits map[string][]string) (string, error) {
	// keeps track of values from JSON that have already been mapped to the PDBx properties
	var str strings.Builder
	str.WriteString(dataName + "\n#\n") //write the data Identifier in the header

	parsedCategories := make([]string, 0)
	for k := range PDBxItems {
		parsedCategories = append(parsedCategories, k)
	}
	allCategories := getOrderCategories(parsedCategories, cU.GetKeys(mmCIFCategories))

	for _, category := range allCategories {
		var duplicatedFlag bool = false
		catDI, ok := PDBxItems[category]
		cifDIs := map[string][]string{}
		var keys []string
		var cifDIlen int
		if ok {
			_, ok := mmCIFCategories[category]
			if ok {
				duplicatedFlag = true
				log.Printf("Category %s exists both in metadata from JSON files and in existing mmCIF file! Data in mmCIF will be used", category)
				cifDIs, keys, cifDIlen = mmCIFStringToPDBxItem(mmCIFCategories[category])
			}
			//
			var size int
			// need to loop here through all data items in category, as it reflects the order of data items in PDBx, it still might not exist in json
			// loop until we find first key that exists in json
			for i := range catDI {
				k, err := getKeyByValue(catDI[i].CategoryID+"."+catDI[i].Name, nameMapper)
				if err != nil {
					//occurs when this PDBx category not in the conversions table - contains the "id" and used to link data items -> use something else for counting
					continue
				}
				//check if that key is present in json file and extract it's size
				_, ok := valuesMap[k]
				if ok {
					size = len(valuesMap[k])
					break
				}
			}
			if size != cifDIlen && cifDIlen != 0 {
				fmt.Fprintf(&str, "%s", mmCIFCategories[category])
				continue
			}
			switch {
			case size > 1:
				var valuesStr strings.Builder
				str.WriteString("loop_\n")
				valuesStr.WriteString("")
				var isRelevantID bool
				for _, dataItem := range catDI {
					if cifDIlen != 0 && size == cifDIlen {
						log.Printf(
							"Category %s exists both in metadata from JSON files and in existing mmCIF file! "+
								"Since your data has many instances of this category, I can't automatically complement "+
								"the instance with data_item %s",
							dataItem.CategoryID, dataItem.Name)
						continue
					}
					jsonKey, err := getKeyByValue(dataItem.CategoryID+"."+dataItem.Name, nameMapper)
					if err != nil {
						// it is _id property -> check if we need it go through all data items and see if it's a parent somewhere!
						isRelevantID = relevantId(PDBxItems, dataItem)
						if isRelevantID {
							if _, ok := cifDIs[dataItem.CategoryID+"."+dataItem.Name]; !ok {
								fmt.Fprintf(&str, "%s\n", dataItem.CategoryID+"."+dataItem.Name)
							}
						}
					} else if valuesMap[jsonKey] == nil {
						continue // not required and not provided in OSCEM
					} else {
						if _, ok := cifDIs[dataItem.CategoryID+"."+dataItem.Name]; !ok {
							fmt.Fprintf(&str, "%s\n", dataItem.CategoryID+"."+dataItem.Name)
						}
					}
				}

				for v := range size {
					for _, dataItem := range catDI {
						if cifDIlen != 0 && size == cifDIlen {
							// do not complement
							continue
						}
						jsonKey, err := getKeyByValue(dataItem.CategoryID+"."+dataItem.Name, nameMapper)
						if err != nil {
							if isRelevantID {
								if _, ok := cifDIs[dataItem.CategoryID+"."+dataItem.Name]; !ok {
									fmt.Fprintf(&valuesStr, "%v ", v+1)
								}
							}
						} else if valuesMap[jsonKey] == nil {
							continue // key was not required and not provided in OSCEM
						} else if correctSlice, ok := valuesMap[jsonKey]; ok {
							if _, ok := cifDIs[dataItem.CategoryID+"."+dataItem.Name]; !ok {
								if unit, ok := OSCEMunits[jsonKey]; ok {
									valueString := checkValue(dataItem, correctSlice[v], jsonKey, unit[v])
									fmt.Fprintf(&valuesStr, "%s", valueString)
								} else {
									valueString := checkValue(dataItem, correctSlice[v], jsonKey, "")
									fmt.Fprintf(&valuesStr, "%s", valueString)
								}
							}
						}
					}
					if cifDIlen != 0 && size == cifDIlen {
						for _, k := range keys {
							var value string
							if strings.Contains(cifDIs[k][v], " ") {
								value = fmt.Sprintf("'%s' ", cifDIs[k][v]) // if name contains whitespaces enclose it in single quotes
							} else {
								value = fmt.Sprintf("%s ", cifDIs[k][v]) // take value as is
							}
							fmt.Fprintf(&valuesStr, "%s", value)
						}
					}
					valuesStr.WriteString("\n")
				}
				if cifDIlen != 0 && size == cifDIlen {
					for _, k := range keys { // write in the same order as extracted from the mmCIF
						fmt.Fprintf(&str, "%s\n", k)
					}
				}
				str.WriteString(valuesStr.String())
				str.WriteString("#\n")
			case size == 1:
				l := getLongestPDBxItem(catDI, cU.GetKeys(cifDIs)) + 5
				var isRelevantID bool
				for _, dataItem := range catDI {
					jsonKey, err := getKeyByValue(dataItem.CategoryID+"."+dataItem.Name, nameMapper)
					if err != nil {
						// it is _id property -> check if we need it go through all data items and see if it's a parent somewhere!
						isRelevantID = relevantId(PDBxItems, dataItem)
						if isRelevantID {
							// remove the ID entry key from list of parsed from mmCIF we will use one from metadata
							// delete(cifDIs, dataItem.CategoryID+"."+dataItem.Name)
							// keys = cU.DeleteElementFromList(keys, dataItem.CategoryID+"."+dataItem.Name)
							if _, ok := cifDIs[dataItem.CategoryID+"."+dataItem.Name]; !ok {
								formatString := fmt.Sprintf("%%-%ds", l)
								fmt.Fprintf(&str, formatString, dataItem.CategoryID+"."+dataItem.Name)
								fmt.Fprintf(&str, "%v", 1)
								str.WriteString("\n")
								continue
							}
						}
					}

					if valuesMap[jsonKey] == nil {
						continue // not required in mmCIF
					}
					if jsonValue, ok := valuesMap[jsonKey]; ok {
						// remove that key from list of parsed from mmCIF we will use one from metadata
						// delete(cifDIs, dataItem.CategoryID+"."+dataItem.Name)
						// keys = cU.DeleteElementFromList(keys, dataItem.CategoryID+"."+dataItem.Name)
						if _, ok := cifDIs[dataItem.CategoryID+"."+dataItem.Name]; !ok {
							formatString := fmt.Sprintf("%%-%ds", l)
							fmt.Fprintf(&str, formatString, dataItem.CategoryID+"."+dataItem.Name)
							if unit, ok := OSCEMunits[jsonKey]; ok {
								// the 0th element, because it's the case where only one value is present
								valueString := checkValue(dataItem, jsonValue[0], jsonKey, unit[0])
								fmt.Fprintf(&str, "%s", valueString)
							} else {
								// values that have no units definition in OSCEM
								valueString := checkValue(dataItem, jsonValue[0], jsonKey, "")
								fmt.Fprintf(&str, "%s", valueString)

							}
						}
					}

					if _, ok := cifDIs[dataItem.CategoryID+"."+dataItem.Name]; !ok {
						str.WriteString("\n")
					}
				}
				if len(cifDIs) != 0 && size == cifDIlen {
					for _, k := range keys {
						v := cifDIs[k]
						formatString := fmt.Sprintf("%%-%ds", l)
						fmt.Fprintf(&str, formatString, k)
						var value string
						if strings.Contains(v[0], " ") {
							value = fmt.Sprintf("'%s' ", v[0]) // if name contains whitespaces enclose it in single quotes
						} else {
							value = fmt.Sprintf("%s ", v[0]) // take value as is
						}
						fmt.Fprintf(&str, "%s", value)
						str.WriteString("\n")
					}
				}
				str.WriteString("#\n")
			default:
				// Based on conversion table, the correspondence in naming between OSCEM and PDBx exist.
				// But for this PDBx data category not OSCEM properties are used in this JSON file.
				continue
			}
		}
		if !duplicatedFlag {
			// this category is not present both in mmCIF and in a new metadata.
			// We won't duplicate it from mmCIF since it was taken from new metadata!
			mmCifLines, ok := mmCIFCategories[category]
			if ok {
				str.WriteString(mmCifLines)
				str.WriteString("#\n")
			}
		}
	}
	return str.String(), nil
}
