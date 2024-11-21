package parser

import (
	"fmt"
	"os"
	"testing"

	"github.com/osc-em/converter-OSCEM-to-mmCIF/converterUtils"
)

func TestGetKeyByValue(t *testing.T) {
	var tests = []struct {
		name           string
		value          string
		dictionary     map[string]string
		expectedResult string
		expectedError  string
	}{
		{"value exists in the map", "world", map[string]string{"hello": "world"}, "hello", ""},
		{"value does not exist in a map", "hello", map[string]string{"hello": "world"}, "", "value hello is not in the conversion table"},
	}

	for _, test := range tests {

		testname := fmt.Sprintf("%v", test.name)
		t.Run(testname, func(t *testing.T) {
			gotValue, gotError := getKeyByValue(test.value, test.dictionary)

			if gotError != nil {
				if gotError.Error() != test.expectedError {
					t.Errorf("got error %v, wanted %v", gotError.Error(), test.expectedError)
				}
			}
			if gotValue != test.expectedResult {
				t.Errorf("got %v, want %v", gotValue, test.expectedResult)
			}
		})
	}
}

// func TestSliceContains(t *testing.T) {
// 	var testSlice = []struct {
// 		name           string
// 		slice          []string
// 		element        string
// 		expectedResult bool
// 	}{
// 		{"element in slice", []string{"hello", "world"}, "hello", true},
// 		{"element not in slice", []string{"hello", "world"}, "foo", false},
// 	}

// 	for _, test := range testSlice {

// 		testname := fmt.Sprintf("%v", test.name)
// 		t.Run(testname, func(t *testing.T) {
// 			gotValue := sliceContains(test.slice, test.element)

// 			if gotValue != test.expectedResult {
// 				t.Errorf("got %v, want %v", gotValue, test.expectedResult)
// 			}
// 		})
// 	}
// }

func TestValidateDateIsRFC3339(t *testing.T) {
	var testDate = []struct {
		name           string
		date           string
		expectedResult string
	}{
		{"date is in correct format UTC+3", "2020-12-09T16:09:53+03:00", "2020-12-09"},
		{"date is in correct format UTC+0 Z", "2020-12-09T16:09:53Z", "2020-12-09"},
		{"date is in correct format UTC+0", "2020-12-09T16:09:53+02:00", "2020-12-09"},
		{"date is in correct format UTC-1", "2020-12-09T16:09:53-01:00", "2020-12-09"},
		{"date is in correct format but space instead of T does not work in time.Parse!", "2020-12-09 16:09:53-01:00", ""},
		{"date is in correct format date only but no specified time- does not work in time.Parse!", "2020-12-09", ""},
		{"date is in wrong format", "July, 17th 2017", ""},
	}

	for _, test := range testDate {

		testname := fmt.Sprintf("%v", test.name)
		t.Run(testname, func(t *testing.T) {
			gotValue := validateDateIsRFC3339(test.date)

			if gotValue != test.expectedResult {
				t.Errorf("got %v, want %v", gotValue, test.expectedResult)
			}
		})
	}
}

func TestValidateRange(t *testing.T) {
	var testRange = []struct {
		name           string
		value          string
		dataItem       converterUtils.PDBxItem
		unitOSCEM      string
		nameOSCEM      string
		expectedResult bool
		expectedError  string
	}{
		{"no units defined in OSCEM and PDBx", "27", converterUtils.PDBxItem{CategoryID: "cat1", Name: "name1", Unit: "", ValueType: "float", RangeMin: "0", RangeMax: "300", EnumValues: []string{}, PDBxEnumValues: []string{}}, "", "myOSCEMname", true, ""},
		{"no units defined in OSCEM", "0.7", converterUtils.PDBxItem{CategoryID: "cat1", Name: "name1", Unit: "%", ValueType: "float", RangeMin: "0", RangeMax: "100", EnumValues: []string{}, PDBxEnumValues: []string{}}, "", "myOSCEMname", true, "No units defined for myOSCEMname in OSCEM! Analogous property cat1.name1 in PDBx has % units. Value will still be used in mmCIF file!"},
		{"no units defined in PDBx", "5", converterUtils.PDBxItem{CategoryID: "cat1", Name: "name1", Unit: "", ValueType: "float", RangeMin: "0", RangeMax: "1", EnumValues: []string{}, PDBxEnumValues: []string{}}, "Da", "myOSCEMname", true, "No units defined for cat1.name1 in PDBx! Analogous property myOSCEMname in OSCEM has Da units. Value will still be used in mmCIF file!"},
		{"no explicit name of units defined in OSCEM", "0.5", converterUtils.PDBxItem{CategoryID: "cat1", Name: "name1", Unit: "mol", ValueType: "float", RangeMin: "0", RangeMax: "1", EnumValues: []string{}, PDBxEnumValues: []string{}}, "mM", "myOSCEMname", true, "No explicit unit name is specified for property myOSCEMname in OSCEM, only a short name mM. Value will still be used in mmCIF file!"},
		{"units don't match in OSCEM and PDBx", "2", converterUtils.PDBxItem{CategoryID: "cat1", Name: "name1", Unit: "second", ValueType: "float", RangeMin: "0", RangeMax: "10", EnumValues: []string{}, PDBxEnumValues: []string{}}, "s", "myOSCEMname", true, "Units for analogous properties myOSCEMname in OSCEM and cat1.name1 in PDBx  don't match! Implement a converter from s in OSCEM to second expected by PDBx. Value will still be used in mmCIF file!"},
		{"units match but value not numeric", "zero", converterUtils.PDBxItem{CategoryID: "cat1", Name: "name1", Unit: "electron_volts", ValueType: "float", RangeMin: "-1", RangeMax: "1", EnumValues: []string{}, PDBxEnumValues: []string{}}, "eV", "myOSCEMname", false, "JSON value zero not numeric, but supposed to be"},
		{"units match but value is greater than maximum allowed in PDBx", "10", converterUtils.PDBxItem{CategoryID: "cat1", Name: "name1", Unit: "electron_volts", ValueType: "float", RangeMin: "-1", RangeMax: "1", EnumValues: []string{}, PDBxEnumValues: []string{}}, "eV", "myOSCEMname", false, ""},
		{"units match but value is smaller than minimum allowed in PDBx", "-1.2", converterUtils.PDBxItem{CategoryID: "cat1", Name: "name1", Unit: "electron_volts", ValueType: "float", RangeMin: "-1", RangeMax: "1", EnumValues: []string{}, PDBxEnumValues: []string{}}, "eV", "myOSCEMname", false, ""},
		{"units match and tange undefined by PDBx", "0", converterUtils.PDBxItem{CategoryID: "cat1", Name: "name1", Unit: "electron_volts", ValueType: "float", EnumValues: []string{}, PDBxEnumValues: []string{}}, "eV", "myOSCEMname", true, ""},
		{"units match and value is in range allowed by PDBx", "0", converterUtils.PDBxItem{CategoryID: "cat1", Name: "name1", Unit: "electron_volts", ValueType: "float", RangeMin: "-1", RangeMax: "1", EnumValues: []string{}, PDBxEnumValues: []string{}}, "eV", "myOSCEMname", true, ""},
		{"units match and value is in range allowed by PDBx (no max)", "0", converterUtils.PDBxItem{CategoryID: "cat1", Name: "name1", Unit: "electron_volts", ValueType: "float", RangeMin: "-1", EnumValues: []string{}, PDBxEnumValues: []string{}}, "eV", "myOSCEMname", true, ""},
		{"units match and value is in range allowed by PDBx (no min)", "0", converterUtils.PDBxItem{CategoryID: "cat1", Name: "name1", Unit: "electron_volts", ValueType: "float", RangeMax: "1", EnumValues: []string{}, PDBxEnumValues: []string{}}, "eV", "myOSCEMname", true, ""},
	}

	for _, test := range testRange {

		testname := fmt.Sprintf("%v", test.name)
		t.Run(testname, func(t *testing.T) {
			gotValue, gotError := validateRange(test.value, test.dataItem, test.unitOSCEM, test.nameOSCEM)

			if gotError != nil {
				if gotError.Error() != test.expectedError {
					t.Errorf("got error %v, wanted %v", gotError.Error(), test.expectedError)
				}
			}
			if gotValue != test.expectedResult {
				t.Errorf("got %v, want %v", gotValue, test.expectedResult)
			}
		})
	}
}

