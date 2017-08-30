package util

func CopySliceOfStrings(a []string) []string {
	result := make([]string, len(a))
	copy(result, a)
	return result
}

func ContainsString(a []string, v string) bool {
	for _, value := range a {
		if value == v {
			return true
		}
	}
	return false
}
