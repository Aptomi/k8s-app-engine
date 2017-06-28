package util

import "reflect"

/*
	This file declares all utility structures and methods required for Slinga processing
*/

// NestedParameterMap allows to work with nested maps [string][string]...[string] -> value
type NestedParameterMap map[string]interface{}

// MakeCopy makes a copy of parameter structure
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

// DeepEqual compares two nested maps. If one is nil and another one is empty, it will return true as well
func (src NestedParameterMap) DeepEqual(dst NestedParameterMap) bool {
	if len(src) == 0 && len(dst) == 0 {
		return true
	}
	return reflect.DeepEqual(src, dst)
}
