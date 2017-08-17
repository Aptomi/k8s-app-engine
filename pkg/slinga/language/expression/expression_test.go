package expression

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func evaluate(t *testing.T, expressionStr string, params *ExpressionParameters) bool {
	expr, err := NewExpression(expressionStr)
	if err != nil {
		t.Logf("Expression didn't compile correctly: %s (%v)", expressionStr, err)
		t.Fail()
	}
	result, err := expr.EvaluateAsBool(params)
	if err != nil {
		t.Logf("Expression evaluated with errors: %s (%v)", expressionStr, err)
		t.Fail()
	}
	return result
}

func TestExpressions(t *testing.T) {
	params := NewExpressionParams(
		map[string]string{
			"foo":         "10",
			"unusedLabel": "3",
			"a":           "valueOfA",
			"bar":         "true",
			"anotherbar":  "t",
		},

		map[string]interface{}{
			"service": struct {
				Name   string
				Labels map[string]string
			}{
				"myservicename",
				map[string]string{
					"Name": "Value",
				},
			},
		},
	)

	// simple case with bool variable
	assert.Equal(t, true, evaluate(t, "anotherbar == true", params), "Evaluate expression with boolean")

	// simple case with bool variable
	assert.Equal(t, true, evaluate(t, "anotherbar", params), "Evaluate expression with boolean")

	// simple case with bool variable
	// TODO: we need to fix this test
	// assert.Equal(t, true, evaluate(t, "anotherbar == 't'", params), "Evaluate expression with boolean")

	// simple case with bool variable
	assert.Equal(t, false, evaluate(t, "anotherbar == 'p'", params), "Evaluate expression with boolean")

	// simple case with bool variable
	assert.Equal(t, true, evaluate(t, "bar == true", params), "Evaluate expression with boolean")

	// simple case with bool variable
	assert.Equal(t, true, evaluate(t, "bar", params), "Evaluate expression with boolean")

	// simple case with integer variable
	assert.Equal(t, true, evaluate(t, "foo > 5", params), "Evaluate expression with integer")

	// simple case with string variable
	assert.Equal(t, true, evaluate(t, "a == 'valueOfA'", params), "Evaluate expression with string")

	// simple case with both variables
	assert.Equal(t, true, evaluate(t, "foo > 5 && a == 'valueOfA'", params), "Evaluate expression with both integer and string")

	// simple case with missing string variable
	assert.Equal(t, false, evaluate(t, "foo > 5 && missingLabel == 'requiredValue'", params), "Evaluate expression with missing string label")

	// simple case with missing integer variable
	assert.Equal(t, false, evaluate(t, "foo > 5 && missingLabel == 239", params), "Evaluate expression with missing integer label")

	// we are explicitly converting all integer-like params to integers, so this will always be false (expected behavior)
	assert.Equal(t, false, evaluate(t, "foo == '10'", params), "All integer values are always converted to ints, they should never be equal to a string")

	// check that struct expressions work
	assert.Equal(t, true, evaluate(t, "service.Name == 'myservicename'", params), "Check that struct.Name expression works correctly")
	assert.Equal(t, false, evaluate(t, "service.Name == 'incorrectservicename'", params), "Check that struct.Name expression works correctly")
	assert.Equal(t, true, evaluate(t, "service.Labels.Name + 'B' == 'ValueB'", params), "Check that struct.Labels expression works correctly")
}