func TestValidateEnum(t *testing.T) {
	var testEnums = []struct {
		name           string
		value          string
		dataItem       converterUtils.PDBxItem
		expectedResult string
	}{
		{"boolean true to yes", "true", converterUtils.PDBxItem{CategoryID: "cat1", Name: "name1", EnumValues: []string{"YES", "NO"}, PDBxEnumValues: []string{}}, "YES"},
		{"boolean false to NO", "false", converterUtils.PDBxItem{CategoryID: "cat1", Name: "name1", EnumValues: []string{"YES", "NO"}, PDBxEnumValues: []string{}}, "NO"},
		{"Titan as microscope name (enum in this case irrelevant)", "FEI titan4", converterUtils.PDBxItem{CategoryID: "_em_imaging", Name: "microscope_model", EnumValues: []string{"YES", "NO"}, PDBxEnumValues: []string{}}, "TFS KRIOS"},
		{"Microscope in enum", "my microscope", converterUtils.PDBxItem{CategoryID: "_em_imaging", Name: "microscope_model", EnumValues: []string{"my microscope", "another microscope"}, PDBxEnumValues: []string{}}, "my microscope"},
		{"Microscope not in enum", "my microscope", converterUtils.PDBxItem{CategoryID: "_em_imaging", Name: "microscope_model", EnumValues: []string{"TFS KRIOS", "TFS GLACIOS", "TFS TALOS"}, PDBxEnumValues: []string{}}, "MY MICROSCOPE"},
		{"Microscope in enum", "TFS glacios", converterUtils.PDBxItem{CategoryID: "_em_imaging", Name: "microscope_model", EnumValues: []string{"TFS KRIOS", "TFS GLACIOS", "TFS TALOS"}, PDBxEnumValues: []string{}}, "TFS GLACIOS"},
		{"Rewrite BrightField with space (enum in this case irrelevant)", "BrightField", converterUtils.PDBxItem{CategoryID: "_em_imaging", Name: "mode", EnumValues: []string{"TFS KRIOS", "TFS GLACIOS", "TFS TALOS"}, PDBxEnumValues: []string{}}, "BRIGHT FIELD"},
		{"Bright field case sensitive match PDBx enum", "bright field", converterUtils.PDBxItem{CategoryID: "_em_imaging", Name: "mode", EnumValues: []string{"TFS KRIOS", "TFS GLACIOS", "TFS TALOS"}, PDBxEnumValues: []string{"BRIGHT FIELD", "SOMETHING"}}, "BRIGHT FIELD"},
		{"Rewrite FieldEmission with space (enum in this case irrelevant)", "FieldEmission", converterUtils.PDBxItem{CategoryID: "_em_imaging", Name: "electron_source", EnumValues: []string{}, PDBxEnumValues: []string{}}, "FIELD EMISSION GUN"},
		{"Bright field case sensitive", "Field Emission Gun", converterUtils.PDBxItem{CategoryID: "_em_imaging", Name: "mode", EnumValues: []string{"BRIGHT FIELD", "SOMETHING"}, PDBxEnumValues: []string{"FIELD EMISSION GUN", "SOMETHING"}}, "FIELD EMISSION GUN"},
		{"In enum case sensitive", "NItrogen", converterUtils.PDBxItem{CategoryID: "cat1", Name: "name1", EnumValues: []string{"ETHANE", "NITROGEN", "OTHER"}, PDBxEnumValues: []string{}}, "NITROGEN"},
		{"Grid material matching RegEx on Graphene", "Graphene", converterUtils.PDBxItem{CategoryID: "_em_sample_support", Name: "grid_material", EnumValues: []string{"COPPER", "COPPER/PALLADIUM", "COPPER/RHODIUM", "GOLD", "GRAPHENE OXIDE", "NICKEL", "NICKEL/TITANIUM", "PLATINUM", "SILICON NITRIDE", "TUNGSTEN", "TITANIUM", "MOLYBDENUM"}}, "GRAPHENE OXIDE"},
		{"Grid material matching RegEx on silicon", "silicon", converterUtils.PDBxItem{CategoryID: "_em_sample_support", Name: "grid_material", EnumValues: []string{"COPPER", "COPPER/PALLADIUM", "COPPER/RHODIUM", "GOLD", "GRAPHENE OXIDE", "NICKEL", "NICKEL/TITANIUM", "PLATINUM", "SILICON NITRIDE", "TUNGSTEN", "TITANIUM", "MOLYBDENUM"}}, "SILICON NITRIDE"},
		{"Grid material matching checmical composition 1", "Cu", converterUtils.PDBxItem{CategoryID: "_em_sample_support", Name: "grid_material", EnumValues: []string{"COPPER", "COPPER/PALLADIUM", "COPPER/RHODIUM", "GOLD", "GRAPHENE OXIDE", "NICKEL", "NICKEL/TITANIUM", "PLATINUM", "SILICON NITRIDE", "TUNGSTEN", "TITANIUM", "MOLYBDENUM"}}, "COPPER"},
		{"Grid material matching checmical composition 2", "Cu/Pd", converterUtils.PDBxItem{CategoryID: "_em_sample_support", Name: "grid_material", EnumValues: []string{"COPPER", "COPPER/PALLADIUM", "COPPER/RHODIUM", "GOLD", "GRAPHENE OXIDE", "NICKEL", "NICKEL/TITANIUM", "PLATINUM", "SILICON NITRIDE", "TUNGSTEN", "TITANIUM", "MOLYBDENUM"}}, "COPPER/PALLADIUM"},
		{"Grid material matching checmical composition 3", "Cu/Rh", converterUtils.PDBxItem{CategoryID: "_em_sample_support", Name: "grid_material", EnumValues: []string{"COPPER", "COPPER/PALLADIUM", "COPPER/RHODIUM", "GOLD", "GRAPHENE OXIDE", "NICKEL", "NICKEL/TITANIUM", "PLATINUM", "SILICON NITRIDE", "TUNGSTEN", "TITANIUM", "MOLYBDENUM"}}, "COPPER/RHODIUM"},
		{"Grid material matching checmical composition 4", "Au", converterUtils.PDBxItem{CategoryID: "_em_sample_support", Name: "grid_material", EnumValues: []string{"COPPER", "COPPER/PALLADIUM", "COPPER/RHODIUM", "GOLD", "GRAPHENE OXIDE", "NICKEL", "NICKEL/TITANIUM", "PLATINUM", "SILICON NITRIDE", "TUNGSTEN", "TITANIUM", "MOLYBDENUM"}}, "GOLD"},
		{"Grid material matching checmical composition 5", "Ni", converterUtils.PDBxItem{CategoryID: "_em_sample_support", Name: "grid_material", EnumValues: []string{"COPPER", "COPPER/PALLADIUM", "COPPER/RHODIUM", "GOLD", "GRAPHENE OXIDE", "NICKEL", "NICKEL/TITANIUM", "PLATINUM", "SILICON NITRIDE", "TUNGSTEN", "TITANIUM", "MOLYBDENUM"}}, "NICKEL"},
		{"Grid material matching checmical composition 6", "Ni/Ti", converterUtils.PDBxItem{CategoryID: "_em_sample_support", Name: "grid_material", EnumValues: []string{"COPPER", "COPPER/PALLADIUM", "COPPER/RHODIUM", "GOLD", "GRAPHENE OXIDE", "NICKEL", "NICKEL/TITANIUM", "PLATINUM", "SILICON NITRIDE", "TUNGSTEN", "TITANIUM", "MOLYBDENUM"}}, "NICKEL/TITANIUM"},
		{"Grid material matching checmical composition 7", "Pt", converterUtils.PDBxItem{CategoryID: "_em_sample_support", Name: "grid_material", EnumValues: []string{"COPPER", "COPPER/PALLADIUM", "COPPER/RHODIUM", "GOLD", "GRAPHENE OXIDE", "NICKEL", "NICKEL/TITANIUM", "PLATINUM", "SILICON NITRIDE", "TUNGSTEN", "TITANIUM", "MOLYBDENUM"}}, "PLATINUM"},
		{"Grid material matching checmical composition 8", "W", converterUtils.PDBxItem{CategoryID: "_em_sample_support", Name: "grid_material", EnumValues: []string{"COPPER", "COPPER/PALLADIUM", "COPPER/RHODIUM", "GOLD", "GRAPHENE OXIDE", "NICKEL", "NICKEL/TITANIUM", "PLATINUM", "SILICON NITRIDE", "TUNGSTEN", "TITANIUM", "MOLYBDENUM"}}, "TUNGSTEN"},
		{"Grid material matching checmical composition 9", "Ti", converterUtils.PDBxItem{CategoryID: "_em_sample_support", Name: "grid_material", EnumValues: []string{"COPPER", "COPPER/PALLADIUM", "COPPER/RHODIUM", "GOLD", "GRAPHENE OXIDE", "NICKEL", "NICKEL/TITANIUM", "PLATINUM", "SILICON NITRIDE", "TUNGSTEN", "TITANIUM", "MOLYBDENUM"}}, "TITANIUM"},
		{"Grid material matching checmical composition 10", "Mo", converterUtils.PDBxItem{CategoryID: "_em_sample_support", Name: "grid_material", EnumValues: []string{"COPPER", "COPPER/PALLADIUM", "COPPER/RHODIUM", "GOLD", "GRAPHENE OXIDE", "NICKEL", "NICKEL/TITANIUM", "PLATINUM", "SILICON NITRIDE", "TUNGSTEN", "TITANIUM", "MOLYBDENUM"}}, "MOLYBDENUM"},
		{"Grid material matching enum case sensitive", "COpper", converterUtils.PDBxItem{CategoryID: "_em_sample_support", Name: "grid_material", EnumValues: []string{"COPPER", "COPPER/PALLADIUM", "COPPER/RHODIUM", "GOLD", "GRAPHENE OXIDE", "NICKEL", "NICKEL/TITANIUM", "PLATINUM", "SILICON NITRIDE", "TUNGSTEN", "TITANIUM", "MOLYBDENUM"}}, "COPPER"},
		{"Detector is Falcon I (enum in this case irrelevant)", "falcon 1", converterUtils.PDBxItem{CategoryID: "_em_image_recording", Name: "film_or_detector_model", EnumValues: []string{}, PDBxEnumValues: []string{}}, "FEI FALCON I (4k x 4k)"},
		{"Detector is Falcon II (enum in this case irrelevant)", "Falcon ii", converterUtils.PDBxItem{CategoryID: "_em_image_recording", Name: "film_or_detector_model", EnumValues: []string{}, PDBxEnumValues: []string{}}, "FEI FALCON II (4k x 4k)"},
		{"Detector is Falcon III (enum in this case irrelevant)", "FALCON3 ", converterUtils.PDBxItem{CategoryID: "_em_image_recording", Name: "film_or_detector_model", EnumValues: []string{}, PDBxEnumValues: []string{}}, "FEI FALCON III (4k x 4k)"},
		{"Detector is Falcon IV (enum in this case irrelevant)", "FalconIV", converterUtils.PDBxItem{CategoryID: "_em_image_recording", Name: "film_or_detector_model", EnumValues: []string{}, PDBxEnumValues: []string{}}, "FEI FALCON IV (4k x 4k)"},
		{"Detector is something else from enum", "GATAN multiscan", converterUtils.PDBxItem{CategoryID: "_em_image_recording", Name: "film_or_detector_model", PDBxEnumValues: []string{"GATAN MULTISCAN", "DECTRIS ELA (1k x 0.5k)"}}, "GATAN MULTISCAN"},
		{"Detector is Falcon IV (enum in this case irrelevant)", "FalconIV", converterUtils.PDBxItem{CategoryID: "_em_image_recording", Name: "film_or_detector_model", EnumValues: []string{}, PDBxEnumValues: []string{}}, "FEI FALCON IV (4k x 4k)"},
		{"Funding from enum", "Swiss National Science Foundation", converterUtils.PDBxItem{CategoryID: "_pdbx_audit_support", Name: "funding_organization", PDBxEnumValues: []string{"Swiss National Science Foundation", "Swiss Cancer League"}}, "Swiss National Science Foundation"},
		{"Funding from enum", "SNSF", converterUtils.PDBxItem{CategoryID: "_pdbx_audit_support", Name: "funding_organization", PDBxEnumValues: []string{"Swiss National Science Foundation", "Swiss Cancer League"}}, "Other government"},
	}

	for _, test := range testEnums {

		testname := fmt.Sprintf("%v", test.name)
		t.Run(testname, func(t *testing.T) {
			gotValue := validateEnum(test.value, test.dataItem)

			if gotValue != test.expectedResult {
				t.Errorf("got %v, want %v", gotValue, test.expectedResult)
			}
		})
	}
}

