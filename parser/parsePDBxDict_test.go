package parser

import (
	"converter/converterUtils"
	"fmt"
	"reflect"
	"testing"
)

func equalPDBxItem(a, b converterUtils.PDBxItem) bool {
	return a.CategoryID == b.CategoryID && a.Name == b.Name && a.Unit == b.Unit && a.ValueType == b.ValueType && a.RangeMin == b.RangeMin && a.RangeMax == b.RangeMax && reflect.DeepEqual(a.EnumValues, b.EnumValues) && reflect.DeepEqual(a.PDBxEnumValues, b.PDBxEnumValues)
}
func TestExtractRangeValue(t *testing.T) {
	var tests = []struct {
		name          string
		line          string
		expectedValue string
		expectedError string
	}{
		{"value missing maximum", "_item_range.maximum", "?", ""},
		{"numeric value", "_item_range.minimum  0.0", "0.0", ""},
		{"value emitested", "_item_range.minimum  .", ".", ""},
		{"value missing minimum", "_item_range.minimum  ?", "?", ""},
		{"value not numeric", "_item_range.minimum  *", "?", "value * is not numeric"},
	}

	for _, test := range tests {

		testname := fmt.Sprintf("%v", test.line)
		t.Run(testname, func(t *testing.T) {
			gotValue, gotError := extractRangeValue(test.line)

			if gotError != nil {
				if gotError.Error() != test.expectedError {
					t.Errorf("got error %v, wanted %v", gotError.Error(), test.expectedError)
				}
			}
			if gotValue != test.expectedValue {
				t.Errorf("got %v, want %v", gotValue, test.expectedValue)
			}
		})
	}
}

func TestAssignCategories(t *testing.T) {
	items1 := []converterUtils.PDBxItem{
		{CategoryID: "cat1", Name: "name1", Unit: "u1", ValueType: "float", RangeMin: "0", RangeMax: "3.5", EnumValues: []string{}, PDBxEnumValues: []string{}},
		{CategoryID: "cat1", Name: "name2", Unit: "u1", ValueType: "float", RangeMin: "0", RangeMax: "3.5", EnumValues: []string{}, PDBxEnumValues: []string{}},
		{CategoryID: "cat1", Name: "name3", Unit: "u1", ValueType: "float", RangeMin: "0", RangeMax: "3.5", EnumValues: []string{}, PDBxEnumValues: []string{}}}
	items2 := []converterUtils.PDBxItem{
		{CategoryID: "cat1", Name: "name1", Unit: "u1", ValueType: "float", RangeMin: "0", RangeMax: "3.5", EnumValues: []string{}, PDBxEnumValues: []string{}},
		{CategoryID: "cat1", Name: "name2", Unit: "u1", ValueType: "float", RangeMin: "0", RangeMax: "3.5", EnumValues: []string{}, PDBxEnumValues: []string{}},
		{CategoryID: "cat2", Name: "name3", Unit: "u1", ValueType: "float", RangeMin: "0", RangeMax: "3.5", EnumValues: []string{}, PDBxEnumValues: []string{}},
		{CategoryID: "cat2", Name: "name4", Unit: "u1", ValueType: "float", RangeMin: "0", RangeMax: "3.5", EnumValues: []string{}, PDBxEnumValues: []string{}}}
	items3 := []converterUtils.PDBxItem{
		{CategoryID: "cat1", Name: "name1", Unit: "u1", ValueType: "float", RangeMin: "0", RangeMax: "3.5", EnumValues: []string{}, PDBxEnumValues: []string{}},
		{CategoryID: "cat2", Name: "name2", Unit: "u1", ValueType: "float", RangeMin: "0", RangeMax: "3.5", EnumValues: []string{}, PDBxEnumValues: []string{}},
		{CategoryID: "cat3", Name: "name3", Unit: "u1", ValueType: "float", RangeMin: "0", RangeMax: "3.5", EnumValues: []string{}, PDBxEnumValues: []string{}}}
	var tests = []struct {
		name  string
		items []converterUtils.PDBxItem
		want  map[string][]converterUtils.PDBxItem
	}{
		{"no items", []converterUtils.PDBxItem{}, map[string][]converterUtils.PDBxItem{}},
		{"all of same category", items1, map[string][]converterUtils.PDBxItem{"cat1": items1}},
		{"two categories", items2, map[string][]converterUtils.PDBxItem{"cat1": items2[0:2], "cat2": items2[2:]}},
		{"all different", items3, map[string][]converterUtils.PDBxItem{"cat1": {items3[0]},
			"cat2": {items3[1]},
			"cat3": {items3[2]}}},
	}

	for _, test := range tests {
		testname := fmt.Sprintf("%v", test.name)
		t.Run(testname, func(t *testing.T) {
			ans := AssignPDBxCategories(test.items)
			eq := reflect.DeepEqual(ans, test.want)
			if !eq {
				t.Errorf("got %v, want %v", ans, test.want)
			}
		})
	}
}

