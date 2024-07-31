package parser

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
)

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

func PDBxDict(path string, relevantNames []string) (map[string][]string, map[string]string) {
	dictFile, err := os.Open(path)
	if err != nil {
		log.Fatal("Error while reading the file ", err)
	}
	defer dictFile.Close()
	reSaveDataItem := regexp.MustCompile(`save_[a-zA-Z0-9]+[a-zA-Z0-9]+`)
	reSaveDataItemChild := regexp.MustCompile(`save__([a-zA-Z1-9_.]+)`)
	reUnits := regexp.MustCompile(`_item_units.code`)

	scanner := bufio.NewScanner(dictFile)

	var dataItems = make(map[string][]string)
	var units = make(map[string]string)
	var dataItem string
	var details bool

	//i := 0

	var category string
	var itemsCategory []string

	for scanner.Scan() {
		//i++
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
			for c := range relevantNames {
				if strings.Split(relevantNames[c], ".")[1] == category {
					itemsCategory = append(itemsCategory, category)
					dataItems[dataItem] = itemsCategory
					break
				}
			}
		}
		// once category was grabbed, scan if this category has a specific units defintion
		matchUnits := reUnits.MatchString(scanner.Text())
		if matchUnits {
			units[dataItem+"."+category] = strings.Fields(scanner.Text())[1]
		}
	}
	fmt.Println(dataItems)
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return dataItems, units
}
