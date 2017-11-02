package util

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/lang/template"
	"github.com/Aptomi/aptomi/pkg/lang/yaml"
	"github.com/d4l3k/messagediff"
	"reflect"
	"strconv"
)

// NestedParameterMap is a nested map of parameters, which allows to work with maps [string][string]...[string] -> string, int, bool values
type NestedParameterMap map[string]interface{}

// UnmarshalYAML is a custom unmarshal function for NestedParameterMap to deal with interface{} -> string conversions
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

// Diff returns a human-readable diff between two nested parameter maps
func (src NestedParameterMap) Diff(dst NestedParameterMap) string {
	// second parameter is a result true/false, indicating whether they are equal or not. we can safely ignore it
	diff, _ := messagediff.PrettyDiff(src, dst)
	return diff
}

// ToString returns a string representation of a nested parameter map
func (src NestedParameterMap) ToString() string {
	return yaml.SerializeObject(src)
}

const (
	ModeCompile  = iota
	ModeEvaluate = iota
)

// ProcessParameterTree processes NestedParameterMap and calculates the whole tree, assuming values are text templates
func ProcessParameterTree(tree NestedParameterMap, parameters *template.Parameters, cache *template.Cache, mode int) (NestedParameterMap, error) {
	if tree == nil {
		return nil, nil
	}
	if cache == nil {
		cache = template.NewCache()
	}

	result := NestedParameterMap{}
	err := processParameterTreeNode(tree, parameters, result, "", cache, mode)
	return result, err
}

func processParameterTreeNode(node interface{}, parameters *template.Parameters, result NestedParameterMap, key string, cache *template.Cache, mode int) error {
	if node == nil {
		return nil
	}

	// If it's a string, evaluate template
	if templateStr, ok := node.(string); ok {
		if mode == ModeEvaluate {
			// evaluate and store
			evaluatedValue, err := cache.Evaluate(templateStr, parameters)
			if err != nil {
				return err
			}
			result[key] = EscapeName(evaluatedValue)
		} else if mode == ModeCompile {
			// just compile
			_, err := template.NewTemplate(templateStr)
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("unknown mode: %d", mode)
		}
		return nil
	}

	// If it's a map, process it recursively
	if paramsMap, ok := node.(NestedParameterMap); ok {
		if len(key) > 0 {
			result = result.GetNestedMap(key)
		}
		for pKey, pValue := range paramsMap {
			result[pKey] = NestedParameterMap{}
			err := processParameterTreeNode(pValue, parameters, result, pKey, cache, mode)
			if err != nil {
				return err
			}
		}
		return nil
	}

	// Unknown type, return an error
	return fmt.Errorf("expected string or NestedParameterMap, but found %v", node)
}
