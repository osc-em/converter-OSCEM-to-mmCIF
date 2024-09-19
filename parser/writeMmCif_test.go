package parser

import (
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
