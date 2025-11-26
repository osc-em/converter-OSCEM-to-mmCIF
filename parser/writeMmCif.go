package parser

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	cU "github.com/osc-em/oscem-converter-mmcif/converterUtils"
)

// given a slice of PDBx items get the length of a longest data item name in it
// used for nice formatting of mmCIF files
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

// When existing cif with coordinates is passed,
// for each category in terms of PDBx items, there is a long string with data items and values.
// This method will break blocks of cif text within the same category into PDBxItems and save as struct members
func (c *Category) breakToPDBxItem(cifLongString string) {
	c.ParsedItems = make(map[string][]string)
	lines := strings.Split(cifLongString, "\n")
	if strings.Contains(cifLongString, LoopKeyword) {
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
				if fields[0] == LoopKeyword {
					continue
				} else if fields[0] == "" {
					continue
				}
				c.ParsedKeys = append(c.ParsedKeys, fields[0])
			} else if len(fields) > 1 && indexData == 0 {
				// data entries start, all keys were collected, assign them to map
				for i := range c.ParsedKeys {
					c.ParsedItems[c.ParsedKeys[i]] = []string{fields[i]}
				}
				indexData++
			} else if len(fields) > 1 {
				for i := range c.ParsedKeys {
					c.ParsedItems[c.ParsedKeys[i]] = append(c.ParsedItems[c.ParsedKeys[i]], fields[i])
				}
				indexData++
			}
		}
		c.ParsedLen = indexData
	} else {
		c.ParsedLen = 1
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
			c.ParsedKeys = append(c.ParsedKeys, k)
			c.ParsedItems[k] = []string{strings.ReplaceAll(v, "'", "")}

		}
	}
}

