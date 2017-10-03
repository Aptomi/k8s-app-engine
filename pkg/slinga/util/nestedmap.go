package util

import (
	"github.com/Aptomi/aptomi/pkg/slinga/lang/yaml"
	"github.com/d4l3k/messagediff"
	"reflect"
	"strconv"
)

/*
	This file declares all utility structures and methods required for Slinga processing
*/

// NestedParameterMap allows to work with nested maps [string][string]...[string] -> value
type NestedParameterMap map[string]interface{}

// Custom unmarshal function to deal with interface{} -> string conversions
func (src *NestedParameterMap) UnmarshalYAML(unmarshal func(interface{}) error) error {
	result := make(map[interface{}]interface{})
	if err := unmarshal(&result); err != nil {
		return err
	}
	*src = NestedParameterMap{}
	put(result, *src, "")
	return nil
}

// Takes src map of map[interface{}]interface{} and puts it into dst
func put(src interface{}, dst NestedParameterMap, key string) {
	if src == nil {
		return
	}

	// If it's a map, process it recursively
	if pMap, ok := src.(map[interface{}]interface{}); ok {
		if len(key) > 0 {
			dst = dst.GetNestedMap(key)
		}
		for pKey, pValue := range pMap {
			dst[pKey.(string)] = NestedParameterMap{}
			put(pValue, dst, pKey.(string))
		}
		return
	}

	// Otherwise, just put string value into the map
	if srcString, ok := src.(string); ok {
		dst[key] = srcString
		return
	}
	if srcInt, ok := src.(int); ok {
		dst[key] = strconv.Itoa(srcInt)
		return
	}
	if srcBool, ok := src.(bool); ok {
		dst[key] = strconv.FormatBool(srcBool)
		return
	}

	panic("Invalid type in map, can't convert to NestedParameterMap")
}

// MakeCopy makes a shallow copy of parameter structure
func (src NestedParameterMap) MakeCopy() NestedParameterMap {
	result := NestedParameterMap{}
	for k, v := range src {
		result[k] = v
	}
	return result
}

// GetNestedMap returns nested parameter map by key
func (src NestedParameterMap) GetNestedMap(key string) NestedParameterMap {
	return src[key].(NestedParameterMap)
}

// DeepEqual compares two nested parameter maps
// If both maps are empty (have zero elements), the method will return true
func (src NestedParameterMap) DeepEqual(dst NestedParameterMap) bool {
	if len(src) == 0 && len(dst) == 0 {
		return true
	}
	return reflect.DeepEqual(src, dst)
}

// Diff feturns a human-readable diff between two nested parameter maps
func (src NestedParameterMap) Diff(dst NestedParameterMap) string {
	// 2nd parameter is true/false, whether the structs are equal or not. We can safely ignore it
	diff, _ := messagediff.PrettyDiff(src, dst)
	return diff
}

// ToString returns a string representation of a nested parameter map
func (src NestedParameterMap) ToString() string {
	return yaml.SerializeObject(src)
}
