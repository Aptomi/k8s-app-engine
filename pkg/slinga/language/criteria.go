package language

import (
	"github.com/Aptomi/aptomi/pkg/slinga/language/expression"
)

// Criteria defines a structure with require-all, require-any and require-none syntax
type Criteria struct {
	// This follows 'AND' logic. This is basically a pre-condition, and all of its expressions are required to evaluate to true
	RequireAll []string `yaml:"require-all"`

	// This follows 'OR' logic. At least one of its expressions is required to evaluate to true
	RequireAny []string `yaml:"require-any"`

	// This follows 'AND NOT' logic. None of its expressions should evaluate to true
	RequireNone []string `yaml:"require-none"`
}

// Whether criteria evaluates to "true" for a given set of labels or not
func (criteria *Criteria) allows(params *expression.ExpressionParameters, cache expression.ExpressionCache) (bool, error) {
	// Make sure all "require-all" criterias evaluate to true
	for _, exprShouldBeTrue := range criteria.RequireAll {
		result, err := criteria.evaluateBool(exprShouldBeTrue, params, cache)
		if err != nil {
			// propagate expression error up, if happened
			return false, err
		}
		if !result {
			return false, nil
		}
	}

	// Make sure that none of "require-none" criterias evaluate to true
	for _, exprShouldBeFalse := range criteria.RequireNone {
		result, err := criteria.evaluateBool(exprShouldBeFalse, params, cache)
		if err != nil {
			// propagate expression error up, if happened
			return false, err
		}
		if result {
			return false, nil
		}
	}

	// Make sure at least one "require-any" criterias evaluates to true
	if len(criteria.RequireAny) > 0 {
		for _, exprShouldBeTrue := range criteria.RequireAny {
			result, err := criteria.evaluateBool(exprShouldBeTrue, params, cache)
			if err != nil {
				// propagate expression error up, if happened
				return false, err
			}
			if result {
				return true, nil
			}
		}

		// If no criteria got evaluated to true, return false
		return false, nil
	}

	// Everything is fine and "require-any" is empty, let's return true
	return true, nil
}

func (criteria *Criteria) evaluateBool(expressionStr string, params *expression.ExpressionParameters, cache expression.ExpressionCache) (bool, error) {
	if cache == nil {
		cache = expression.NewExpressionCache()
	}
	return cache.EvaluateAsBool(expressionStr, params)
}
