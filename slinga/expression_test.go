package slinga

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestExpressions(t *testing.T) {
	labels := LabelSet{Labels: map[string]string{"foo": "10", "unusedLabel": "3", "a": "valueOfA"}};

	// simple case with integer variable
	assert.Equal(t, true, evaluate("foo > 5", labels), "Evaluate expression with integer");

	// simple case with string variable
	assert.Equal(t, true, evaluate("a == 'valueOfA'", labels), "Evaluate expression with string");

	// simple case with both variables
	assert.Equal(t, true, evaluate("foo > 5 && a == 'valueOfA'", labels), "Evaluate expression with both integer and string");

	// simple case with missing string variable
	assert.Equal(t, false, evaluate("foo > 5 && missingLabel == 'requiredValue'", labels), "Evaluate expression with missing string label");

	// simple case with missing integer variable
	assert.Equal(t, false, evaluate("foo > 5 && missingLabel == 239", labels), "Evaluate expression with missing integer label");

	// we are explicitly converting all integer-like labels to integers, so this will always be false (expected behavior)
	assert.Equal(t, false, evaluate("foo == '10'", labels), "All integer values are always converted to ints, they should never be equal to a string");
}
