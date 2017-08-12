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

	cachedExpressions map[string]*expression.Expression
}

// Whether criteria evaluates to "true" for a given set of labels or not
func (criteria *Criteria) allows(params *expression.ExpressionParameters) bool {
	// Make sure all "require-all" criterias evaluate to true
	for _, exprShouldBeTrue := range criteria.RequireAll {
		result, err := criteria.evaluateBool(exprShouldBeTrue, params)
		if err != nil {
			// TODO: we probably want to fail the whole criteria (with false) and propagate error to the user
			panic(err)
		}
		if !result {
			return false
		}
	}

	// Make sure that none of "require-none" criterias evaluate to true
	for _, exprShouldBeFalse := range criteria.RequireNone {
		result, err := criteria.evaluateBool(exprShouldBeFalse, params)
		if err != nil {
			// TODO: we probably want to fail the whole criteria (with false) and propagate error to the user
			panic(err)
		}
		if result {
			return false
		}
	}

	// Make sure at least one "require-any" criterias evaluates to true
	if len(criteria.RequireAny) > 0 {
		for _, exprShouldBeTrue := range criteria.RequireAny {
			result, err := criteria.evaluateBool(exprShouldBeTrue, params)
			if err != nil {
				// TODO: we probably want to fail the whole criteria (with false) and propagate error to the user
				panic(err)
			}
			if result {
				return true
			}
		}

		// If no criteria got evaluated to true, return false
		return false
	}

	// Everything is fine and "require-any" is empty, let's return true
	return true
}

func (criteria *Criteria) evaluateBool(expressionStr string, params *expression.ExpressionParameters) (bool, error) {
	// Initialize cache if it's empty
	if criteria.cachedExpressions == nil {
		criteria.cachedExpressions = make(map[string]*expression.Expression)
	}

	// Look up expression from cache or compile
	var expr *expression.Expression
	var ok bool
	expr, ok = criteria.cachedExpressions[expressionStr]
	if !ok {
		var err error
		expr, err = expression.NewExpression(expressionStr)
		if err != nil {
			return false, err
		}
		criteria.cachedExpressions[expressionStr] = expr
	}

	// Evaluate
	return expr.EvaluateAsBool(params)
}
