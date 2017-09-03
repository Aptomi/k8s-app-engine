package template

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	ResSuccess      = iota
	ResCompileError = iota
	ResEvalError    = iota
)

func evaluate(t *testing.T, templateStr string, expectedResult int, expectedStr string, params *TemplateParameters) {
	// Check for compilation
	tmpl, err := NewTemplate(templateStr)
	if !assert.Equal(t, expectedResult != ResCompileError, err == nil, "Template compilation (success vs. error): "+templateStr) || expectedResult == ResCompileError {
		return
	}

	// Check for evaluation
	resultStr, err := tmpl.Evaluate(params)
	if !assert.Equal(t, expectedResult != ResEvalError, err == nil, "Template evaluation (success vs. error): "+templateStr) || expectedResult == ResEvalError {
		return
	}

	// Check for result
	assert.Equal(t, expectedStr, resultStr, "Template evaluation result: "+templateStr)
}

func evaluateWithCache(t *testing.T, templateStr string, expectedResult int, expectedStr string, params *TemplateParameters, cache *TemplateCache) {
	// Check for compilation & evaluation
	resultStr, err := cache.Evaluate(templateStr, params)
	if !assert.Equal(t, expectedResult != ResCompileError && expectedResult != ResEvalError, err == nil, "[Cache] Template compilation (success vs. error): "+templateStr) || expectedResult == ResCompileError || expectedResult == ResEvalError {
		return
	}

	// Check for result
	assert.Equal(t, expectedStr, resultStr, "[Cache] Template evaluation result: "+templateStr)
}

func TestTemplateEvaluation(t *testing.T) {
	params := NewTemplateParams(struct {
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

		// missing fields
		{"test-{{.User.MissingField}}-{{.MissingObject}}", ResEvalError, ""},
		{"test-{{.User.Labels.missinglabel}}", ResEvalError, ""},

		// cannot be compiled
		{"{{ bs }}", ResCompileError, ""},
		{"{{ index a a a a a }}", ResCompileError, ""},
		{"{{ hello", ResCompileError, ""},
	}

	for _, test := range tests {
		evaluate(t, test.template, test.result, test.expectedString, params)
	}

	cache := NewTemplateCache()
	for _, test := range tests {
		evaluateWithCache(t, test.template, test.result, test.expectedString, params, cache)
	}

}
