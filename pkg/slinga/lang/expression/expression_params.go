package expression

import (
	"strconv"
)

type ExpressionParameters map[string]interface{}

func NewExpressionParams(stringParams map[string]string, structParams map[string]interface{}) *ExpressionParameters {
	// Populate parameter map
	result := ExpressionParameters{}

	for k, v := range stringParams {
		// string parameters have to be casted to the appropriate type before evaluation
		if vInt, err := strconv.Atoi(v); err == nil {
			result[k] = vInt
		} else if vBool, err := strconv.ParseBool(v); err == nil {
			result[k] = vBool
		} else {
			result[k] = v
		}
	}

	for k, v := range structParams {
		result[k] = v
	}

	return &result
}