func TestPDBxDict(t *testing.T) {
	items1 := []converterUtils.PDBxItem{
		{CategoryID: "_category1", Name: "name1", Unit: "", ValueType: "code", RangeMin: "", RangeMax: "", EnumValues: []string{}, PDBxEnumValues: []string{}},
		{CategoryID: "_category1", Name: "name2", Unit: "", ValueType: "code", RangeMin: "", RangeMax: "", EnumValues: []string{}, PDBxEnumValues: []string{}},
		{CategoryID: "_category2", Name: "name1", Unit: "milliradians", ValueType: "float", RangeMin: "0.0", RangeMax: ".", EnumValues: []string{}, PDBxEnumValues: []string{}},
		{CategoryID: "_category2", Name: "name2", Unit: "", ValueType: "code", RangeMin: "", RangeMax: "", EnumValues: []string{}, PDBxEnumValues: []string{}},
	}
	items2 := []converterUtils.PDBxItem{
		{CategoryID: "_category1", Name: "name1", Unit: "", ValueType: "code", RangeMin: "", RangeMax: "", EnumValues: []string{}, PDBxEnumValues: []string{}},
		{CategoryID: "_category1", Name: "name10", Unit: "", ValueType: "code", RangeMin: "", RangeMax: "", EnumValues: []string{}, PDBxEnumValues: []string{"explicit", "implicit"}},
		{CategoryID: "_category1", Name: "microscope_model", Unit: "", ValueType: "line", RangeMin: "", RangeMax: "", EnumValues: []string{"FEI TECNAI F30", "FEI TECNAI ARCTICA", "FEI TECNAI SPHERA", "FEI TECNAI SPIRIT", "FEI TITAN", "FEI TITAN KRIOS", "FEI/PHILIPS CM10", "FEI/PHILIPS CM12", "FEI/PHILIPS CM120T", "FEI/PHILIPS CM200FEG", "FEI/PHILIPS CM200FEG/SOPHIE", "JEOL CRYO ARM 300", "SIEMENS SULEIKA", "TFS GLACIOS", "TFS KRIOS", "TFS TALOS", "TFS TALOS F200C", "TFS TALOS L120C", "TFS TUNDRA", "ZEISS LEO912"}, PDBxEnumValues: []string{}},
		{CategoryID: "_category1", Name: "name11", Unit: "", ValueType: "text", RangeMin: "", RangeMax: "", EnumValues: []string{}, PDBxEnumValues: []string{}},
	}

	var tests = []struct {
		name           string
		path           string
		names          []string
		expectedValues []converterUtils.PDBxItem
		expectedError  string
	}{
		{"no dict file", "./testData/dictionary.dic", []string{}, []converterUtils.PDBxItem{}, "open ./testData/dictionary.dic: no such file or directory"},
		{"empty dict file", "./testData/emptyDict.dic", []string{}, []converterUtils.PDBxItem{}, ""},
		{"comments only", "./testData/commented.dic", []string{"_category.name"}, []converterUtils.PDBxItem{}, ""},
		{"no intersect with JSON properties", "./testData/sample.dic", []string{"_category.name1"}, []converterUtils.PDBxItem{}, ""},
		{"intersect with JSON properties", "./testData/sample.dic", []string{"_category1.name1", "_category1.name2", "_category2.name1", "_category2.name2"}, items1, ""},
		{"intersect with JSON properties range not numeric", "./testData/sample.dic", []string{"_category1.name_units"}, []converterUtils.PDBxItem{}, "value zero is not numeric"},
		{"enum collection and order of items is mixed", "./testData/sample.dic", []string{"_category1.name1", "_category1.microscope_model", "_category1.name10", "_category1.name11"}, items2, ""},
	}

	for _, test := range tests {
		testname := fmt.Sprintf("%v", test.name)
		t.Run(testname, func(t *testing.T) {

			gotValues, gotError := PDBxDict(test.path, test.names)
			if gotError == nil {
				if len(test.expectedError) > 0 {
					t.Errorf("Expected error: %q, but got no error", test.expectedError)
				} else if len(gotValues) != len(test.expectedValues) {
					t.Errorf("Expected output slice: %v, got: %v", test.expectedValues, gotValues)
				} else {
					for i := range gotValues {
						if !equalPDBxItem(gotValues[i], test.expectedValues[i]) {
							//if !reflect.DeepEqual(gotValues[i], test.expectedValues[i]) {
							t.Errorf("index %v :Expected output slice: %v, got: %v", i, test.expectedValues[i], gotValues[i])
						}
					}
				}
			} else {
				if test.expectedError == "" {
					t.Errorf("Expected no error but got: %q", gotError.Error())
				} else if gotError.Error() != test.expectedError {
					t.Errorf("Expected error: %q, got: %q", test.expectedError, gotError.Error())
				}
			}
		})
	}
}
