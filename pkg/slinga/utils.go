package slinga

import "strings"

// EscapeName escapes provided string by replacing # and _ with -
func EscapeName(str string) string {
	r := strings.NewReplacer("#", "-", "_", "-")
	return r.Replace(str)
}