// Reads an entire mmCIF file and organizes it by categories.
// It keeps track of which data container has the most content
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
		if strings.HasPrefix(line, CategorySeparator) || len(strings.Fields(line)) == 0 {
			// break between categories is denoted by # in PDB-related software, Phenix uses an empty line.
			// category ends, appends to the map
			if category != "" {
				mmCIFfields[category] = str.String()
				str.Reset()
				inCategoryFlag = true // record the next category name
			}
		} else {
			if !strings.HasPrefix(line, LoopKeyword) {

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

// handles json cases with array ( or loop_ in mmCIF)
func (c *Category) createCategoryLoop(
	str *strings.Builder,
	nameMapper map[string]string,
	PDBxItems map[string][]cU.PDBxItem,
	OSCEMvalues map[string][]string,
	OSCEMunits map[string][]string,
) {
	var valuesStr strings.Builder
	str.WriteString(LoopKeyword)
	str.WriteString("\n")
	valuesStr.WriteString("")
	var isRelevantID bool
	for _, dataItem := range c.DataItems {
		// In this for-loop we add a new data item name per line
		if c.ParsedLen != 0 && c.Size == c.ParsedLen {
			log.Printf(
				"Category %s exists both in metadata from JSON files and in existing mmCIF file! "+
					"Since your data has many instances of this category, I can't automatically complement "+
					"the instance with data_item %s",
				dataItem.CategoryID, dataItem.Name)
			continue
		}
		fullItemName := dataItem.CategoryID + "." + dataItem.Name
		OSCEMkey, err := getKeyByValue(fullItemName, nameMapper)
		if err != nil {
			// it is an _id property, check if we need it go through all data items
			// and see if it's a parent of another data item, in which case we should add this data item
			isRelevantID = isIdUsedAsParentKey(PDBxItems, dataItem)
			if isRelevantID {
				if _, ok := c.ParsedItems[fullItemName]; !ok {
					fmt.Fprintf(str, "%s\n", fullItemName)
				}
			}
		} else if OSCEMvalues[OSCEMkey] == nil {
			continue // not required and not provided in OSCEM
		} else {
			// writes the data item name
			if _, ok := c.ParsedItems[fullItemName]; !ok {
				fmt.Fprintf(str, "%s\n", fullItemName)
			}
		}
	}

	for v := range c.Size {
		// if this loop we build values one-line-string per instance of array:
		for _, dataItem := range c.DataItems {
			if c.ParsedLen != 0 && c.Size == c.ParsedLen {
				// do not complement
				continue
			}
			fullItemName := dataItem.CategoryID + "." + dataItem.Name
			OSCEMkey, err := getKeyByValue(fullItemName, nameMapper)
			if err != nil {
				// similarly checks for _id properties
				isRelevantID = isIdUsedAsParentKey(PDBxItems, dataItem)
				if isRelevantID {
					if _, ok := c.ParsedItems[fullItemName]; !ok {
						fmt.Fprintf(&valuesStr, "%v ", v+1)
					}
				}
			} else if OSCEMvalues[OSCEMkey] == nil {
				// key was not required and not provided in OSCEM
				continue
			} else if correctSlice, ok := OSCEMvalues[OSCEMkey]; ok {
				// writes the value from OSCEM json
				if _, ok := c.ParsedItems[fullItemName]; !ok {
					if unit, ok := OSCEMunits[OSCEMkey]; ok {
						valueString := checkValue(dataItem, correctSlice[v], OSCEMkey, unit[v])
						fmt.Fprintf(&valuesStr, "%s", valueString)
					} else {
						valueString := checkValue(dataItem, correctSlice[v], OSCEMkey, "")
						fmt.Fprintf(&valuesStr, "%s", valueString)
					}
				}
			}
		}
		// take care of additional info coming from existing mmCIF and add the **values** into the current string
		if c.ParsedLen != 0 && c.Size == c.ParsedLen {
			for _, k := range c.ParsedKeys {
				var value string
				if strings.Contains(c.ParsedItems[k][v], " ") {
					// if name contains whitespaces enclose it in single quotes
					value = fmt.Sprintf("'%s' ", c.ParsedItems[k][v])
				} else {
					value = fmt.Sprintf("%s ", c.ParsedItems[k][v])
				}
				fmt.Fprintf(&valuesStr, "%s", value)
			}
		}
		valuesStr.WriteString("\n")
	}
	if c.ParsedLen != 0 && c.Size == c.ParsedLen {
		// similarly, take care of additional info coming from existing mmCIF and add the **keys** into the current string
		for _, k := range c.ParsedKeys {
			// write in the same order as extracted from the mmCIF
			fmt.Fprintf(str, "%s\n", k)
		}
	}
	str.WriteString(valuesStr.String())
	str.WriteString(CategorySeparator)
	str.WriteString("\n")
}

// handles simple cases of key value pairs
func (c *Category) createCategorySingleValue(
	str *strings.Builder,
	nameMapper map[string]string,
	PDBxItems map[string][]cU.PDBxItem,
	OSCEMvalues map[string][]string,
	OSCEMunits map[string][]string,
) {
	l := getLongestPDBxItem(c.DataItems, cU.GetKeys(c.ParsedItems)) + ColumnPadding
	var isRelevantID bool
	for _, dataItem := range c.DataItems {
		fullItemName := dataItem.CategoryID + "." + dataItem.Name
		OSCEMkey, err := getKeyByValue(dataItem.CategoryID+"."+dataItem.Name, nameMapper)
		if err != nil {
			// it is an _id property, check if we need it go through all data items
			// and see if it's a parent of another data item, in which case we should add this data item
			isRelevantID = isIdUsedAsParentKey(PDBxItems, dataItem)
			if isRelevantID {
				if _, ok := c.ParsedItems[fullItemName]; !ok {
					formatString := fmt.Sprintf("%%-%ds", l)
					fmt.Fprintf(str, formatString, fullItemName)
					fmt.Fprintf(str, "%v", 1)
					str.WriteString("\n")
					continue
				}
			}
		}

		if OSCEMvalues[OSCEMkey] == nil {
			// not required in mmCIF
			continue
		}
		if OSCEMvalue, ok := OSCEMvalues[OSCEMkey]; ok {
			if _, ok := c.ParsedItems[fullItemName]; !ok {
				formatString := fmt.Sprintf("%%-%ds", l)
				fmt.Fprintf(str, formatString, fullItemName)
				if unit, ok := OSCEMunits[OSCEMkey]; ok {
					// the 0th element, because it's the case where only one value is present
					valueString := checkValue(dataItem, OSCEMvalue[0], OSCEMkey, unit[0])
					fmt.Fprintf(str, "%s", valueString)
				} else {
					// values that have no units definition in OSCEM
					valueString := checkValue(dataItem, OSCEMvalue[0], OSCEMkey, "")
					fmt.Fprintf(str, "%s", valueString)

				}
			}
		}

		if _, ok := c.ParsedItems[fullItemName]; !ok {
			str.WriteString("\n")
		}
	}
	if len(c.ParsedItems) != 0 && c.Size == c.ParsedLen {
		// take care of additional info coming from existing mmCIF and add the **item** into the current string
		for _, k := range c.ParsedKeys {
			v := c.ParsedItems[k]
			formatString := fmt.Sprintf("%%-%ds", l)
			fmt.Fprintf(str, formatString, k)
			var value string
			if strings.Contains(v[0], " ") {
				// if name contains whitespaces enclose it in single quotes
				value = fmt.Sprintf("'%s' ", v[0])
			} else {
				value = fmt.Sprintf("%s ", v[0])
			}
			fmt.Fprintf(str, "%s", value)
			str.WriteString("\n")
		}
	}
	str.WriteString(CategorySeparator)
	str.WriteString("\n")
}

// get size of the PDBx category based on OSCEM json values
func (c *Category) getSize(nameMapper map[string]string, OSCEMvalues map[string][]string) {
	// loop through all data items in category,
	// as this reflects the order of data items in PDBx,
	// certain items might not exist in OSCEM json
	// loop until we find first key that exists in json
	for i := range c.DataItems {
		k, err := getKeyByValue(c.DataItems[i].CategoryID+"."+c.DataItems[i].Name, nameMapper)
		if err != nil {
			// occurs when this PDBx category not in the conversions table
			continue
		}
		//check if that key is present in json file and extract it's size
		_, ok := OSCEMvalues[k]
		if ok {
			c.Size = len(OSCEMvalues[k])
			break
		}
	}
}

// data name is the name of data identifier at the top of the new file
// mmCIFCategories is a map created from coordinates file cif. where key is a category name and value is a long string with the text that includes data items, their values and loop_ if applicable
// nameMapper is a map that maps PDBx data items to json keys
// PDBxItems is a map where key is category name and value is a slice of PDBxItems that belong to that category; this comes from PDBx dictionary
// valuesMap is a map where key is json key and value is a slice of strings with values for that key from json metadata
// OSCEMunits is a map where key is json key and value is a slice of strings with units for that key from json metadata
func createCifText(
	dataName string,
	mmCIFCategories map[string]string,
	nameMapper map[string]string,
	PDBxItems map[string][]cU.PDBxItem,
	valuesMap map[string][]string,
	OSCEMunits map[string][]string,
) (string, error) {
	// keeps track of values from JSON that have already been mapped to the PDBx properties
	var str strings.Builder
	//write the data Identifier in the header
	str.WriteString(dataName)
	str.WriteString("\n")
	str.WriteString(CategorySeparator)
	str.WriteString("\n")
	// join and sort all the categories (the parsed ones from existing mmCif file with coordinates and PDBx dictionary)
	parsedCategories := make([]string, 0)
	for k := range PDBxItems {
		parsedCategories = append(parsedCategories, k)
	}
	allCategories := getOrderCategories(parsedCategories, cU.GetKeys(mmCIFCategories))
	// loop through all categories in order to write a file
	for i := range allCategories {
		var c Category
		var ok bool
		c.Name = allCategories[i]
		c.DataItems, ok = PDBxItems[c.Name]
		if ok {
			// check if that category exists in provided mmCIF with coordinates and initialize it accordingly
			_, ok := mmCIFCategories[c.Name]
			if ok {
				c.DuplicatedFlag = true
				log.Printf("Category %s exists both in metadata from JSON files and in existing mmCIF file! Data in mmCIF will be used", c.Name)
				c.breakToPDBxItem(mmCIFCategories[c.Name])
			}
			c.getSize(nameMapper, valuesMap)
			if c.Size != c.ParsedLen && c.ParsedLen != 0 {
				fmt.Fprintf(&str, "%s", mmCIFCategories[c.Name])
				continue
			}
			switch {
			case c.Size > 1:
				c.createCategoryLoop(
					&str, nameMapper, PDBxItems, valuesMap, OSCEMunits,
				)
			case c.Size == 1:
				c.createCategorySingleValue(
					&str, nameMapper, PDBxItems, valuesMap, OSCEMunits)
			default:
				// Based on conversion table, the correspondence in naming between OSCEM and PDBx exist.
				// But for this PDBx data category no OSCEM properties are used in this JSON file.
				continue
			}
		}
		if !c.DuplicatedFlag {
			// this category is not present both in mmCIF and in a new metadata.
			// We won't duplicate it from mmCIF since it was taken from new metadata!
			mmCifLines, ok := mmCIFCategories[c.Name]
			if ok {
				str.WriteString(mmCifLines)
				str.WriteString(CategorySeparator)
				str.WriteString("\n")
			}
		}
	}
	return str.String(), nil
}
