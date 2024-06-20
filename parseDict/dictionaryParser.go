package main

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

func main() {
	dictFile, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer dictFile.Close()

	reSaveDataItem := regexp.MustCompile(`save__([a-zA-Z1-9_.]+)`)

	scanner := bufio.NewScanner(dictFile)
	var dataItem string
	var details bool
	i := 0
	for scanner.Scan() {
		i++
		// ignore multi-line comment/detail lines
		details = detailLines(scanner.Text(), details)
		if details {
			continue
		}

		// grab the save__ elements
		matchSaveDataItem := reSaveDataItem.MatchString(scanner.Text())
		if matchSaveDataItem {
			dataItem = strings.Split(scanner.Text(), "save__")[1]
			fmt.Println(dataItem)
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}
