package resolve

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/slinga/lang/template"
	"github.com/Aptomi/aptomi/pkg/slinga/util"
)

func processParameterTreeNode(node interface{}, parameters *template.Parameters, result util.NestedParameterMap, key string, cache *template.Cache) error {
	if node == nil {
		return nil
	}

	// If it's a string, evaluate template
	if paramsStr, ok := node.(string); ok {
		evaluatedValue, err := cache.Evaluate(paramsStr, parameters)
		if err != nil {
			return err
		}
		result[key] = util.EscapeName(evaluatedValue)
		return nil
	}

	// If it's a map, process it recursively
	if paramsMap, ok := node.(util.NestedParameterMap); ok {
		if len(key) > 0 {
			result = result.GetNestedMap(key)
		}
		for pKey, pValue := range paramsMap {
			result[pKey] = util.NestedParameterMap{}
			err := processParameterTreeNode(pValue, parameters, result, pKey, cache)
			if err != nil {
				return err
			}
		}
		return nil
	}

	// Unknown type, return an error
	return fmt.Errorf("There should be a string or NestedParameterMap, but found %v", node)
}

// evaluateParameterTree processes code or discovery params and calculates the whole tree
func evaluateParameterTree(tree util.NestedParameterMap, parameters *template.Parameters, cache *template.Cache) (util.NestedParameterMap, error) {
	if tree == nil {
		return nil, nil
	}
	if cache == nil {
		cache = template.NewCache()
	}

	result := util.NestedParameterMap{}
	err := processParameterTreeNode(tree, parameters, result, "", cache)
	return result, err
}
