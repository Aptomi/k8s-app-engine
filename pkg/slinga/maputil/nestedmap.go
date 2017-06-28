package slinga

import "reflect"

/*
	This file declares all utility structures and methods required for Slinga processing
*/

// NestedParameterMap allows to work with nested maps [string][string]...[string] -> value
type NestedParameterMap map[string]interface{}

// Makes of copy of parameter structure
func (src NestedParameterMap) makeCopy() NestedParameterMap {
	result := NestedParameterMap{}
	for k, v := range src {
		result[k] = v
	}
	return result
}

// Gets nested parameter map
func (src NestedParameterMap) getNestedMap(key string) NestedParameterMap {
	return src[key].(NestedParameterMap)
}

// Function to compare two nested maps. If one is nil and another one is empty, it will return true as well
func (src NestedParameterMap) deepEqual(dst NestedParameterMap) bool {
	if len(src) == 0 && len(dst) == 0 {
		return true
	}
	return reflect.DeepEqual(src, dst)
}
