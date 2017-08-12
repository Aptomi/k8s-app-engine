package expression

import (
	"github.com/ralekseenkov/govaluate"
	"errors"
)

type Expression struct {
	expressionStr      string
	expressionCompiled *govaluate.EvaluableExpression
}

func NewExpression(expressionStr string) (*Expression, error) {
	expressionCompiled, e := govaluate.NewEvaluableExpression(expressionStr)
	if e != nil {
		return nil, e
	}
	return &Expression{
		expressionStr:      expressionStr,
		expressionCompiled: expressionCompiled,
	}, nil
}

// Evaluate an expression, given a set of labels
func (expression *Expression) EvaluateAsBool(params *ExpressionParameters) (bool, error) {
	// Evaluate
	result, e := expression.expressionCompiled.Evaluate(*params)
	if e != nil {
		// Return false and swallow the error if we encountered a missing parameter
		if _, ok := e.(*govaluate.MissingParameterError); ok {
			return false, nil
		}
		return false, e
	}

	// Convert result to bool
	value, ok := result.(bool)
	if !ok {
		return false, errors.New("Expression doesn't evaluate to boolean: " + expression.expressionStr)
	}

	return value, nil
}
