package language

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestExpressions(t *testing.T) {
	labels := LabelSet{Labels: map[string]string{
		"foo":          "10",
		"unusedLabel":  "3",
		"a":            "valueOfA",
		"bar":          "true",
		"anotherbar":   "t",
		"service.Name": "myservicename",
	}}

	// simple case with bool variable
	assert.Equal(t, true, evaluate("anotherbar == true", labels), "Evaluate expression with boolean")

	// simple case with bool variable
	assert.Equal(t, true, evaluate("anotherbar", labels), "Evaluate expression with boolean")

	// simple case with bool variable
	// TODO: we need to fix this test
	// assert.Equal(t, true, evaluate("anotherbar == 't'", labels), "Evaluate expression with boolean")

	// simple case with bool variable
	assert.Equal(t, false, evaluate("anotherbar == 'p'", labels), "Evaluate expression with boolean")

	// simple case with bool variable
	assert.Equal(t, true, evaluate("bar == true", labels), "Evaluate expression with boolean")

	// simple case with bool variable
	assert.Equal(t, true, evaluate("bar", labels), "Evaluate expression with boolean")

	// simple case with integer variable
	assert.Equal(t, true, evaluate("foo > 5", labels), "Evaluate expression with integer")

	// simple case with string variable
	assert.Equal(t, true, evaluate("a == 'valueOfA'", labels), "Evaluate expression with string")

	// simple case with both variables
	assert.Equal(t, true, evaluate("foo > 5 && a == 'valueOfA'", labels), "Evaluate expression with both integer and string")

	// simple case with missing string variable
	assert.Equal(t, false, evaluate("foo > 5 && missingLabel == 'requiredValue'", labels), "Evaluate expression with missing string label")

	// simple case with missing integer variable
	assert.Equal(t, false, evaluate("foo > 5 && missingLabel == 239", labels), "Evaluate expression with missing integer label")

	// we are explicitly converting all integer-like labels to integers, so this will always be false (expected behavior)
	assert.Equal(t, false, evaluate("foo == '10'", labels), "All integer values are always converted to ints, they should never be equal to a string")

	// check that service name expression works
	assert.Equal(t, true, evaluate("service.Name == 'myservicename'", labels), "Check that service.Name reference works correctly")
	assert.Equal(t, false, evaluate("service.Name == 'incorrectservicename'", labels), "Check that service.Name reference works correctly")
}
