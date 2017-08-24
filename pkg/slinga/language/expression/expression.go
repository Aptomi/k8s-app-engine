package expression

import (
	"github.com/ralekseenkov/govaluate"
	"fmt"
	"github.com/Aptomi/aptomi/pkg/slinga/errors"
)

type Expression struct {
	expressionStr      string
	expressionCompiled *govaluate.EvaluableExpression
}

func NewExpression(expressionStr string) (*Expression, error) {
	expressionCompiled, err := govaluate.NewEvaluableExpression(expressionStr)
	if err != nil {
		return nil, fmt.Errorf("Unable to compile expression '%s': %s", expressionStr, err.Error())
	}
	return &Expression{
		expressionStr:      expressionStr,
		expressionCompiled: expressionCompiled,
	}, nil
}

// Evaluate an expression, given a set of labels
func (expression *Expression) EvaluateAsBool(params *ExpressionParameters) (bool, error) {
	// Evaluate
	result, err := expression.expressionCompiled.Evaluate(*params)
	if err != nil {
		// Return false and swallow the error if we encountered a missing parameter
		if _, ok := err.(*govaluate.MissingParameterError); ok {
			return false, nil
		}
		return false, errors.NewErrorWithDetails(
			fmt.Sprintf("Unable to evaluate expression '%s': %s", expression.expressionStr, err.Error()),
			errors.Details{
				"expression": expression.expressionStr,
				"params": params,
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
				"params": params,
			},
		)
	}

	return value, nil
}
