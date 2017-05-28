package slinga

import (
	"github.com/Knetic/govaluate"
	"strconv"
	"strings"
	log "github.com/Sirupsen/logrus"
)

// Evaluate an expression, given a set of labels
func evaluate(expression string, params LabelSet) bool {
	// Create an expression
	expressionObject, e := govaluate.NewEvaluableExpression(expression)
	if e != nil {
		debug.WithFields(log.Fields{
			"expression": expression,
			"error": e,
		}).Fatal("Invalid expression")
	}

	// Populate parameter map
	parameters := make(map[string]interface{}, len(params.Labels))
	for k, v := range params.Labels {
		// all labels are strings. we need to cast them to the appropriate type before evaluation
		if vInt, err := strconv.Atoi(v); err == nil {
			parameters[k] = vInt
		} else if vBool, err := strconv.ParseBool(v); err == nil {
			parameters[k] = vBool
		} else {
			parameters[k] = v
		}
	}

	// Evaluate
	result, e := expressionObject.Evaluate(parameters)
	if e != nil {
		// see if it's missing parameter? then return false
		// TODO: this is a hack to deal with missing labels. Will need to rewrite it
		if strings.Contains(e.Error(), "No parameter") && strings.Contains(e.Error(), "found") {
			return false
		}
		debug.WithFields(log.Fields{
			"expression": expression,
			"parameters": parameters,
			"error": e,
		}).Fatal("Cannot evaluate expression")
	}

	// Convert result to bool
	resultBool, ok := result.(bool)
	if !ok {
		debug.WithFields(log.Fields{
			"expression": expression,
			"parameters": parameters,
			"result": result,
		}).Fatal("Expression doesn't evaluate to boolean")
	}

	return resultBool
}
