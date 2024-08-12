package parser

import (
	"testing"
)

type addTest struct {
	path           string
	col            string
	expectedValues []string
	expectedError  string
}

var addTests = []addTest{
	addTest{"./testData/notCsv.cs", "PDBx", make([]string, 0), "open ./testData/notCsv.cs: no such file or directory"},
	addTest{"./testData/notCsv.csv", "PDBx", make([]string, 0), "Column PDBx does not exist in table ./testData/notCsv.csv"},
	addTest{"./testData/notCsv.csv", "mmCIF", make([]string, 0), "Column mmCIF does not exist in table ./testData/notCsv.csv"},
	addTest{"./testData/OSCEMfirst.csv", "OSCEM", []string{"",
		"Instrument.Microscope",
		"Instrument.Illumination",
		"Instrument.Imaging",
		"Instrument.Electron_source",
		"Instrument.Acceleration.Voltage"},
		""},
	addTest{"./testData/OSCEMmiddle.csv", "OSCEM", []string{"",
		"Instrument.Microscope",
		"Instrument.Illumination",
		"Instrument.Imaging",
		"Instrument.Electron_source",
		"Instrument.Acceleration_Voltage",
		"Instrument.C2_Aperture",
		"Instrument.CS"},
		""},
}

// because the delimiter used in this file is '\t' the whole table is read as one column, so PDBx column will not be found
func TestConversionTableReadColumn(t *testing.T) {

	for _, test := range addTests {
		gotValues, gotError := ConversionTableReadColumn(test.path, test.col)
		if gotError == nil {
			if len(test.expectedError) > 0 {
				t.Errorf("Expected error: %q, but got no error", test.expectedError)
			} else if len(gotValues) != len(test.expectedValues) {
				t.Errorf("Expected output slice: %q, got: %q", test.expectedValues, gotValues)
			} else {
				for i := range gotValues {
					if gotValues[i] != test.expectedValues[i] {
						t.Errorf("Expected output slice: %q, got: %q", test.expectedValues, gotValues)
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
	}
}
