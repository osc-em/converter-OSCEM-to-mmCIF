package converterUtils

func GetKeys[K string, V any](m map[string]V) []string {
	keys := make([]string, 0)
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func GetValues[K string, V string](m map[string]string) []string {
	values := make([]string, 0)
	for _, v := range m {
		values = append(values, v)
	}
	return values
}

func GetKeyByValue(value string, m map[string]string) string {
	for k, v := range m {
		if v == value {
			return k
		}
	}
	return ""
}

func StringJoiner(stringsArray []string) string {
	var joinedString string
	for i := range stringsArray {
		if stringsArray[i] != "" {
			if joinedString != "" {
				joinedString += "."
			}
			joinedString += stringsArray[i]
		}
	}
	return joinedString
}

func Contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func GetLongest(s []string) int {
	var r int
	for _, a := range s {
		if len(a) > r {
			r = len(a)
		}
	}
	return r
}
