package engine

import (
	. "github.com/Aptomi/aptomi/pkg/slinga/util"
	. "github.com/Aptomi/aptomi/pkg/slinga/language"
	"github.com/Aptomi/aptomi/pkg/slinga/language/template"
	"fmt"
)

func processParameterTreeNode(node ParameterTree, parameters *template.TemplateParameters, result NestedParameterMap, key string, cache template.TemplateCache) error {
	if node == nil {
		return nil
	}

	// If it's a string, evaluate template
	if paramsStr, ok := node.(string); ok {
		evaluatedValue, err := cache.Evaluate(paramsStr, parameters)
		if err != nil {
			return nil
		}
		result[key] = EscapeName(evaluatedValue)
		return nil
	}

	// If it's an int, put it directly
	if paramsInt, ok := node.(int); ok {
		result[key] = paramsInt
		return nil
	}

	// If it's a bool, put it directly
	if paramsBool, ok := node.(bool); ok {
		result[key] = paramsBool
		return nil
	}

	// If it's a map, process it recursively
	if paramsMap, ok := node.(map[interface{}]interface{}); ok {
		for key, value := range paramsMap {
			result[key.(string)] = NestedParameterMap{}
			err := processParameterTreeNode(value, parameters, result, key.(string), cache)
			if err != nil {
				return err
			}
		}
		return nil
	}

	// Unknown type, return an error
	return fmt.Errorf("There should be map[string]interface{} or a primitive type, but found %v", node)
}

// evaluateParameterTree processes code or discovery params and calculates the whole tree
func evaluateParameterTree(tree ParameterTree, parameters *template.TemplateParameters, cache template.TemplateCache) (NestedParameterMap, error) {
	if tree == nil {
		return nil, nil
	}
	if cache == nil {
		cache = template.NewTemplateCache()
	}

	result := NestedParameterMap{}
	err := processParameterTreeNode(tree, parameters, result, "", cache)
	return result, err
}
