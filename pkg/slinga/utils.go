package slinga

import "strings"

func EscapeName(str string) string {
	r := strings.NewReplacer("#", "-", "_", "-")
	return r.Replace(str)
}
