package parser

import (
	"fmt"
	"strings"

	cU "github.com/osc-em/oscem-converter-mmcif/converterUtils"
	"golang.org/x/exp/slices"
)

// PDBxItems is PDBxItems map[string][]cU.PDBxItem
//
//	so catDI is []cU.PDBxItem
type Category struct {
	Name           string              // category name
	DataItems      []cU.PDBxItem       // data items according to PDBx dictionary
	ParsedItems    map[string][]string // data category.dataItem and its values from parsed mmCIF; Is a slice to support "loop_" instances
	ParsedKeys     []string            // keys of the map just above
	ParsedLen      int                 // size of slices in ParsedItems (all should be of the same size)
	IsLoop         bool
	DuplicatedFlag bool // to mark that category is duplicated in mmCIF file
	Size           int  // size of arrays from OSCEM json
}

const CategorySeparator = "#"
const LoopKeyword = "loop_"
const ColumnPadding = 5

type MmCIFDocument struct {
	Categories []Category
}

func getKeyByValue(value string, m map[string]string) (string, error) {
	for k, v := range m {
		if v == value {
			return k, nil
		}
	}
	return "", fmt.Errorf("value %v is not in the conversion table", value)
}

// Checks if any other fields reference it as a parent getKeyByValue
func isIdUsedAsParentKey(PDBxItems map[string][]cU.PDBxItem, dataItem cU.PDBxItem) bool {
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

// Sorts categories in a specified order
func getOrderCategories(parsedCategories []string, mmCIFCategories []string) []string {
	var order []string
	allCategories := append(parsedCategories, mmCIFCategories...)
	slices.Sort(allCategories)
	// sort based on the pre-defined (administrative, polymer related entities, ligand (non-polymer) related instances, and structure level description)
	for _, category := range cU.PDBxCategoriesOrder {
		category = "_" + category
		if slices.Contains(allCategories, category) {
			order = append(order, category)
		}
	}
	// add the rest not atom-related in some order (it can be random)
	for _, c := range allCategories {
		if !slices.Contains(order, c) && !slices.Contains(cU.PDBxCategoriesOrderAtom, c[1:]) && !(len(c) > 5 && c[0:5] == "data_") {
			order = append(order, c)
		}

	}
	// add atoms categories
	for _, category := range cU.PDBxCategoriesOrderAtom {
		category = "_" + category
		if slices.Contains(allCategories, category) {
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