func TestToEMDB(t *testing.T) {
	var testCases = []struct {
		name          string
		namesMap      map[string]string
		PDBxItems     map[string][]converterUtils.PDBxItem
		jsonValues    map[string][]string
		unitsOSCEM    map[string][]string
		expectedText  string
		expectedError string
	}{
		{
			"a data category with no arrays in JSON and valid values",
			map[string]string{"foo.boo": "cat1.name1", "foo.goo": "cat1.name2", "foo.foo": "cat1.name22", "foo.doo": "cat1.name3"},
			map[string][]converterUtils.PDBxItem{
				"cat1": {
					{CategoryID: "cat1", Name: "name1", Unit: "seconds", ValueType: "float", RangeMin: "0", RangeMax: "3.5", EnumValues: []string{}, PDBxEnumValues: []string{}},
					{CategoryID: "cat1", Name: "name2", Unit: "u3", ValueType: "float", RangeMin: "0", RangeMax: "3.5", EnumValues: []string{}, PDBxEnumValues: []string{}},
					{CategoryID: "cat1", Name: "name22", ValueType: "float", RangeMin: "0", RangeMax: "3.5", EnumValues: []string{}, PDBxEnumValues: []string{}},
					{CategoryID: "cat1", Name: "name3", EnumValues: []string{"hello", "world"}},
				},
			},
			map[string][]string{"foo.boo": {"1"}, "foo.goo": {"3.14157"}, "foo.foo": {"2.3"}},
			map[string][]string{"foo.boo": {"s"}, "foo.goo": {"u2"}},
			"data_myID\n#\ncat1.name1      1 \ncat1.name2      3.14157 \ncat1.name22     2.3 \n#\n",
			"",
		},
		{
			"a data category with array in JSON and valid values",
			map[string]string{"foo.boo": "cat1.name1", "foo.goo": "cat1.name2", "foo.foo": "cat1.name22", "foo.doo": "cat1.name3"},
			map[string][]converterUtils.PDBxItem{
				"cat1": {
					{CategoryID: "cat1", Name: "name1", Unit: "seconds", ValueType: "float", RangeMin: "0", RangeMax: "3.5", EnumValues: []string{}, PDBxEnumValues: []string{}},
					{CategoryID: "cat1", Name: "name2", Unit: "u3", ValueType: "float", RangeMin: "0", RangeMax: "3.5", EnumValues: []string{}, PDBxEnumValues: []string{}},
					{CategoryID: "cat1", Name: "name22", ValueType: "float", RangeMin: "0", RangeMax: "3.5", EnumValues: []string{}, PDBxEnumValues: []string{}},
					{CategoryID: "cat1", Name: "name3", EnumValues: []string{"hello", "world"}},
				},
			},
			map[string][]string{"foo.boo": {"1", "1.5"}, "foo.goo": {"3.14157", "2.4"}, "foo.foo": {"2.3", "1.1"}},
			map[string][]string{"foo.boo": {"s", "s"}, "foo.goo": {"u2", "u2"}},
			"data_myID\n#\nloop_\ncat1.name1\ncat1.name2\ncat1.name22\n1 3.14157 2.3 \n1.5 2.4 1.1 \n#\n",
			"",
		},
		{
			"function call to date and that a whole PDBx category is skipped on no JSON values are present for it ",
			map[string]string{"foo.roo": "cat1.name1", "foo.goo": "cat1.name2", "foo.foo": "cat1.name3", "foo.doo": "cat2.name1"},
			map[string][]converterUtils.PDBxItem{
				"cat1": {
					{CategoryID: "cat1", Name: "name1", ValueType: "yyyy-mm-dd"},
					{CategoryID: "cat1", Name: "name2", Unit: "u3", ValueType: "float", RangeMin: "0", RangeMax: "3.5", EnumValues: []string{}, PDBxEnumValues: []string{}},
					{CategoryID: "cat1", Name: "name3", EnumValues: []string{"hello", "world"}},
				},
				"cat2": {
					{CategoryID: "cat2", Name: "name1", ValueType: "float", RangeMin: "0", RangeMax: "3.5", EnumValues: []string{}, PDBxEnumValues: []string{}},
				},
			},
			map[string][]string{"foo.roo": {"2024-03-08T19:11:59+01:00"}},
			map[string][]string{},
			"data_myID\n#\ncat1.name1     2024-03-08 \n#\n",
			"",
		},
		{
			"a non-numeric value for a range",
			map[string]string{"foo.roo": "cat1.name1", "foo.goo": "cat1.name2", "foo.foo": "cat1.name3", "foo.doo": "cat2.name1"},
			map[string][]converterUtils.PDBxItem{
				"cat1": {
					{CategoryID: "cat1", Name: "name1", ValueType: "yyyy-mm-dd"},
					{CategoryID: "cat1", Name: "name2", Unit: "seconds", ValueType: "float", RangeMin: "0", RangeMax: "60", EnumValues: []string{}, PDBxEnumValues: []string{}},
					{CategoryID: "cat1", Name: "name3", EnumValues: []string{"hello", "world"}},
				},
				"cat2": {
					{CategoryID: "cat2", Name: "name1", ValueType: "float", RangeMin: "0", RangeMax: "3.5", EnumValues: []string{}, PDBxEnumValues: []string{}},
				},
			},
			map[string][]string{"foo.goo": {"ten"}},
			map[string][]string{"foo.goo": {"s"}},
			"data_myID\n#\ncat1.name2     ?\n#\n",
			"",
		},
		{
			"function call to enum checkers and single quotes are added when space in the value",
			map[string]string{"foo.roo": "cat1.name1", "foo.goo": "cat1.name2", "foo.foo": "cat1.name3", "foo.doo": "cat2.name1"},
			map[string][]converterUtils.PDBxItem{
				"cat1": {
					{CategoryID: "cat1", Name: "name1", ValueType: "yyyy-mm-dd"},
					{CategoryID: "cat1", Name: "name2", Unit: "u3", ValueType: "float", RangeMin: "0", RangeMax: "3.5", EnumValues: []string{}, PDBxEnumValues: []string{}},
					{CategoryID: "cat1", Name: "name3", EnumValues: []string{"hello world", "foo boo"}},
				},
				"cat2": {
					{CategoryID: "cat2", Name: "name1", ValueType: "float", RangeMin: "0", RangeMax: "3.5", EnumValues: []string{}, PDBxEnumValues: []string{}},
				},
			},
			map[string][]string{"foo.foo": {"hello world"}},
			map[string][]string{},
			"data_myID\n#\ncat1.name3     'hello world' \n#\n",
			"",
		},
		{
			"sign switch for defocus",
			map[string]string{"foo.roo": "cat1.name1", "foo.goo": "cat1.name2", "foo.foo": "cat1.name3", "metadata.defocus.min": "_em_imaging.nominal_defocus_min", "metadata.defocus.max": "_em_imaging.nominal_defocus_max"},
			map[string][]converterUtils.PDBxItem{
				"cat1": {
					{CategoryID: "cat1", Name: "name1", ValueType: "yyyy-mm-dd"},
					{CategoryID: "cat1", Name: "name2", Unit: "u3", ValueType: "float", RangeMin: "0", RangeMax: "3.5", EnumValues: []string{}, PDBxEnumValues: []string{}},
					{CategoryID: "cat1", Name: "name3", EnumValues: []string{"hello world", "foo boo"}},
				},
				"_em_imaging": {
					{CategoryID: "_em_imaging", Name: "nominal_defocus_min", ValueType: "float", RangeMin: "0", RangeMax: "1500", EnumValues: []string{}, PDBxEnumValues: []string{}},
					{CategoryID: "_em_imaging", Name: "nominal_defocus_max", ValueType: "float", RangeMin: "0", RangeMax: "1500", EnumValues: []string{}, PDBxEnumValues: []string{}},
				},
			},
			map[string][]string{"metadata.defocus.min": {"-1100"}, "metadata.defocus.max": {"300"}},
			map[string][]string{},
			"data_myID\n#\n_em_imaging.nominal_defocus_min     1100 \n_em_imaging.nominal_defocus_max     -300 \n#\n",
			"",
		},
	}

	for _, test := range testCases {
		testname := fmt.Sprintf("%v", test.name)
		t.Run(testname, func(t *testing.T) {
			gotValue, gotError := CreteMetadataCif(test.namesMap, test.PDBxItems, test.jsonValues, test.unitsOSCEM)

			if gotError != nil {
				if gotError.Error() != test.expectedError {
					t.Errorf("got error %v, wanted %v", gotError.Error(), test.expectedError)
				}
			}
			if gotValue != test.expectedText {
				t.Errorf("got:\n%v, want:\n%v", gotValue, test.expectedText)

			}
		})
	}
}

