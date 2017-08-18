package engine

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/slinga/language/template"
	. "github.com/Aptomi/aptomi/pkg/slinga/util"
)

func processParameterTreeNode(node interface{}, parameters *template.TemplateParameters, result NestedParameterMap, key string, cache template.TemplateCache) error {
	if node == nil {
		return nil
	}

	// If it's a string, evaluate template
	if paramsStr, ok := node.(string); ok {
		evaluatedValue, err := cache.Evaluate(paramsStr, parameters)
		if err != nil {
			return err
		}
		result[key] = EscapeName(evaluatedValue)
		return nil
	}

	// If it's a map, process it recursively
	if paramsMap, ok := node.(NestedParameterMap); ok {
		if len(key) > 0 {
			result = result.GetNestedMap(key)
		}
		for pKey, pValue := range paramsMap {
			result[pKey] = NestedParameterMap{}
			err := processParameterTreeNode(pValue, parameters, result, pKey, cache)
			if err != nil {
				return err
			}
		}
		return nil
	}

	// Unknown type, return an error
	return fmt.Errorf("There should a string or NestedParameterMap, but found %v", node)
}

// evaluateParameterTree processes code or discovery params and calculates the whole tree
func evaluateParameterTree(tree NestedParameterMap, parameters *template.TemplateParameters, cache template.TemplateCache) (NestedParameterMap, error) {
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
