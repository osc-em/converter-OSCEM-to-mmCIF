package parser

import (
	"converter/converterUtils"
	"fmt"
	"testing"
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

	for _, tt := range tests {

		testname := fmt.Sprintf("%v", tt.name)
		t.Run(testname, func(t *testing.T) {
			gotValue, gotError := getKeyByValue(tt.value, tt.dictionary)

			if gotError != nil {
				if gotError.Error() != tt.expectedError {
					t.Errorf("got error %v, wanted %v", gotError.Error(), tt.expectedError)
				}
			}
			if gotValue != tt.expectedResult {
				t.Errorf("got %v, want %v", gotValue, tt.expectedResult)
			}
		})
	}
}

func TestSliceContains(t *testing.T) {
	var tests = []struct {
		name           string
		slice          []string
		element        string
		expectedResult bool
	}{
		{"element in slice", []string{"hello", "world"}, "hello", true},
		{"element not in slice", []string{"hello", "world"}, "foo", false},
	}

	for _, tt := range tests {

		testname := fmt.Sprintf("%v", tt.name)
		t.Run(testname, func(t *testing.T) {
			gotValue := sliceContains(tt.slice, tt.element)

			if gotValue != tt.expectedResult {
				t.Errorf("got %v, want %v", gotValue, tt.expectedResult)
			}
		})
	}
}

func TestValidateDateIsRFC3339(t *testing.T) {
	var tests = []struct {
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

	for _, tt := range tests {

		testname := fmt.Sprintf("%v", tt.name)
		t.Run(testname, func(t *testing.T) {
			gotValue := validateDateIsRFC3339(tt.date)

			if gotValue != tt.expectedResult {
				t.Errorf("got %v, want %v", gotValue, tt.expectedResult)
			}
		})
	}
}

func TestValidateRange(t *testing.T) {
	var tests = []struct {
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

	for _, tt := range tests {

		testname := fmt.Sprintf("%v", tt.name)
		t.Run(testname, func(t *testing.T) {
			gotValue, gotError := validateRange(tt.value, tt.dataItem, tt.unitOSCEM, tt.nameOSCEM)

			if gotError != nil {
				if gotError.Error() != tt.expectedError {
					t.Errorf("got error %v, wanted %v", gotError.Error(), tt.expectedError)
				}
			}
			if gotValue != tt.expectedResult {
				t.Errorf("got %v, want %v", gotValue, tt.expectedResult)
			}
		})
	}
}

func TestValidateEnum(t *testing.T) {
	var tests = []struct {
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

	for _, tt := range tests {

		testname := fmt.Sprintf("%v", tt.name)
		t.Run(testname, func(t *testing.T) {
			gotValue := validateEnum(tt.value, tt.dataItem)

			if gotValue != tt.expectedResult {
				t.Errorf("got %v, want %v", gotValue, tt.expectedResult)
			}
		})
	}
}

// func TestExtractRangeValue(t *testing.T) {
// 	var tests = []struct {
// 		line          string
// 		expectedValue string
// 		expectedError string
// 	}{
// 		{"_item_range.maximum", "?", ""},
// 		{"_item_range.minimum  0.0", "0.0", ""},
// 		{"_item_range.minimum  .", ".", ""},
// 		{"_item_range.minimum  ?", "?", ""},
// 		{"_item_range.minimum  *", "?", "value * is not numeric"},
// 	}

// 	for _, tt := range tests {

// 		testname := fmt.Sprintf("%v", tt.line)
// 		t.Run(testname, func(t *testing.T) {
// 			gotValue, gotError := extractRangeValue(tt.line)

// 			if gotError != nil {
// 				if gotError.Error() != tt.expectedError {
// 					t.Errorf("got error %v, wanted %v", gotError.Error(), tt.expectedError)
// 				}
// 			}
// 			if gotValue != tt.expectedValue {
// 				t.Errorf("got %v, want %v", gotValue, tt.expectedValue)
// 			}
// 		})
// 	}
// }
