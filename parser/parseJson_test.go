package parser

import (
	"fmt"
	"reflect"
	"testing"
)

// test json files are created with chatGPT
func TestFromJson(t *testing.T) {
	values1 := map[string][]string{
		"astronomyData.planets.earth.properties.distanceFromSun.measurement": {"149.6"},
		"astronomyData.planets.earth.properties.radius.measurement":          {"6371"},
		"astronomyData.planets.mars.properties.distanceFromSun.measurement":  {"227.9"},
		"astronomyData.planets.mars.properties.radius.measurement":           {"3389.5"},
		"astronomyData.stars.sun.radius":                                     {"696340"},
		"astronomyData.stars.sun.mass":                                       {"1.989"},
		"astronomyData.stars.proximaCentauri.radius":                         {"200000"},
		"astronomyData.stars.proximaCentauri.mass":                           {"0.123"},
	}
	units1 := map[string][]string{
		"astronomyData.planets.earth.properties.distanceFromSun.measurement": {"million km"},
		"astronomyData.planets.earth.properties.radius.measurement":          {"km"},
		"astronomyData.planets.mars.properties.distanceFromSun.measurement":  {"million km"},
		"astronomyData.planets.mars.properties.radius.measurement":           {"km"},
		"astronomyData.stars.sun.radius":                                     {"km"},
		"astronomyData.stars.sun.mass":                                       {"× 10^30 kg"},
		"astronomyData.stars.proximaCentauri.radius":                         {"km"},
		"astronomyData.stars.proximaCentauri.mass":                           {"× 10^30 kg"},
	}

	values2 := map[string][]string{
		"astronomyData.planets.name":                                   {"Earth", "Mars", "Jupiter"},
		"astronomyData.planets.properties.radius.measurement":          {"6371", "3389.5", "69911"},
		"astronomyData.planets.properties.distanceFromSun.measurement": {"149.6", "227.9", "778.5"},
		"astronomyData.planets.properties.mass.measurement":            {"", "", "1.898"},
		"astronomyData.stars.sun.radius":                               {"696340"},
		"astronomyData.stars.sun.mass":                                 {"1.989"},
		"astronomyData.stars.proximaCentauri.radius":                   {"200000"},
		"astronomyData.stars.proximaCentauri.mass":                     {"0.123"},
		"astronomyData.galaxies.name":                                  {"MilkyWay"},
		"astronomyData.galaxies.chocolate":                             {"true"},
		"astronomyData.galaxies.diameter":                              {"105700"},
		"astronomyData.spaceStation.name":                              {"ISS"},
		"astronomyData.spaceStation.launch":                            {"1995"},
		"astronomyData.spaceStation.agencies.name":                     {"NASA", "Roscosmos", "ESA", "JAXA", "CSA"},
		"astronomyData.spaceStation.agencies.country":                  {"USA", "Russia", "Europe", "Japan", "Canada"},
		"astronomyData.spaceStation.agencies.memeberStates":            {"", "", "22", "", ""},
	}
	units2 := map[string][]string{
		"astronomyData.planets.properties.radius.measurement":          {"km", "km", "km"},
		"astronomyData.planets.properties.distanceFromSun.measurement": {"million km", "million km", "million km"},
		"astronomyData.planets.properties.mass.measurement":            {"", "", "× 10^27 kg"},
		"astronomyData.stars.sun.radius":                               {"km"},
		"astronomyData.stars.sun.mass":                                 {"× 10^30 kg"},
		"astronomyData.stars.proximaCentauri.radius":                   {"km"},
		"astronomyData.stars.proximaCentauri.mass":                     {"× 10^30 kg"},
		"astronomyData.galaxies.diameter":                              {"light years"},
	}

	var tests = []struct {
		name              string
		file              string
		level             string
		expectedValuesMap map[string][]string
		expectedUnitsMap  map[string][]string
		expectedError     string
	}{
		{"no JSON file", "./testData/noJson.json", "", make(map[string][]string, 0), make(map[string][]string, 0), "error while reading the JSON file: open ./testData/noJson.json: no such file or directory"},
		{"broken JSON file: not quoted JSON propoerty", "./testData/badJson.json", "", make(map[string][]string, 0), make(map[string][]string, 0), "error while unmarshaling JSON: invalid character 'r' looking for beginning of object key string"},
		{"nested JSON with values and units at different levels, no arrays in JSON", "./testData/simpleNested.json", "", values1, units1, ""},
		{"nested JSON with values and units at different levels, start at 'earth'", "./testData/simpleNested.json", "earth", map[string][]string{"properties.distanceFromSun.measurement": {"149.6"}, "properties.radius.measurement": {"6371"}}, map[string][]string{"properties.distanceFromSun.measurement": {"million km"}, "properties.radius.measurement": {"km"}}, ""},
		{"nested JSON with values and units at different levels and arrays", "./testData/nestedWithArrays.json", "", values2, units2, ""},
		{"nested JSON with nested arrays ", "./testData/nestedArrays.json", "", map[string][]string{"astronomyData.planets.name": {"Earth", "Mars", "Jupiter"}}, make(map[string][]string, 0), ""},
		{"nested JSON with no nesting after starting with 'chocolate'", "./testData/nestedWithArrays.json", "chocolate", make(map[string][]string, 0), make(map[string][]string, 0), ""},
	}

	for _, tt := range tests {
		testname := fmt.Sprintf("%v", tt.name)

		gotValuesMap := make(map[string][]string, 0)
		gotUnitsMap := make(map[string][]string, 0)
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
