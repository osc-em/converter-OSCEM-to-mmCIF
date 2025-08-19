package parser

import (
	"fmt"
	"reflect"
	"testing"
)

type convTest struct {
	name           string
	path           string
	col            string
	expectedValues []string
	expectedError  string
}

var convTableTest = []convTest{
	{"file does not exist", "./testData/notCsv.cs", "PDBx", make([]string, 0), "open ./testData/notCsv.cs: no such file or directory"},
	{"csv broken: in 2nd line not as many fields as defined in header", "./testData/badCsv.csv", "units", make([]string, 0), "record on line 2: wrong number of fields"},
	{"tries to extract column that does not exist", "./testData/notCsv.csv", "mmCIF", make([]string, 0), "column mmCIF does not exist in table ./testData/notCsv.csv"},
	{"tries to extract a column, but file uses other delim as comma, table parsed as single column", "./testData/notCsv.csv", "units", make([]string, 0), "column units does not exist in table ./testData/notCsv.csv"},
	{"OSCEM is first column", "./testData/OSCEMfirst.csv", "OSCEM", []string{
		"Instrument.Microscope",
		"Instrument.Illumination",
		"Instrument.Imaging",
		"Instrument.Electron_source",
		"Instrument.AccelerationVoltage"},
		""},
	{"OSCEM in the middle of header", "./testData/OSCEMmiddle.csv", "OSCEM", []string{
		"Instrument.Microscope",
		"Instrument.Illumination",
		"Instrument.Imaging",
		"Instrument.Electron_source",
		"Instrument.AccelerationVoltage",
		"Instrument.C2_Aperture",
		"Instrument.CS"},
		""},
}

func TestConversionTableReadColumn(t *testing.T) {
	for _, test := range convTableTest {
		testname := fmt.Sprintf("%v", test.name)
		t.Run(testname, func(t *testing.T) {
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
		})
	}
}
