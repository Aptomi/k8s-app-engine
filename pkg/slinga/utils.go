package slinga

import (
	"strings"
)

// EscapeName escapes provided string by replacing # and _ with -
func EscapeName(str string) string {
	r := strings.NewReplacer("#", "-", "_", "-")
	return r.Replace(str)
}

func stringContainsAny(str string, substrs ...string) bool {
	for _, substr := range substrs {
		if strings.Contains(str, substr) {
			return true
		}
	}

	return false
}
