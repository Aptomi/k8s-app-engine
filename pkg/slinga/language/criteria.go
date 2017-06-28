package language

import (
	. "github.com/Frostman/aptomi/pkg/slinga/log"
	"github.com/Knetic/govaluate"
	log "github.com/Sirupsen/logrus"
	"strconv"
	"strings"
)

// Criteria defines a structure with criteria accept/reject syntax
type Criteria struct {
	Accept []string
	Reject []string
}

// Whether criteria evaluates to "true" for a given set of labels or not
func (criteria *Criteria) allows(labels LabelSet) bool {
	// If one of the reject criterias matches, then it's not allowed
	for _, reject := range criteria.Reject {
		if evaluate(reject, labels) {
			return false
		}
	}

	// If one of the accept criterias matches, then it's allowed
	for _, reject := range criteria.Accept {
		if evaluate(reject, labels) {
			return true
		}
	}

	// If the accept section is empty, return true
	if len(criteria.Accept) == 0 {
		return true
	}

	return false
}

// Evaluate an expression, given a set of labels
func evaluate(expression string, params LabelSet) bool {
	// Create an expression
	expressionObject, e := govaluate.NewEvaluableExpression(expression)
	if e != nil {
		Debug.WithFields(log.Fields{
			"expression": expression,
			"error":      e,
		}).Panic("Invalid expression")
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
		Debug.WithFields(log.Fields{
			"expression": expression,
			"parameters": parameters,
			"error":      e,
		}).Panic("Cannot evaluate expression")
	}

	// Convert result to bool
	resultBool, ok := result.(bool)
	if !ok {
		Debug.WithFields(log.Fields{
			"expression": expression,
			"parameters": parameters,
			"result":     result,
		}).Panic("Expression doesn't evaluate to boolean")
	}

	return resultBool
}
