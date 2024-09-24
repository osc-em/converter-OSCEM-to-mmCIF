package parser

import (
	"bufio"
	"converter/converterUtils"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func extractRangeValue(line string) (string, error) {
	if len(strings.Fields(line)) > 1 {
		rangeVal := strings.Fields(line)[1]
		if rangeVal != "." && rangeVal != "?" {
			_, err := strconv.ParseFloat(rangeVal, 64) // ( how to distnguish 0 from +inf? --> happens more often)
			if err != nil {
				// log.Fatal("not a numeric value found to be a range", err)
				return "?", fmt.Errorf("value %v is not numeric", rangeVal)
			}
			return rangeVal, nil
		} else {
			return rangeVal, nil
		}
	}
	return "?", nil
}

func AssignPDBxCategories(dataItems []converterUtils.PDBxItem) map[string][]converterUtils.PDBxItem {
	var itemsInCategory = make(map[string][]converterUtils.PDBxItem)
	for i := range dataItems {
		category := dataItems[i].CategoryID
		val, ok := itemsInCategory[category]
		if ok {
			itemsInCategory[category] = append(val, dataItems[i])
		} else {
			itemsInCategory[category] = []converterUtils.PDBxItem{dataItems[i]}
		}
	}
	return itemsInCategory
}

// PDBxDict parses full dictionary and returns a map, where key is data category name and value is slice of structs ordered the same way as in the dictionary.
// This struct contains relevant properties of a data item in the dictionary.
// PDBx contains a few thousands of data items. For a single experiment done with a certain technique it is redundant no keep track of most of data items as they are highly specific to this technique.
// To avoid that relevantNames argument makes this function record only data items references in the slice.
func PDBxDict(path string, relevantNames []string) ([]converterUtils.PDBxItem, error) {
	var dataItems = make([]converterUtils.PDBxItem, 0)
	dictFile, err := os.Open(path)
	if err != nil {
		return dataItems, err
		//log.Fatal("Error while reading the file ", err)
	}
	defer dictFile.Close()

	reSaveCategory := regexp.MustCompile(`save_[a-zA-Z0-9]+[a-zA-Z0-9]+`)
	reSaveItem := regexp.MustCompile(`save__([a-zA-Z1-9_.]+)`)
	reSaveEnd := regexp.MustCompile(`save_$`)
	//resomeItem := regexp.MustCompile(`_item.`)
	reUnits := regexp.MustCompile(`_item_units.code`)
	reType := regexp.MustCompile(`_item_type.code`)
	reRangeMin := regexp.MustCompile(`_item_range.minimum`)
	reRangeMax := regexp.MustCompile(`_item_range.maximum`)
	reEnum := regexp.MustCompile(`_item_enumeration`) // we will require additional string matching to be sure which tab delimited entry it is
	rePDBxEnum := regexp.MustCompile(`_pdbx_item_enumeration`)
	reSplitEnum := regexp.MustCompile(`[\s]{2,}`)
	reParentName := regexp.MustCompile(`_item_linked.parent_name`)
	reChildName := regexp.MustCompile(`_item_linked.child_name`)
	//reIsIdentifier := regexp.MustCompile(`\.((id)|([a-zA-Z0-9]+_id))`)

	scanner := bufio.NewScanner(dictFile)

	var flagItem = false   // Am I within a PDBx property definition?
	var details bool       // Am I inside of a multi-line comment
	var presentInJson bool // Is this PDBx property present in OSC-EM?

	var comment string

	//var orderedItems []converterUtils.PDBxItem
	var categoryDataItem string

	var category string
	var item string
	var unit string
	var valueType string
	var rangeMinValue string
	var rangeMaxValue string
	var enumValues = make([]string, 0)
	var enumValuesPDBx = make([]string, 0)
	var parentName = make([]string, 0)
	var childName = make([]string, 0)

	enumMatchCount := 0
	enumPDBxMatchCount := 0
	enumParentCount := 0
	enumChildCount := 0
	recordEnumFlag := false
	recordPDBxEnumFlag := false
	recordChildFlag := false
	recordParentFlag := false

	for scanner.Scan() {
		// ignore multi-line comment/detail lines
		if strings.HasPrefix(scanner.Text(), ";") {

			if details {
				details = false
				continue
			} else {
				details = true
			}
		}
		if details {
			comment += scanner.Text()
			continue
		}
		// ignore empty lines - those often come after the multi-line comment lines
		if len(strings.Fields(scanner.Text())) == 0 {
			continue
		}
		matchEnd := reSaveEnd.MatchString(scanner.Text())
		if matchEnd && flagItem {
			flagItem = false
			// processing of single property is done: add to accumulating dict:
			if presentInJson {
				newItem := converterUtils.PDBxItem{
					//categoryItem: categoryDataItem,
					CategoryID:     category,
					Name:           item,
					Unit:           unit,
					ValueType:      valueType,
					RangeMin:       rangeMinValue,
					RangeMax:       rangeMaxValue,
					EnumValues:     enumValues,
					PDBxEnumValues: enumValuesPDBx,
					ParentName:     parentName,
					ChildName:      childName}

				//  reset dataItem property before collecting
				unit = ""
				rangeMinValue = ""
				rangeMaxValue = ""
				enumValues = make([]string, 0)
				enumValuesPDBx = make([]string, 0)
				parentName = make([]string, 0)
				childName = make([]string, 0)

				dataItems = append(dataItems, newItem)
			}
		}
		// grab the save__ elements
		matchCategory := reSaveCategory.MatchString(scanner.Text())

		if matchCategory {
			continue
		}
		// grab the save__ elements
		matchItem := reSaveItem.MatchString(scanner.Text())

		if matchItem {
			flagItem = true
			// extract the category and data item names respectively
			categoryDataItem = strings.Split(scanner.Text(), "save_")[1]
			category = strings.Split(categoryDataItem, ".")[0]
			item = strings.Split(categoryDataItem, ".")[1]

			// only continue if it's relevant for our task or contains an "id"
			for c := range relevantNames {
				if relevantNames[c] == category {
					presentInJson = true
					break
				} else {
					presentInJson = false
				}
			}
			continue
		}
		// once a relevant category and data item were grabbed
		if flagItem && presentInJson {

			// scan if this item has a specific units defintion
			if reUnits.MatchString(scanner.Text()) {
				unit = strings.Fields(scanner.Text())[1]
			}

			// scan if this item has a specific type defintion
			if reType.MatchString(scanner.Text()) {
				valueType = strings.Fields(scanner.Text())[1]
			}

			// .. if value needs to be in certain range
			if reRangeMin.MatchString(scanner.Text()) {
				rangeMinValue, err = extractRangeValue(scanner.Text())
				if err != nil {
					return dataItems, err
				}
			}

			// .. if value needs to be in certain range
			if reRangeMax.MatchString(scanner.Text()) {
				rangeMaxValue, err = extractRangeValue(scanner.Text())
				if err != nil {
					return dataItems, err
				}
			}
			// if linkage to other PDBx terms exist by certain key, extract it:
			if reParentName.MatchString(scanner.Text()) {
				if len(strings.Fields(scanner.Text())) == 2 {
					parentName = []string{strings.Fields(scanner.Text())[1]}
				} else if !recordParentFlag {
					if strings.Fields(scanner.Text())[0] == "_item_linked.parent_name" {
						recordParentFlag = true
						continue
					} else {
						enumParentCount += 1
					}
				} else if recordParentFlag {
					continue
				}

			}
			if reChildName.MatchString(scanner.Text()) {
				if len(strings.Fields(scanner.Text())) == 2 {
					childName = []string{strings.Fields(scanner.Text())[1]}
				} else if !recordChildFlag {
					if strings.Fields(scanner.Text())[0] == "_item_linked.child_name" {
						recordChildFlag = true
						continue
					} else {
						enumChildCount += 1
					}
				} else if recordChildFlag {
					continue
				}

			}
			// .. if enum values are provided (and are not already supposed to be recorded)
			if reEnum.MatchString(scanner.Text()) && !recordEnumFlag && !rePDBxEnum.MatchString((scanner.Text())) {
				if strings.Fields(scanner.Text())[0] == "_item_enumeration.value" {
					recordEnumFlag = true // turn on the flag and start recording from the next one
					continue
				} else {
					enumMatchCount += 1
				}
			} else if reEnum.MatchString(scanner.Text()) && recordEnumFlag {
				continue
			}
			if rePDBxEnum.MatchString((scanner.Text())) && !recordPDBxEnumFlag {
				if strings.Fields(scanner.Text())[0] == "_pdbx_item_enumeration.value" {
					recordPDBxEnumFlag = true // turn on the flag and start recording from the next one
					continue
				} else {
					enumPDBxMatchCount += 1
				}
			} else if rePDBxEnum.MatchString(scanner.Text()) && recordPDBxEnumFlag {
				continue
			}
			if recordEnumFlag {
				if strings.Fields(scanner.Text())[0] == "#" { // end of enum iteration
					recordEnumFlag = false
					enumMatchCount = 0
				} else {
					splittedEnum := reSplitEnum.Split(scanner.Text(), -1)
					valueEnum := splittedEnum[enumMatchCount+1]
					if string(valueEnum[0]) == "\"" {
						valueEnum = valueEnum[1 : len(valueEnum)-1]
					}
					enumValues = append(enumValues, valueEnum)
				}
			}
			if recordPDBxEnumFlag {
				if strings.Fields(scanner.Text())[0] == "#" { // end of enum iteration
					recordPDBxEnumFlag = false
					enumPDBxMatchCount = 0
				} else {
					splittedEnum := reSplitEnum.Split(scanner.Text(), -1)
					valueEnum := splittedEnum[enumPDBxMatchCount+1]
					if string(valueEnum[0]) == "\"" {
						valueEnum = valueEnum[1 : len(valueEnum)-1]
					}
					enumValuesPDBx = append(enumValuesPDBx, valueEnum)
				}
			}
			if recordParentFlag {
				if strings.Fields(scanner.Text())[0] == "#" { // end of enum iteration
					recordParentFlag = false
					enumParentCount = 0
				} else {
					splittedEnum := reSplitEnum.Split(scanner.Text(), -1)
					valueEnum := splittedEnum[enumParentCount+1]
					if string(valueEnum[0]) == "\"" {
						valueEnum = valueEnum[1 : len(valueEnum)-1]
					}
					parentName = append(parentName, valueEnum)
				}
			}
			if recordChildFlag {
				if strings.Fields(scanner.Text())[0] == "#" { // end of enum iteration
					recordChildFlag = false
					enumChildCount = 0
				} else {
					splittedEnum := reSplitEnum.Split(scanner.Text(), -1)
					valueEnum := splittedEnum[enumChildCount+1]
					if string(valueEnum[0]) == "\"" {
						valueEnum = valueEnum[1 : len(valueEnum)-1]
					}
					childName = append(childName, valueEnum)
				}
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return dataItems, err
	}
	return dataItems, nil
}
