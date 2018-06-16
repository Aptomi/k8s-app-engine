package template

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	ResSuccess      = iota
	ResCompileError = iota
	ResEvalError    = iota
)

func evaluate(t *testing.T, templateStr string, expectedResult int, expectedStr string, params *Parameters) {
	// Check for compilation
	tmpl, err := NewTemplate(templateStr)
	if !assert.Equal(t, expectedResult != ResCompileError, err == nil, "Template compilation (success vs. error): %s [%s]", templateStr, err) || expectedResult == ResCompileError {
		return
	}

	// Check for evaluation
	resultStr, err := tmpl.Evaluate(params)
	if !assert.Equal(t, expectedResult != ResEvalError, err == nil, "Template evaluation (success vs. error): %s [%s]", templateStr, err) || expectedResult == ResEvalError {
		return
	}

	// Check for result
	assert.Equal(t, expectedStr, resultStr, "Template evaluation result: %s", templateStr)
}

func evaluateWithCache(t *testing.T, templateStr string, expectedResult int, expectedStr string, params *Parameters, cache *Cache) {
	for i := 0; i < 10; i++ {
		// Check for compilation & evaluation
		resultStr, err := cache.Evaluate(templateStr, params)
		if !assert.Equal(t, expectedResult != ResCompileError && expectedResult != ResEvalError, err == nil, "[Cache] Template compilation (success vs. error): %s", templateStr) || expectedResult == ResCompileError || expectedResult == ResEvalError {
			return
		}

		// Check for result
		assert.Equal(t, expectedStr, resultStr, "[Cache] Template evaluation result: %s", templateStr)
	}
}

func TestTemplateEvaluation(t *testing.T) {
	params := NewParams(struct {
		Labels interface{}
		User   interface{}
	}{
		map[string]string{
			"tagname": "tagvalue",
		},

		struct {
			Labels map[string]string
		}{
			map[string]string{
				"team": "platform_services",
			},
		},
	})

	tests := []struct {
		template       string
		result         int
		expectedString string
	}{
		// successful evaluation
		{"test-{{.User.Labels.team}}-{{.Labels.tagname}}", ResSuccess, "test-platform_services-tagvalue"},
		{"val-{{ default \"abc\" .Labels.tagname }}", ResSuccess, "val-tagvalue"},
		{"val-{{ default \"abc\" .Labels.missinglabel }}", ResSuccess, "val-abc"},
		{"val-{{ default .Labels.tagname }}", ResSuccess, "val-tagvalue"},
		{"val-{{ default .Labels.missinglabel }}", ResSuccess, "val-"},

		// missing fields
		{"test-{{.User.MissingField}}-{{.MissingObject}}", ResEvalError, ""},
		{"test-{{.User.Labels.missinglabel}}", ResEvalError, ""},
		{"{{ default }}", ResEvalError, ""},
		{"{{ default \"a\" \"b\" \"c\" }}", ResEvalError, ""},

		// cannot be compiled
		{"{{ bs }}", ResCompileError, ""},
		{"{{ index a a a a a }}", ResCompileError, ""},
		{"{{ hello", ResCompileError, ""},
	}

	for _, test := range tests {
		evaluate(t, test.template, test.result, test.expectedString, params)
	}

	cache := NewCache()
	for _, test := range tests {
		evaluateWithCache(t, test.template, test.result, test.expectedString, params, cache)
	}

}
