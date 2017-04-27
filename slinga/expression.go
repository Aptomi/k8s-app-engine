package slinga

import (
	"github.com/Knetic/govaluate"
	"log"
	"strconv"
	"strings"
)

// Evaluate an
func evaluate(expression string, params LabelSet) bool {
	// Create an expression
	expressionObject, e := govaluate.NewEvaluableExpression(expression);
	if e != nil {
		log.Fatalf("Invalid expression: %v", e)
	}

	// Populate parameter map
	parameters := make(map[string]interface{}, len(params.Labels))
	for k, v := range params.Labels {
		// all labels are strings. we need to cast them to the appropriate type before evaluation

		if vInt, err := strconv.Atoi(v); err == nil {
			parameters[k] = vInt;
		} else {
			parameters[k] = v;
		}
	}

	// Evaluate
	result, e := expressionObject.Evaluate(parameters);
	if e != nil {
		// see if it's missing parameter? then return false
		if strings.Contains(e.Error(), "No parameter") && strings.Contains(e.Error(), "found") {
			return false
		}
		log.Fatalf("Cannot evaluate expression: %v", e)
	}

	// Convert result to bool
	resultBool, ok := result.(bool)
	if !ok {
		log.Fatalf("Expression doesn't evaluate to boolean: %v", result)
	}

	return resultBool
}