func TestToPDB(t *testing.T) {
	var testCases = []struct {
		name          string
		namesMap      map[string]string
		PDBxItems     map[string][]converterUtils.PDBxItem
		jsonValues    map[string][]string
		unitsOSCEM    map[string][]string
		pathExisting  string
		expectedText  string
		expectedError string
	}{
		{
			"no input mmCIF",
			map[string]string{"foo.boo": "cat1.name1", "foo.goo": "cat1.name2", "foo.foo": "cat1.name22", "foo.doo": "cat1.name3"},
			map[string][]converterUtils.PDBxItem{
				"cat1": {
					{CategoryID: "cat1", Name: "name1", Unit: "seconds", ValueType: "float", RangeMin: "0", RangeMax: "3.5", EnumValues: []string{}, PDBxEnumValues: []string{}},
					{CategoryID: "cat1", Name: "name2", Unit: "u3", ValueType: "float", RangeMin: "0", RangeMax: "3.5", EnumValues: []string{}, PDBxEnumValues: []string{}},
					{CategoryID: "cat1", Name: "name22", ValueType: "float", RangeMin: "0", RangeMax: "3.5", EnumValues: []string{}, PDBxEnumValues: []string{}},
					{CategoryID: "cat1", Name: "name3", EnumValues: []string{"hello", "world"}},
				},
			},
			map[string][]string{"foo.boo": {"1"}, "foo.goo": {"3.14157"}, "foo.foo": {"2.3"}},
			map[string][]string{"foo.boo": {"s"}, "foo.goo": {"u2"}},
			"someFile.cif",
			"",
			"mmCIF file someFile.cif does not exist!",
		},
		{
			"valid input mmCIF",
			map[string]string{"foo.boo": "cat1.name1", "foo.goo": "cat1.name2", "foo.foo": "cat1.name22", "foo.doo": "cat1.name3"},
			map[string][]converterUtils.PDBxItem{
				"cat1": {
					{CategoryID: "cat1", Name: "name1", Unit: "seconds", ValueType: "float", RangeMin: "0", RangeMax: "3.5", EnumValues: []string{}, PDBxEnumValues: []string{}},
					{CategoryID: "cat1", Name: "name2", Unit: "u3", ValueType: "float", RangeMin: "0", RangeMax: "3.5", EnumValues: []string{}, PDBxEnumValues: []string{}},
					{CategoryID: "cat1", Name: "name22", ValueType: "float", RangeMin: "0", RangeMax: "3.5", EnumValues: []string{}, PDBxEnumValues: []string{}},
					{CategoryID: "cat1", Name: "name3", EnumValues: []string{"hello", "world"}},
				},
			},
			map[string][]string{"foo.boo": {"1"}, "foo.goo": {"3.14157"}, "foo.foo": {"2.3"}},
			map[string][]string{"foo.boo": {"s"}, "foo.goo": {"u2"}},
			"testData/example.cif",
			"data_K3DAK4\n#\nloop_\n_citation.id\n_citation.title\n_citation.journal_abbrev\n_citation.journal_volume\n_citation.page_first\n_citation.page_last\n_citation.year\n_citation.journal_id_ASTM\n_citation.journal_id_ISSN\n_citation.journal_id_CSD\nphenix.real_space_refine 'Real-space refinement in PHENIX for cryo-EM and crystallography' 'Acta Crystallogr., Sect. D: Biol. Crystallogr.' 74 531 544 2018 ABCRE6 0907-4449 0766\n#\nloop_\n_chem_comp.id\nALA\nARG\nASN\nASP\nCYS\nGLN\nGLU\nGLY\nHIS\nILE\nLEU\nLYS\nMET\nPHE\nPRO\nSER\nTHR\nTRP\nTYR\nVAL\n#\nloop_\n_software.pdbx_ordinal\n_software.name\n_software.version\n_software.type\n_software.contact_author\n_software.contact_author_email\n_software.location\n_software.classification\n_software.citation_id\n_software.language\n1 phenix.real_space_refine 1.20rc4_4425 program 'Pavel Afonine' pafonine@lbl.gov https://www.phenix-online.org/ refinement phenix.real_space_refine Python/C++\n1 Phenix 1.20rc4_4425 program 'Paul D. Adams' pdadams@lbl.gov https://www.phenix-online.org/ refinement phenix Python/C++\n#\nloop_\n_space_group_symop.id\n_space_group_symop.operation_xyz\n1 x,y,z\n#\ncat1.name1      1 \ncat1.name2      3.14157 \ncat1.name22     2.3 \n#\nloop_\n_atom_site.group_PDB\n_atom_site.id\n_atom_site.label_atom_id\n_atom_site.label_alt_id\n_atom_site.label_comp_id\n_atom_site.auth_asym_id\n_atom_site.auth_seq_id\n_atom_site.pdbx_PDB_ins_code\n_atom_site.Cartn_x\n_atom_site.Cartn_y\n_atom_site.Cartn_z\n_atom_site.occupancy\n_atom_site.B_iso_or_equiv\n_atom_site.type_symbol\n_atom_site.pdbx_formal_charge\n_atom_site.label_asym_id\n_atom_site.label_entity_id\n_atom_site.label_seq_id\n_atom_site.pdbx_PDB_model_num\nATOM 1 N . SER B 535 ? 270.43781 345.22081 281.42585 1.000 465.59921 N ? A ? 1 1\nATOM 2 CA . SER B 535 ? 270.07764 346.63283 281.41226 1.000 465.59921 C ? A ? 1 1\nATOM 3 C . SER B 535 ? 268.70572 346.84369 282.04532 1.000 465.59921 C ? A ? 1 1\nATOM 4 O . SER B 535 ? 268.12252 345.92083 282.61134 1.000 465.59921 O ? A ? 1 1\nATOM 5 CB . SER B 535 ? 270.08494 347.17423 279.98114 1.000 465.59921 C ? A ? 1 1\nATOM 6 OG . SER B 535 ? 271.31962 346.90634 279.34048 1.000 465.59921 O ? A ? 1 1\nATOM 7 N . VAL B 536 ? 268.19595 348.07334 281.94291 1.000 465.59921 N ? A ? 2 1\nATOM 8 CA . VAL B 536 ? 266.88271 348.37783 282.50166 1.000 465.59921 C ? A ? 2 1\nATOM 9 C . VAL B 536 ? 265.76647 347.61528 281.79180 1.000 465.59921 C ? A ? 2 1\nATOM 10 O . VAL B 536 ? 265.89177 347.21437 280.62991 1.000 465.59921 O ? A ? 2 1\nATOM 11 CB . VAL B 536 ? 266.60814 349.89588 282.55950 1.000 465.59921 C ? A ? 2 1\nATOM 12 CG1 . VAL B 536 ? 266.47176 350.50312 281.15572 1.000 465.59921 C ? A ? 2 1\nATOM 13 CG2 . VAL B 536 ? 265.37232 350.19041 283.40137 1.000 465.59921 C ? A ? 2 1\nATOM 14 N . VAL B 537 ? 264.66075 347.41139 282.50950 1.000 465.59921 N ? A ? 3 1\nATOM 15 CA . VAL B 537 ? 263.53325 346.68375 281.94275 1.000 465.59921 C ? A ? 3 1\nATOM 16 C . VAL B 537 ? 263.01065 347.41204 280.70782 1.000 465.59921 C ? A ? 3 1\nATOM 17 O . VAL B 537 ? 263.07692 348.64465 280.60360 1.000 465.59921 O ? A ? 3 1\nATOM 18 CB . VAL B 537 ? 262.41900 346.51315 282.98883 1.000 465.59921 C ? A ? 3 1\nATOM 19 CG1 . VAL B 537 ? 261.74793 347.85051 283.27814 1.000 465.59921 C ? A ? 3 1\nATOM 20 CG2 . VAL B 537 ? 261.39935 345.49031 282.51421 1.000 465.59921 C ? A ? 3 1\n#\nloop_\n_atom_site_anisotrop.id\n_atom_site_anisotrop.pdbx_auth_atom_id\n_atom_site_anisotrop.pdbx_label_alt_id\n_atom_site_anisotrop.pdbx_auth_comp_id\n_atom_site_anisotrop.pdbx_auth_asym_id\n_atom_site_anisotrop.pdbx_auth_seq_id\n_atom_site_anisotrop.pdbx_PDB_ins_code\n_atom_site_anisotrop.U[1][1]\n_atom_site_anisotrop.U[2][2]\n_atom_site_anisotrop.U[3][3]\n_atom_site_anisotrop.U[1][2]\n_atom_site_anisotrop.U[1][3]\n_atom_site_anisotrop.U[2][3]\n1 N . SER B 535 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n2 CA . SER B 535 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n3 C . SER B 535 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n4 O . SER B 535 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n5 CB . SER B 535 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n6 OG . SER B 535 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n7 N . VAL B 536 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n8 CA . VAL B 536 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n9 C . VAL B 536 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n10 O . VAL B 536 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n11 CB . VAL B 536 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n12 CG1 . VAL B 536 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n13 CG2 . VAL B 536 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n14 N . VAL B 537 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n15 CA . VAL B 537 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n16 C . VAL B 537 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n17 O . VAL B 537 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n18 CB . VAL B 537 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n19 CG1 . VAL B 537 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n20 CG2 . VAL B 537 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n#\n",
			"",
		},
		{
			"valid input mmCIF, value present in both mmCIF and metadata",
			map[string]string{"foo.boo": "cat1.name1", "foo.goo": "cat1.name2", "foo.foo": "cat1.name22", "foo.doo": "cat1.name3", "mydata1": "_space_group_symop.id", "mydata2": "_space_group_symop.operation_xyz"},
			map[string][]converterUtils.PDBxItem{
				"cat1": {
					{CategoryID: "cat1", Name: "name1", Unit: "seconds", ValueType: "float", RangeMin: "0", RangeMax: "3.5", EnumValues: []string{}, PDBxEnumValues: []string{}},
					{CategoryID: "cat1", Name: "name2", Unit: "u3", ValueType: "float", RangeMin: "0", RangeMax: "3.5", EnumValues: []string{}, PDBxEnumValues: []string{}},
					{CategoryID: "cat1", Name: "name22", ValueType: "float", RangeMin: "0", RangeMax: "3.5", EnumValues: []string{}, PDBxEnumValues: []string{}},
					{CategoryID: "cat1", Name: "name3", EnumValues: []string{"hello", "world"}},
				},
				"_space_group_symop": {
					{CategoryID: "_space_group_symop", Name: "id", ValueType: "float", RangeMin: "0", RangeMax: "3.5"},
					{CategoryID: "_space_group_symop", Name: "operation_xyz"},
				},
			},
			map[string][]string{"foo.boo": {"1"}, "foo.goo": {"3.14157"}, "foo.foo": {"2.3"}, "mydata1": {"1", "2", "3"}, "mydata2": {"x,y,z", "x,y,z", "x,y,z"}},
			map[string][]string{},
			"testData/example.cif",
			"data_K3DAK4\n#\nloop_\n_citation.id\n_citation.title\n_citation.journal_abbrev\n_citation.journal_volume\n_citation.page_first\n_citation.page_last\n_citation.year\n_citation.journal_id_ASTM\n_citation.journal_id_ISSN\n_citation.journal_id_CSD\nphenix.real_space_refine 'Real-space refinement in PHENIX for cryo-EM and crystallography' 'Acta Crystallogr., Sect. D: Biol. Crystallogr.' 74 531 544 2018 ABCRE6 0907-4449 0766\n#\nloop_\n_chem_comp.id\nALA\nARG\nASN\nASP\nCYS\nGLN\nGLU\nGLY\nHIS\nILE\nLEU\nLYS\nMET\nPHE\nPRO\nSER\nTHR\nTRP\nTYR\nVAL\n#\nloop_\n_software.pdbx_ordinal\n_software.name\n_software.version\n_software.type\n_software.contact_author\n_software.contact_author_email\n_software.location\n_software.classification\n_software.citation_id\n_software.language\n1 phenix.real_space_refine 1.20rc4_4425 program 'Pavel Afonine' pafonine@lbl.gov https://www.phenix-online.org/ refinement phenix.real_space_refine Python/C++\n1 Phenix 1.20rc4_4425 program 'Paul D. Adams' pdadams@lbl.gov https://www.phenix-online.org/ refinement phenix Python/C++\n#\nloop_\n_space_group_symop.id\n_space_group_symop.operation_xyz\n1 x,y,z \n2 x,y,z \n3 x,y,z \n#\ncat1.name1      1 \ncat1.name2      3.14157 \ncat1.name22     2.3 \n#\nloop_\n_atom_site.group_PDB\n_atom_site.id\n_atom_site.label_atom_id\n_atom_site.label_alt_id\n_atom_site.label_comp_id\n_atom_site.auth_asym_id\n_atom_site.auth_seq_id\n_atom_site.pdbx_PDB_ins_code\n_atom_site.Cartn_x\n_atom_site.Cartn_y\n_atom_site.Cartn_z\n_atom_site.occupancy\n_atom_site.B_iso_or_equiv\n_atom_site.type_symbol\n_atom_site.pdbx_formal_charge\n_atom_site.label_asym_id\n_atom_site.label_entity_id\n_atom_site.label_seq_id\n_atom_site.pdbx_PDB_model_num\nATOM 1 N . SER B 535 ? 270.43781 345.22081 281.42585 1.000 465.59921 N ? A ? 1 1\nATOM 2 CA . SER B 535 ? 270.07764 346.63283 281.41226 1.000 465.59921 C ? A ? 1 1\nATOM 3 C . SER B 535 ? 268.70572 346.84369 282.04532 1.000 465.59921 C ? A ? 1 1\nATOM 4 O . SER B 535 ? 268.12252 345.92083 282.61134 1.000 465.59921 O ? A ? 1 1\nATOM 5 CB . SER B 535 ? 270.08494 347.17423 279.98114 1.000 465.59921 C ? A ? 1 1\nATOM 6 OG . SER B 535 ? 271.31962 346.90634 279.34048 1.000 465.59921 O ? A ? 1 1\nATOM 7 N . VAL B 536 ? 268.19595 348.07334 281.94291 1.000 465.59921 N ? A ? 2 1\nATOM 8 CA . VAL B 536 ? 266.88271 348.37783 282.50166 1.000 465.59921 C ? A ? 2 1\nATOM 9 C . VAL B 536 ? 265.76647 347.61528 281.79180 1.000 465.59921 C ? A ? 2 1\nATOM 10 O . VAL B 536 ? 265.89177 347.21437 280.62991 1.000 465.59921 O ? A ? 2 1\nATOM 11 CB . VAL B 536 ? 266.60814 349.89588 282.55950 1.000 465.59921 C ? A ? 2 1\nATOM 12 CG1 . VAL B 536 ? 266.47176 350.50312 281.15572 1.000 465.59921 C ? A ? 2 1\nATOM 13 CG2 . VAL B 536 ? 265.37232 350.19041 283.40137 1.000 465.59921 C ? A ? 2 1\nATOM 14 N . VAL B 537 ? 264.66075 347.41139 282.50950 1.000 465.59921 N ? A ? 3 1\nATOM 15 CA . VAL B 537 ? 263.53325 346.68375 281.94275 1.000 465.59921 C ? A ? 3 1\nATOM 16 C . VAL B 537 ? 263.01065 347.41204 280.70782 1.000 465.59921 C ? A ? 3 1\nATOM 17 O . VAL B 537 ? 263.07692 348.64465 280.60360 1.000 465.59921 O ? A ? 3 1\nATOM 18 CB . VAL B 537 ? 262.41900 346.51315 282.98883 1.000 465.59921 C ? A ? 3 1\nATOM 19 CG1 . VAL B 537 ? 261.74793 347.85051 283.27814 1.000 465.59921 C ? A ? 3 1\nATOM 20 CG2 . VAL B 537 ? 261.39935 345.49031 282.51421 1.000 465.59921 C ? A ? 3 1\n#\nloop_\n_atom_site_anisotrop.id\n_atom_site_anisotrop.pdbx_auth_atom_id\n_atom_site_anisotrop.pdbx_label_alt_id\n_atom_site_anisotrop.pdbx_auth_comp_id\n_atom_site_anisotrop.pdbx_auth_asym_id\n_atom_site_anisotrop.pdbx_auth_seq_id\n_atom_site_anisotrop.pdbx_PDB_ins_code\n_atom_site_anisotrop.U[1][1]\n_atom_site_anisotrop.U[2][2]\n_atom_site_anisotrop.U[3][3]\n_atom_site_anisotrop.U[1][2]\n_atom_site_anisotrop.U[1][3]\n_atom_site_anisotrop.U[2][3]\n1 N . SER B 535 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n2 CA . SER B 535 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n3 C . SER B 535 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n4 O . SER B 535 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n5 CB . SER B 535 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n6 OG . SER B 535 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n7 N . VAL B 536 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n8 CA . VAL B 536 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n9 C . VAL B 536 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n10 O . VAL B 536 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n11 CB . VAL B 536 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n12 CG1 . VAL B 536 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n13 CG2 . VAL B 536 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n14 N . VAL B 537 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n15 CA . VAL B 537 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n16 C . VAL B 537 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n17 O . VAL B 537 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n18 CB . VAL B 537 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n19 CG1 . VAL B 537 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n20 CG2 . VAL B 537 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n#\n",
			"",
		},
	}

	for _, test := range testCases {
		testname := fmt.Sprintf("%v", test.name)
		t.Run(testname, func(t *testing.T) {
			gotValue, gotError := SupplementCoordinatesFromPath(test.namesMap, test.PDBxItems, test.jsonValues, test.unitsOSCEM, test.pathExisting)

			if gotError != nil {
				if gotError.Error() != test.expectedError {
					t.Errorf("got error %v, wanted %v", gotError.Error(), test.expectedError)
				}
			}
			if gotValue != test.expectedText {
				t.Errorf("got:\n%v, want:\n%v", gotValue, test.expectedText)

			}
		})
	}
}

