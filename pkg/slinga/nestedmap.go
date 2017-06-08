package slinga

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
