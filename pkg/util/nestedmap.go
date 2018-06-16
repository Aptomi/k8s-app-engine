package util

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/Aptomi/aptomi/pkg/lang/template"
	"github.com/d4l3k/messagediff"
	log "github.com/sirupsen/logrus"
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
		dst[key] = srcInt
		return
	}
	if srcBool, ok := src.(bool); ok {
		dst[key] = srcBool
		return
	}

	panic("invalid type in NestedParameterMap (expected string, int, or bool)")
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

// GetString returns string located by provided key
func (src NestedParameterMap) GetString(key string, defaultValue string) (string, error) {
	value, exist := src[key]
	if !exist {
		return defaultValue, nil
	}

	str, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("value for key %s isn't string", key)
	}
	if len(str) == 0 {
		return defaultValue, nil
	}

	return str, nil
}

const (
	includeMacrosPrefix = "@include "
)

// ProcessIncludeMacros walks through specified NestedParameterMap and resolves @include macros
func ProcessIncludeMacros(node NestedParameterMap, baseDir string) error {
	for key, value := range node {
		// If it's a string, evaluate macros
		if str, strOk := value.(string); strOk && strings.HasPrefix(str, includeMacrosPrefix) {
			filePathsStr := strings.TrimSpace(str[len(includeMacrosPrefix):])
			if len(filePathsStr) == 0 {
				return fmt.Errorf("@include macros should have exactly one parameter - paths to the files to be included")
			}
			filePaths := strings.Split(filePathsStr, ",")
			if len(filePaths) == 0 {
				return fmt.Errorf("@include macros should point to at least one file")
			}

			files, err := findIncludeFiles(baseDir, filePaths)
			if err != nil {
				return err
			}

			manifest := ""
			for _, file := range files {
				data, dataErr := ioutil.ReadFile(file)
				if dataErr != nil {
					return fmt.Errorf("can't read file to include %s: %s", file, err)
				}
				manifest += "\n---\n" + string(data)
			}
			node[key] = manifest
		} else if nestedMap, mapOk := value.(NestedParameterMap); mapOk {
			err := ProcessIncludeMacros(nestedMap, baseDir)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func findIncludeFiles(baseDir string, filePaths []string) ([]string, error) {
	for idx, file := range filePaths {
		if !filepath.IsAbs(file) {
			filePaths[idx] = filepath.Join(baseDir, file)
		}
	}

	allFiles, err := FindYamlFiles(filePaths)
	if err != nil {
		return nil, err
	}

	log.Debug("Including data from files:")
	for _, file := range allFiles {
		log.Debugf("  [*] %s", file)
	}

	return allFiles, nil
}

const (
	// ModeCompile just compiles all text templates on parameter tree without evaluating
	ModeCompile = iota

	// ModeEvaluate evaluates the whole parameter tree and all of its text templates, given a set of parameters
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

			result[key] = evaluatedValue
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

	// If it's a int, put as is
	if valueInt, ok := node.(int); ok {
		result[key] = valueInt
		return nil
	}

	// If it's a bool, put as is
	if valueBool, ok := node.(bool); ok {
		result[key] = valueBool
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
	return fmt.Errorf("invalid type in NestedParameterMap (expected string, int, or bool): %v", node)
}