func TestToPDB2(t *testing.T) {
	var testCases = []struct {
		name          string
		namesMap      map[string]string
		PDBxItems     map[string][]converterUtils.PDBxItem
		jsonValues    map[string][]string
		unitsOSCEM    map[string][]string
		pathExisting  string
		expectedText  string
		expectedError string
	}{
		{
			"no input mmCIF",
			map[string]string{"foo.boo": "cat1.name1", "foo.goo": "cat1.name2", "foo.foo": "cat1.name22", "foo.doo": "cat1.name3"},
			map[string][]converterUtils.PDBxItem{
				"cat1": {
					{CategoryID: "cat1", Name: "name1", Unit: "seconds", ValueType: "float", RangeMin: "0", RangeMax: "3.5", EnumValues: []string{}, PDBxEnumValues: []string{}},
					{CategoryID: "cat1", Name: "name2", Unit: "u3", ValueType: "float", RangeMin: "0", RangeMax: "3.5", EnumValues: []string{}, PDBxEnumValues: []string{}},
					{CategoryID: "cat1", Name: "name22", ValueType: "float", RangeMin: "0", RangeMax: "3.5", EnumValues: []string{}, PDBxEnumValues: []string{}},
					{CategoryID: "cat1", Name: "name3", EnumValues: []string{"hello", "world"}},
				},
			},
			map[string][]string{"foo.boo": {"1"}, "foo.goo": {"3.14157"}, "foo.foo": {"2.3"}},
			map[string][]string{"foo.boo": {"s"}, "foo.goo": {"u2"}},
			"someFile.cif",
			"",
			"mmCIF file someFile.cif does not exist!",
		},
		{
			"valid input mmCIF",
			map[string]string{"foo.boo": "cat1.name1", "foo.goo": "cat1.name2", "foo.foo": "cat1.name22", "foo.doo": "cat1.name3"},
			map[string][]converterUtils.PDBxItem{
				"cat1": {
					{CategoryID: "cat1", Name: "name1", Unit: "seconds", ValueType: "float", RangeMin: "0", RangeMax: "3.5", EnumValues: []string{}, PDBxEnumValues: []string{}},
					{CategoryID: "cat1", Name: "name2", Unit: "u3", ValueType: "float", RangeMin: "0", RangeMax: "3.5", EnumValues: []string{}, PDBxEnumValues: []string{}},
					{CategoryID: "cat1", Name: "name22", ValueType: "float", RangeMin: "0", RangeMax: "3.5", EnumValues: []string{}, PDBxEnumValues: []string{}},
					{CategoryID: "cat1", Name: "name3", EnumValues: []string{"hello", "world"}},
				},
			},
			map[string][]string{"foo.boo": {"1"}, "foo.goo": {"3.14157"}, "foo.foo": {"2.3"}},
			map[string][]string{"foo.boo": {"s"}, "foo.goo": {"u2"}},
			"testData/example.cif",
			"data_K3DAK4\n#\nloop_\n_citation.id\n_citation.title\n_citation.journal_abbrev\n_citation.journal_volume\n_citation.page_first\n_citation.page_last\n_citation.year\n_citation.journal_id_ASTM\n_citation.journal_id_ISSN\n_citation.journal_id_CSD\nphenix.real_space_refine 'Real-space refinement in PHENIX for cryo-EM and crystallography' 'Acta Crystallogr., Sect. D: Biol. Crystallogr.' 74 531 544 2018 ABCRE6 0907-4449 0766\n#\nloop_\n_chem_comp.id\nALA\nARG\nASN\nASP\nCYS\nGLN\nGLU\nGLY\nHIS\nILE\nLEU\nLYS\nMET\nPHE\nPRO\nSER\nTHR\nTRP\nTYR\nVAL\n#\nloop_\n_software.pdbx_ordinal\n_software.name\n_software.version\n_software.type\n_software.contact_author\n_software.contact_author_email\n_software.location\n_software.classification\n_software.citation_id\n_software.language\n1 phenix.real_space_refine 1.20rc4_4425 program 'Pavel Afonine' pafonine@lbl.gov https://www.phenix-online.org/ refinement phenix.real_space_refine Python/C++\n1 Phenix 1.20rc4_4425 program 'Paul D. Adams' pdadams@lbl.gov https://www.phenix-online.org/ refinement phenix Python/C++\n#\nloop_\n_space_group_symop.id\n_space_group_symop.operation_xyz\n1 x,y,z\n#\ncat1.name1      1 \ncat1.name2      3.14157 \ncat1.name22     2.3 \n#\nloop_\n_atom_site.group_PDB\n_atom_site.id\n_atom_site.label_atom_id\n_atom_site.label_alt_id\n_atom_site.label_comp_id\n_atom_site.auth_asym_id\n_atom_site.auth_seq_id\n_atom_site.pdbx_PDB_ins_code\n_atom_site.Cartn_x\n_atom_site.Cartn_y\n_atom_site.Cartn_z\n_atom_site.occupancy\n_atom_site.B_iso_or_equiv\n_atom_site.type_symbol\n_atom_site.pdbx_formal_charge\n_atom_site.label_asym_id\n_atom_site.label_entity_id\n_atom_site.label_seq_id\n_atom_site.pdbx_PDB_model_num\nATOM 1 N . SER B 535 ? 270.43781 345.22081 281.42585 1.000 465.59921 N ? A ? 1 1\nATOM 2 CA . SER B 535 ? 270.07764 346.63283 281.41226 1.000 465.59921 C ? A ? 1 1\nATOM 3 C . SER B 535 ? 268.70572 346.84369 282.04532 1.000 465.59921 C ? A ? 1 1\nATOM 4 O . SER B 535 ? 268.12252 345.92083 282.61134 1.000 465.59921 O ? A ? 1 1\nATOM 5 CB . SER B 535 ? 270.08494 347.17423 279.98114 1.000 465.59921 C ? A ? 1 1\nATOM 6 OG . SER B 535 ? 271.31962 346.90634 279.34048 1.000 465.59921 O ? A ? 1 1\nATOM 7 N . VAL B 536 ? 268.19595 348.07334 281.94291 1.000 465.59921 N ? A ? 2 1\nATOM 8 CA . VAL B 536 ? 266.88271 348.37783 282.50166 1.000 465.59921 C ? A ? 2 1\nATOM 9 C . VAL B 536 ? 265.76647 347.61528 281.79180 1.000 465.59921 C ? A ? 2 1\nATOM 10 O . VAL B 536 ? 265.89177 347.21437 280.62991 1.000 465.59921 O ? A ? 2 1\nATOM 11 CB . VAL B 536 ? 266.60814 349.89588 282.55950 1.000 465.59921 C ? A ? 2 1\nATOM 12 CG1 . VAL B 536 ? 266.47176 350.50312 281.15572 1.000 465.59921 C ? A ? 2 1\nATOM 13 CG2 . VAL B 536 ? 265.37232 350.19041 283.40137 1.000 465.59921 C ? A ? 2 1\nATOM 14 N . VAL B 537 ? 264.66075 347.41139 282.50950 1.000 465.59921 N ? A ? 3 1\nATOM 15 CA . VAL B 537 ? 263.53325 346.68375 281.94275 1.000 465.59921 C ? A ? 3 1\nATOM 16 C . VAL B 537 ? 263.01065 347.41204 280.70782 1.000 465.59921 C ? A ? 3 1\nATOM 17 O . VAL B 537 ? 263.07692 348.64465 280.60360 1.000 465.59921 O ? A ? 3 1\nATOM 18 CB . VAL B 537 ? 262.41900 346.51315 282.98883 1.000 465.59921 C ? A ? 3 1\nATOM 19 CG1 . VAL B 537 ? 261.74793 347.85051 283.27814 1.000 465.59921 C ? A ? 3 1\nATOM 20 CG2 . VAL B 537 ? 261.39935 345.49031 282.51421 1.000 465.59921 C ? A ? 3 1\n#\nloop_\n_atom_site_anisotrop.id\n_atom_site_anisotrop.pdbx_auth_atom_id\n_atom_site_anisotrop.pdbx_label_alt_id\n_atom_site_anisotrop.pdbx_auth_comp_id\n_atom_site_anisotrop.pdbx_auth_asym_id\n_atom_site_anisotrop.pdbx_auth_seq_id\n_atom_site_anisotrop.pdbx_PDB_ins_code\n_atom_site_anisotrop.U[1][1]\n_atom_site_anisotrop.U[2][2]\n_atom_site_anisotrop.U[3][3]\n_atom_site_anisotrop.U[1][2]\n_atom_site_anisotrop.U[1][3]\n_atom_site_anisotrop.U[2][3]\n1 N . SER B 535 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n2 CA . SER B 535 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n3 C . SER B 535 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n4 O . SER B 535 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n5 CB . SER B 535 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n6 OG . SER B 535 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n7 N . VAL B 536 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n8 CA . VAL B 536 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n9 C . VAL B 536 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n10 O . VAL B 536 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n11 CB . VAL B 536 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n12 CG1 . VAL B 536 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n13 CG2 . VAL B 536 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n14 N . VAL B 537 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n15 CA . VAL B 537 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n16 C . VAL B 537 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n17 O . VAL B 537 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n18 CB . VAL B 537 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n19 CG1 . VAL B 537 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n20 CG2 . VAL B 537 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n#\n",
			"",
		},
		{
			"valid input mmCIF, value present in both mmCIF and metadata",
			map[string]string{"foo.boo": "cat1.name1", "foo.goo": "cat1.name2", "foo.foo": "cat1.name22", "foo.doo": "cat1.name3", "mydata1": "_space_group_symop.id", "mydata2": "_space_group_symop.operation_xyz"},
			map[string][]converterUtils.PDBxItem{
				"cat1": {
					{CategoryID: "cat1", Name: "name1", Unit: "seconds", ValueType: "float", RangeMin: "0", RangeMax: "3.5", EnumValues: []string{}, PDBxEnumValues: []string{}},
					{CategoryID: "cat1", Name: "name2", Unit: "u3", ValueType: "float", RangeMin: "0", RangeMax: "3.5", EnumValues: []string{}, PDBxEnumValues: []string{}},
					{CategoryID: "cat1", Name: "name22", ValueType: "float", RangeMin: "0", RangeMax: "3.5", EnumValues: []string{}, PDBxEnumValues: []string{}},
					{CategoryID: "cat1", Name: "name3", EnumValues: []string{"hello", "world"}},
				},
				"_space_group_symop": {
					{CategoryID: "_space_group_symop", Name: "id", ValueType: "float", RangeMin: "0", RangeMax: "3.5"},
					{CategoryID: "_space_group_symop", Name: "operation_xyz"},
				},
			},
			map[string][]string{"foo.boo": {"1"}, "foo.goo": {"3.14157"}, "foo.foo": {"2.3"}, "mydata1": {"1", "2", "3"}, "mydata2": {"x,y,z", "x,y,z", "x,y,z"}},
			map[string][]string{},
			"testData/example.cif",
			"data_K3DAK4\n#\nloop_\n_citation.id\n_citation.title\n_citation.journal_abbrev\n_citation.journal_volume\n_citation.page_first\n_citation.page_last\n_citation.year\n_citation.journal_id_ASTM\n_citation.journal_id_ISSN\n_citation.journal_id_CSD\nphenix.real_space_refine 'Real-space refinement in PHENIX for cryo-EM and crystallography' 'Acta Crystallogr., Sect. D: Biol. Crystallogr.' 74 531 544 2018 ABCRE6 0907-4449 0766\n#\nloop_\n_chem_comp.id\nALA\nARG\nASN\nASP\nCYS\nGLN\nGLU\nGLY\nHIS\nILE\nLEU\nLYS\nMET\nPHE\nPRO\nSER\nTHR\nTRP\nTYR\nVAL\n#\nloop_\n_software.pdbx_ordinal\n_software.name\n_software.version\n_software.type\n_software.contact_author\n_software.contact_author_email\n_software.location\n_software.classification\n_software.citation_id\n_software.language\n1 phenix.real_space_refine 1.20rc4_4425 program 'Pavel Afonine' pafonine@lbl.gov https://www.phenix-online.org/ refinement phenix.real_space_refine Python/C++\n1 Phenix 1.20rc4_4425 program 'Paul D. Adams' pdadams@lbl.gov https://www.phenix-online.org/ refinement phenix Python/C++\n#\nloop_\n_space_group_symop.id\n_space_group_symop.operation_xyz\n1 x,y,z \n2 x,y,z \n3 x,y,z \n#\ncat1.name1      1 \ncat1.name2      3.14157 \ncat1.name22     2.3 \n#\nloop_\n_atom_site.group_PDB\n_atom_site.id\n_atom_site.label_atom_id\n_atom_site.label_alt_id\n_atom_site.label_comp_id\n_atom_site.auth_asym_id\n_atom_site.auth_seq_id\n_atom_site.pdbx_PDB_ins_code\n_atom_site.Cartn_x\n_atom_site.Cartn_y\n_atom_site.Cartn_z\n_atom_site.occupancy\n_atom_site.B_iso_or_equiv\n_atom_site.type_symbol\n_atom_site.pdbx_formal_charge\n_atom_site.label_asym_id\n_atom_site.label_entity_id\n_atom_site.label_seq_id\n_atom_site.pdbx_PDB_model_num\nATOM 1 N . SER B 535 ? 270.43781 345.22081 281.42585 1.000 465.59921 N ? A ? 1 1\nATOM 2 CA . SER B 535 ? 270.07764 346.63283 281.41226 1.000 465.59921 C ? A ? 1 1\nATOM 3 C . SER B 535 ? 268.70572 346.84369 282.04532 1.000 465.59921 C ? A ? 1 1\nATOM 4 O . SER B 535 ? 268.12252 345.92083 282.61134 1.000 465.59921 O ? A ? 1 1\nATOM 5 CB . SER B 535 ? 270.08494 347.17423 279.98114 1.000 465.59921 C ? A ? 1 1\nATOM 6 OG . SER B 535 ? 271.31962 346.90634 279.34048 1.000 465.59921 O ? A ? 1 1\nATOM 7 N . VAL B 536 ? 268.19595 348.07334 281.94291 1.000 465.59921 N ? A ? 2 1\nATOM 8 CA . VAL B 536 ? 266.88271 348.37783 282.50166 1.000 465.59921 C ? A ? 2 1\nATOM 9 C . VAL B 536 ? 265.76647 347.61528 281.79180 1.000 465.59921 C ? A ? 2 1\nATOM 10 O . VAL B 536 ? 265.89177 347.21437 280.62991 1.000 465.59921 O ? A ? 2 1\nATOM 11 CB . VAL B 536 ? 266.60814 349.89588 282.55950 1.000 465.59921 C ? A ? 2 1\nATOM 12 CG1 . VAL B 536 ? 266.47176 350.50312 281.15572 1.000 465.59921 C ? A ? 2 1\nATOM 13 CG2 . VAL B 536 ? 265.37232 350.19041 283.40137 1.000 465.59921 C ? A ? 2 1\nATOM 14 N . VAL B 537 ? 264.66075 347.41139 282.50950 1.000 465.59921 N ? A ? 3 1\nATOM 15 CA . VAL B 537 ? 263.53325 346.68375 281.94275 1.000 465.59921 C ? A ? 3 1\nATOM 16 C . VAL B 537 ? 263.01065 347.41204 280.70782 1.000 465.59921 C ? A ? 3 1\nATOM 17 O . VAL B 537 ? 263.07692 348.64465 280.60360 1.000 465.59921 O ? A ? 3 1\nATOM 18 CB . VAL B 537 ? 262.41900 346.51315 282.98883 1.000 465.59921 C ? A ? 3 1\nATOM 19 CG1 . VAL B 537 ? 261.74793 347.85051 283.27814 1.000 465.59921 C ? A ? 3 1\nATOM 20 CG2 . VAL B 537 ? 261.39935 345.49031 282.51421 1.000 465.59921 C ? A ? 3 1\n#\nloop_\n_atom_site_anisotrop.id\n_atom_site_anisotrop.pdbx_auth_atom_id\n_atom_site_anisotrop.pdbx_label_alt_id\n_atom_site_anisotrop.pdbx_auth_comp_id\n_atom_site_anisotrop.pdbx_auth_asym_id\n_atom_site_anisotrop.pdbx_auth_seq_id\n_atom_site_anisotrop.pdbx_PDB_ins_code\n_atom_site_anisotrop.U[1][1]\n_atom_site_anisotrop.U[2][2]\n_atom_site_anisotrop.U[3][3]\n_atom_site_anisotrop.U[1][2]\n_atom_site_anisotrop.U[1][3]\n_atom_site_anisotrop.U[2][3]\n1 N . SER B 535 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n2 CA . SER B 535 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n3 C . SER B 535 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n4 O . SER B 535 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n5 CB . SER B 535 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n6 OG . SER B 535 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n7 N . VAL B 536 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n8 CA . VAL B 536 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n9 C . VAL B 536 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n10 O . VAL B 536 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n11 CB . VAL B 536 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n12 CG1 . VAL B 536 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n13 CG2 . VAL B 536 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n14 N . VAL B 537 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n15 CA . VAL B 537 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n16 C . VAL B 537 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n17 O . VAL B 537 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n18 CB . VAL B 537 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n19 CG1 . VAL B 537 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n20 CG2 . VAL B 537 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n#\n",
			"",
		},
	}

	for _, test := range testCases {
		testname := fmt.Sprintf("%v", test.name)
		t.Run(testname, func(t *testing.T) {
			dictFile, _ := os.Open(test.pathExisting)
			defer dictFile.Close()
			gotValue, gotError := SupplementCoordinatesFromFile(test.namesMap, test.PDBxItems, test.jsonValues, test.unitsOSCEM, dictFile)

			if gotError != nil {
				if gotError.Error() != test.expectedError {
					t.Errorf("got error %v, wanted %v", gotError.Error(), test.expectedError)
				}
			}
			if gotValue != test.expectedText {
				t.Errorf("got:\n%v, want:\n%v", gotValue, test.expectedText)

			}
		})
	}
}
