package expression

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/errors"
	"github.com/ralekseenkov/govaluate"
)

// Expression struct contains expression string as well as its compiled version
type Expression struct {
	expressionStr      string
	expressionCompiled *govaluate.EvaluableExpression
}

// NewExpression compiles an expression and returns the result in Expression struct
func NewExpression(expressionStr string) (*Expression, error) {
	functions := map[string]govaluate.ExpressionFunction{
		"in": func(args ...interface{}) (interface{}, error) {
			if len(args) == 0 {
				return nil, fmt.Errorf("can't evaluate in() function when zero arguments supplied")
			}
			v := args[0]
			for i := 1; i < len(args); i++ {
				if v == args[i] {
					return true, nil
				}
			}
			return false, nil
		},
	}

	expressionCompiled, err := govaluate.NewEvaluableExpressionWithFunctions(expressionStr, functions)
	if err != nil {
		return nil, fmt.Errorf("unable to compile expression '%s': %s", expressionStr, err)
	}
	return &Expression{
		expressionStr:      expressionStr,
		expressionCompiled: expressionCompiled,
	}, nil
}

// EvaluateAsBool evaluates a compiled boolean expression given a set of parameters
func (expression *Expression) EvaluateAsBool(params *Parameters) (bool, error) {
	// Evaluate
	result, err := expression.expressionCompiled.Evaluate(*params)
	if err != nil {
		// Return false and swallow the error if we encountered a missing parameter
		if _, ok := err.(*govaluate.MissingParameterError); ok {
			return false, nil
		}
		return false, errors.NewErrorWithDetails(
			fmt.Sprintf("Unable to evaluate expression '%s': %s", expression.expressionStr, err),
			errors.Details{
				"expression": expression.expressionStr,
				"params":     params,
			},
		)
	}

	// Convert result to bool
	value, ok := result.(bool)
	if !ok {
		return false, errors.NewErrorWithDetails(
			fmt.Sprintf("Expression '%s' didn't evaluate to boolean", expression.expressionStr),
			errors.Details{
				"expression": expression.expressionStr,
				"params":     params,
			},
		)
	}

	return value, nil
}
