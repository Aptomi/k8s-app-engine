package language

import (
	"github.com/Aptomi/aptomi/pkg/slinga/language/expression"
)

// Criteria defines a structure with criteria filter/accept/reject syntax
type Criteria struct {
	// This follows 'AND' logic. This is basically a pre-condition, and all expressions should be evaluated to true in order to proceed to the next section
	Filter []string

	// This follows 'OR' logic. If any of this evaluates to true, we will proceed to the next section
	Accept []string

	// This follows 'AND NOT' logic. If any of this evaluates to true, criteria will evaluate to false immediately
	Reject []string

	cachedExpressions map[string]*expression.Expression
}

// Whether criteria evaluates to "true" for a given set of labels or not
func (criteria *Criteria) allows(params *expression.ExpressionParameters) bool {
	// If one of the reject expressions matches, then the criteria is not allowed
	for _, rejectExpr := range criteria.Reject {
		result, err := criteria.evaluateBool(rejectExpr, params)
		if err != nil {
			// TODO: we probably want to fail the whole criteria (with false) and propagate error to the user
			panic(err)
		}
		if result {
			return false
		}
	}

	// If one of the filter expressions does not match, then the criteria is not allowed
	for _, filterExpr := range criteria.Filter {
		result, err := criteria.evaluateBool(filterExpr, params)
		if err != nil {
			// TODO: we probably want to fail the whole criteria (with false) and propagate error to the user
			panic(err)
		}
		if !result {
			return false
		}
	}

	// If one of the accept expressions matches, then the criteria is allowed
	for _, acceptExpr := range criteria.Accept {
		result, err := criteria.evaluateBool(acceptExpr, params)
		if err != nil {
			// TODO: we probably want to fail the whole criteria (with false) and propagate error to the user
			panic(err)
		}
		if result {
			return true
		}
	}

	// If the accept section is empty, return true
	if len(criteria.Accept) == 0 {
		return true
	}

	return false
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
