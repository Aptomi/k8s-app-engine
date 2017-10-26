package lang

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/errors"
	"github.com/Aptomi/aptomi/pkg/lang/expression"
)

// Criteria is a structure which allows users to define complex matching expressions in the policy. Criteria
// expressions can refer to labels through variables. It supports require-all, require-any and require-none clauses,
// with a list of expressions under each clause.
//
// Criteria gets evaluated to true only when
// (1) All RequireAll expression evaluate to true,
// (2) At least one of RequireAny expressions evaluates to true,
// (3) None of RequireNone expressions evaluate to true.
//
// If any of RequireAll, RequireAny, RequireNone are absent, the corresponding clause will be skipped. So it's
// perfectly fine to have a criteria with fewer than 3 clauses (e.g. just RequireAll), or with no sections at all. Empty
// criteria without any clauses always evaluates to true
type Criteria struct {
	// RequireAll follows 'AND' logic
	RequireAll []string `yaml:"require-all" validate:"expression"`

	// RequireAny follows 'OR' logic
	RequireAny []string `yaml:"require-any" validate:"expression"`

	// RequireNone follows 'AND NOT'
	RequireNone []string `yaml:"require-none" validate:"expression"`
}

// Returns whether criteria evaluates to "true", given a set of parameters for its expressions and a cache
func (criteria *Criteria) allows(params *expression.Parameters, cache *expression.Cache) (bool, error) {
	// Make sure all "require-all" criteria evaluate to true
	for _, exprShouldBeTrue := range criteria.RequireAll {
		result, err := criteria.evaluateBool(exprShouldBeTrue, params, cache)
		if err != nil {
			// propagate expression error up, if happened
			return false, errors.NewErrorWithDetails(
				fmt.Sprintf("Can't evaluate 'require-all' in criteria: %s", err),
				errors.Details{
					"criteria":   criteria,
					"expression": exprShouldBeTrue,
				},
			)
		}
		if !result {
			return false, nil
		}
	}

	// Make sure that none of "require-none" criteria evaluate to true
	for _, exprShouldBeFalse := range criteria.RequireNone {
		result, err := criteria.evaluateBool(exprShouldBeFalse, params, cache)
		if err != nil {
			// propagate expression error up, if happened
			return false, errors.NewErrorWithDetails(
				fmt.Sprintf("Can't evaluate 'require-node' in criteria: %s", err),
				errors.Details{
					"criteria":   criteria,
					"expression": exprShouldBeFalse,
				},
			)
		}
		if result {
			return false, nil
		}
	}

	// Make sure at least one "require-any" criteria evaluates to true
	if len(criteria.RequireAny) > 0 {
		for _, exprShouldBeTrue := range criteria.RequireAny {
			result, err := criteria.evaluateBool(exprShouldBeTrue, params, cache)
			if err != nil {
				// propagate expression error up, if happened
				return false, errors.NewErrorWithDetails(
					fmt.Sprintf("Can't evaluate 'require-any' in criteria: %s", err),
					errors.Details{
						"criteria":   criteria,
						"expression": exprShouldBeTrue,
					},
				)
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

// Evaluates bool expression, given a set of parameters and a cache. If cache is nil, it will still be evaluated
// successfully, but without a cache
func (criteria *Criteria) evaluateBool(expressionStr string, params *expression.Parameters, cache *expression.Cache) (bool, error) {
	if cache == nil {
		cache = expression.NewCache()
	}
	return cache.EvaluateAsBool(expressionStr, params)
}
