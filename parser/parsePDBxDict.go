package parser

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type pdbItem struct {
	category   string
	name       string
	unit       string
	rangeMin   float32
	rangeMax   float32
	enumValues []string
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

func PDBxDict(path string, relevantNames []string) map[string][]pdbItem {

	dictFile, err := os.Open(path)
	if err != nil {
		log.Fatal("Error while reading the file ", err)
	}
	defer dictFile.Close()
	reSaveDataItem := regexp.MustCompile(`save_[a-zA-Z0-9]+[a-zA-Z0-9]+`)
	reSaveDataItemChild := regexp.MustCompile(`save__([a-zA-Z1-9_.]+)`)
	reUnits := regexp.MustCompile(`_item_units.code`)
	reRangeMin := regexp.MustCompile(`_item_range.minimum`)
	reRangeMax := regexp.MustCompile(`_item_range.maximum`)
	reEnum := regexp.MustCompile(`_item_enumeration`) // we will require additional string matching to be sure which tab delimited entry it is

	scanner := bufio.NewScanner(dictFile)

	var dataItems = make(map[string][]pdbItem)
	var dataItem string
	var details bool

	var itemsCategory []pdbItem
	var category string
	var categoryDataItem string
	var unit string
	var rangeMinValue float32
	var rangeMaxValue float32
	var enumValues []string

	enumMatchCount := 0
	recordEnumFlag := false

	for scanner.Scan() {
		// ignore multi-line comment/detail lines
		details = detailLines(scanner.Text(), details)
		if details {
			continue
		}

		// grab the save__ elements
		matchDataItem := reSaveDataItem.MatchString(scanner.Text())
		if matchDataItem {
			dataItem = strings.Split(scanner.Text(), "save_")[1]
			itemsCategory = make([]pdbItem, 0)
		}
		// once dataItem was grabbed scan for category properties within it:
		matchCategory := reSaveDataItemChild.MatchString(scanner.Text())
		if matchCategory {
			category = strings.Split(scanner.Text(), "save_")[1]
			categoryDataItem = strings.Split(category, ".")[1]
			for c := range relevantNames {
				if relevantNames[c] == category {
					break
				}
			}
		}
		// once category was grabbed, scan if this category has a specific units defintion
		if reUnits.MatchString(scanner.Text()) {
			unit = strings.Fields(scanner.Text())[1]
		}

		// .. if value needs to be in certain range
		if reRangeMin.MatchString(scanner.Text()) {
			range_val := strings.Fields(scanner.Text())[1]
			if range_val != "." {
				value, err := strconv.ParseFloat(range_val, 32) // ( how to distnguish 0 from nan?)
				if err != nil {
					log.Fatal("not a numeric value found to be a range", err)
				}
				rangeMinValue = float32(value)
			}
		}

		// .. if value needs to be in certain range
		if reRangeMax.MatchString(scanner.Text()) {
			range_val := strings.Fields(scanner.Text())[1]
			if range_val != "." {
				value, err := strconv.ParseFloat(range_val, 32) // ( how to distnguish 0 from +inf? --> happens more often)
				if err != nil {
					log.Fatal("not a numeric value found to be a range", err)
				}
				rangeMaxValue = float32(value)
			}
		}

		// .. if enum values are provided (and are not already supposed to be recorded)
		if reEnum.MatchString(scanner.Text()) && !recordEnumFlag {
			if strings.Fields(scanner.Text())[0] == "_item_enumeration.value" {
				recordEnumFlag = true // turn on the flag and start recording from the next one
				continue
			} else {
				enumMatchCount += 1
			}
		}
		if recordEnumFlag {
			if strings.Fields(scanner.Text())[0] == "#" {
				recordEnumFlag = false
			} else {
				enumValues = append(enumValues, strings.Fields(scanner.Text())[enumMatchCount])
			}
		}

		itemsCategory = append(itemsCategory, pdbItem{category: category,
			name:       categoryDataItem,
			unit:       unit,
			rangeMin:   rangeMinValue,
			rangeMax:   rangeMaxValue,
			enumValues: enumValues})
		dataItems[dataItem] = itemsCategory

		// finally reset units, ranges and enums, because they not appear for all categories
	}
	fmt.Println(dataItems)
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	fmt.Println(dataItems)
	return dataItems
}
