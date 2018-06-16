package expression

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	ResTrue         = iota
	ResFalse        = iota
	ResCompileError = iota
	ResEvalError    = iota
)

func evaluate(t *testing.T, expressionStr string, params *Parameters, expectedResult int) {
	// Check for compilation
	expr, err := NewExpression(expressionStr)
	if !assert.Equal(t, expectedResult != ResCompileError, err == nil, "Expression compilation (success vs. error): %s", expressionStr) || expectedResult == ResCompileError {
		return
	}

	// Check for evaluation
	result, err := expr.EvaluateAsBool(params)
	if !assert.Equal(t, expectedResult != ResEvalError, err == nil, "Expression evaluation (success vs. error): %s", expressionStr) || expectedResult == ResEvalError {
		return
	}

	// Check for result
	assert.Equal(t, expectedResult == ResTrue, result, "Expression evaluation result: %s", expressionStr)
}

func evaluateWithCache(t *testing.T, expressionStr string, params *Parameters, expectedResult int, cache *Cache) {
	// Check for compilation & evaluation
	for i := 0; i < 10; i++ {
		result, err := cache.EvaluateAsBool(expressionStr, params)
		if !assert.Equal(t, expectedResult != ResCompileError && expectedResult != ResEvalError, err == nil, "[Cache] Expression compilation && evaluation (success vs. error): %s", expressionStr) || expectedResult == ResCompileError || expectedResult == ResEvalError {
			return
		}

		// Check for result
		assert.Equal(t, expectedResult == ResTrue, result, "[Cache] Expression evaluation result: %s", expressionStr)
	}
}

func TestExpressions(t *testing.T) {
	params := NewParams(
		map[string]string{
			"foo":         "10",
			"unusedLabel": "3",
			"a":           "valueOfA",
			"bar":         "true",
			"anotherbar":  "t",
		},

		map[string]interface{}{
			"Service": struct {
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

	tests := []struct {
		expression string
		result     int
	}{
		// true (checking bool, int, string)
		{"anotherbar == true", ResTrue},
		{"anotherbar", ResTrue},
		{"bar == true", ResTrue},
		{"bar", ResTrue},
		{"foo > 5", ResTrue},
		{"a == 'valueOfA'", ResTrue},
		{"foo > 5 && a == 'valueOfA'", ResTrue},

		// false
		{"anotherbar == 'p'", ResFalse},
		{"'A' + 'B' == 5", ResFalse},

		// IN function
		{"in(a, 'valueOfC', 'valueOfB', 'valueOfA')", ResTrue},
		{"in(foo, 10, 20, 30)", ResTrue},
		{"in(a, 'valueOfX', 'valueOfY', 'valueOfZ')", ResFalse},
		{"in(a, )", ResCompileError},
		{"in()", ResEvalError},
		{"in(5)", ResFalse},

		// check when expression involves a missing label
		{"foo > 5 && missingLabel == 'requiredValue'", ResFalse},
		{"foo > 5 && missingLabel == 239", ResFalse},

		// we are explicitly converting all integer-like params to integers, so this should always be false (expected behavior)
		{"foo == '10'", ResFalse},

		// we are explicitly converting all bool-like params to bool, so this should always be false (expected behavior)
		{"anotherbar == 't'", ResFalse},

		// check that struct expressions work
		{"Service.Name == 'myservicename'", ResTrue},
		{"Service.Name == 'incorrectservicename'", ResFalse},
		{"Service.Labels.Name + 'B' == 'ValueB'", ResTrue},
		{"ServiceMissing.LabelsMissing.Name + 'B' == 'ValueB'", ResFalse},

		// evaluation error
		{"foo + 10 + 'test' > 0", ResEvalError},

		// not a boolean
		{"'a' + 'b' + bar", ResEvalError},

		// compilation error
		{"(5 + 10 > 9", ResCompileError},
	}

	// Evaluate without cache
	for _, test := range tests {
		evaluate(t, test.expression, params, test.result)
	}

	cache := NewCache()
	for _, test := range tests {
		evaluateWithCache(t, test.expression, params, test.result, cache)
	}
}
