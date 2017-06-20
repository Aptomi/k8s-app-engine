package slinga

import (
	"reflect"
	log "github.com/Sirupsen/logrus"
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
			result += countElements(v.Index(i).Interface())
		}
	case reflect.Map:
		for _, key := range v.MapKeys() {
			result += countElements(v.MapIndex(key).Interface())
		}
	default:
		result++
	}

	return result
}

// GetSortedStringKeys assumes m is a map[string]interface{} and returns an array of sorted keys
func GetSortedStringKeys(m interface{}) []string {
	result := []string{}
	if m == nil {
		return result
	}

	v := reflect.ValueOf(m)
	if v.Kind() != reflect.Map {
		debug.WithFields(log.Fields{
			"data": m,
		}).Fatal("Not a map")
	}
	for _, key := range v.MapKeys() {
		k, ok := key.Interface().(string)
		if !ok {
			debug.WithFields(log.Fields{
				"data": m,
				"key": key,
			}).Fatal("Expected a string key")
		}
		result = append(result, k)
	}

	return result
}