package expression

import (
	"strconv"
)

// Parameters is a set of named parameters for the expression
type Parameters map[string]interface{}

// NewParams creates a new set of named Parameters. Each parameter should be int, bool, or string.
func NewParams(stringParams map[string]string, structParams map[string]interface{}) *Parameters {
	// Populate parameter map
	result := Parameters{}

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
