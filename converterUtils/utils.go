// Package converterUtils provides set of simple functions that are required in converter Package
package converterUtils

// GetKeys returns a slice of string arrays with values in map. A key must be a string value.
func GetKeys[K string, V any](m map[string]V) []string {

	keys := make([]string, 0)
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// PDBxItem type defines attributes of the data item property in PDBx dictionary.
// It contains most important fields, such as name of the item and its parental catrgory,
// type of value it should take on and units, range or allowed values if there are any.
type PDBxItem struct {
	CategoryID string
	Name       string
	Unit       string
	ValueType  string
	RangeMin   float64
	RangeMax   float64
	EnumValues []string
}
