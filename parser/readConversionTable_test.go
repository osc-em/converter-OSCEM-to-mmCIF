package parser

import (
	"reflect"
	"testing"
)

type convTest struct {
	path           string
	col            string
	expectedValues []string
	expectedError  string
}

// Tests:
// * non-existing file
// * broken csv - in seconf line remove bunch of fields defined in header
// * extracting a column that does not exist
// * trying to extract a column that exists but the file is not a true csv, in this case uses tabs, so the whole header is parsed as one column
// * extract OSCEM column, that is the first column in the header
// * extract OSCEM column, that is the middle of the header

var convTableTest = []convTest{
	{"./testData/notCsv.cs", "PDBx", make([]string, 0), "open ./testData/notCsv.cs: no such file or directory"},
	{"./testData/badCsv.csv", "units", make([]string, 0), "record on line 2: wrong number of fields"},
	{"./testData/notCsv.csv", "PDBx", make([]string, 0), "column PDBx does not exist in table ./testData/notCsv.csv"},
	{"./testData/notCsv.csv", "units", make([]string, 0), "column units does not exist in table ./testData/notCsv.csv"},
	{"./testData/OSCEMfirst.csv", "OSCEM", []string{"",
		"Instrument.Microscope",
		"Instrument.Illumination",
		"Instrument.Imaging",
		"Instrument.Electron_source",
		"Instrument.Acceleration.Voltage"},
		""},
	{"./testData/OSCEMmiddle.csv", "OSCEM", []string{"",
		"Instrument.Microscope",
		"Instrument.Illumination",
		"Instrument.Imaging",
		"Instrument.Electron_source",
		"Instrument.Acceleration_Voltage",
		"Instrument.C2_Aperture",
		"Instrument.CS"},
		""},
}

func TestConversionTableReadColumn(t *testing.T) {
	for _, test := range convTableTest {
		gotValues, gotError := ConversionTableReadColumn(test.path, test.col)
		if gotError == nil {
			if len(test.expectedError) > 0 {
				t.Errorf("Expected error: %q, but got no error", test.expectedError)
			} else if len(gotValues) != len(test.expectedValues) {
				t.Errorf("Expected output slice: %q, got: %q", test.expectedValues, gotValues)
			} else {
				if !reflect.DeepEqual(gotValues, test.expectedValues) {
					t.Errorf("Expected output slice: %v, got: %v", test.expectedValues, gotValues)
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
