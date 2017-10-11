package util

// CopySliceOfStrings makes a copy of a given slice of strings
func CopySliceOfStrings(a []string) []string {
	result := make([]string, len(a))
	copy(result, a)
	return result
}

// ContainsString checks if a slice of strings contains a given string
func ContainsString(a []string, v string) bool {
	for _, value := range a {
		if value == v {
			return true
		}
	}
	return false
}
