package parser

import (
	"fmt"
	"reflect"
	"testing"
)

// test json files are created with chatGPT
func TestFromJson(t *testing.T) {
	values1 := map[string]any{
		"astronomyData.planets.earth.properties.distanceFromSun.measurement": "149.6",
		"astronomyData.planets.earth.properties.radius.measurement":          "6371",
		"astronomyData.planets.mars.properties.distanceFromSun.measurement":  "227.9",
		"astronomyData.planets.mars.properties.radius.measurement":           "3389.5",
		"astronomyData.stars.sun.radius":                                     "696340",
		"astronomyData.stars.sun.mass":                                       "1.989",
		"astronomyData.stars.proximaCentauri.radius":                         "200000",
		"astronomyData.stars.proximaCentauri.mass":                           "0.123",
	}
	units1 := map[string]any{
		"astronomyData.planets.earth.properties.distanceFromSun.measurement": "million km",
		"astronomyData.planets.earth.properties.radius.measurement":          "km",
		"astronomyData.planets.mars.properties.distanceFromSun.measurement":  "million km",
		"astronomyData.planets.mars.properties.radius.measurement":           "km",
		"astronomyData.stars.sun.radius":                                     "km",
		"astronomyData.stars.sun.mass":                                       "× 10^30 kg",
		"astronomyData.stars.proximaCentauri.radius":                         "km",
		"astronomyData.stars.proximaCentauri.mass":                           "× 10^30 kg",
	}

	values2 := map[string]any{
		"astronomyData.planets.name":                                   []string{"Earth", "Mars", "Jupiter"},
		"astronomyData.planets.properties.radius.measurement":          []string{"6371", "3389.5", "69911"},
		"astronomyData.planets.properties.distanceFromSun.measurement": []string{"149.6", "227.9", "778.5"},
		"astronomyData.stars.sun.radius":                               "696340",
		"astronomyData.stars.sun.mass":                                 "1.989",
		"astronomyData.stars.proximaCentauri.radius":                   "200000",
		"astronomyData.stars.proximaCentauri.mass":                     "0.123",
		"astronomyData.galaxies.name":                                  "MilkyWay",
		"astronomyData.galaxies.chocolate":                             "true",
		"astronomyData.galaxies.diameter":                              "105700",
	}
	units2 := map[string]any{
		"astronomyData.planets.properties.radius.measurement":          []string{"km", "km", "km"},
		"astronomyData.planets.properties.distanceFromSun.measurement": []string{"million km", "million km", "million km"},
		"astronomyData.stars.sun.radius":                               "km",
		"astronomyData.stars.sun.mass":                                 "× 10^30 kg",
		"astronomyData.stars.proximaCentauri.radius":                   "km",
		"astronomyData.stars.proximaCentauri.mass":                     "× 10^30 kg",
		"astronomyData.galaxies.diameter":                              "light years",
	}

	var tests = []struct {
		name              string
		file              string
		level             string
		expectedValuesMap map[string]any
		expectedUnitsMap  map[string]any
		expectedError     string
	}{
		{"no JSON file", "./testData/noJson.json", "", make(map[string]any, 0), make(map[string]any, 0), "error while reading the JSON file: open ./testData/noJson.json: no such file or directory"},
		{"broken JSON file: not quoted JSON propoerty", "./testData/badJson.json", "", make(map[string]any, 0), make(map[string]any, 0), "error while unmarshaling JSON: invalid character 'r' looking for beginning of object key string"},
		{"nested JSON with values and units at different levels, no arrays in JSON", "./testData/simpleNested.json", "", values1, units1, ""},
		{"nested JSON with values and units at different levels, start at 'earth'", "./testData/simpleNested.json", "earth", map[string]any{"properties.distanceFromSun.measurement": "149.6", "properties.radius.measurement": "6371"}, map[string]any{"properties.distanceFromSun.measurement": "million km", "properties.radius.measurement": "km"}, ""},
		{"nested JSON with values and units at different levels and arrays", "./testData/nestedWithArrays.json", "", values2, units2, ""},
		{"nested JSON with nested arrays ", "./testData/nestedArrays.json", "", map[string]any{"astronomyData.planets.name": []string{"Earth", "Mars", "Jupiter"}}, make(map[string]any, 0), ""},
		{"nested JSON with no nesting after starting with 'chocolate'", "./testData/nestedWithArrays.json", "chocolate", make(map[string]any, 0), make(map[string]any, 0), ""},
	}

	for _, tt := range tests {
		testname := fmt.Sprintf("%v", tt.name)

		gotValuesMap := make(map[string]any, 0)
		gotUnitsMap := make(map[string]any, 0)

		t.Run(testname, func(t *testing.T) {
			gotError := FromJson(tt.file, &gotValuesMap, &gotUnitsMap, tt.level)
			if gotError != nil {
				if gotError.Error() != tt.expectedError {
					t.Errorf("got error '%v', wanted '%v'", gotError.Error(), tt.expectedError)
				}

			} else {
				if !reflect.DeepEqual(gotValuesMap, tt.expectedValuesMap) {
					t.Errorf("JSON values error: got '%v', want '%v'", gotValuesMap, tt.expectedValuesMap)
				}
				if !reflect.DeepEqual(gotUnitsMap, tt.expectedUnitsMap) {
					t.Errorf("JSON units error: got '%v', want '%v'", gotUnitsMap, tt.expectedUnitsMap)
				}
			}
		})
	}
}
