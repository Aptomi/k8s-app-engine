package slinga

import (
	"reflect"
)

func countElements(m interface{}) int {
	result := 0
	if m == nil {
		return result
	}

	v := reflect.ValueOf(m)
	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			result += countElements(v.Index(i))
		}
	case reflect.Map:
		for _, key := range v.MapKeys() {
			result += countElements(v.MapIndex(key))
		}
	default:
		result++
	}

	return result
}
