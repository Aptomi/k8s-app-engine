package language

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/slinga/language/expression"
	"github.com/Aptomi/aptomi/pkg/slinga/language/template"
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	ResTrue  = iota
	ResFalse = iota
	ResError = iota
)

func match(t *testing.T, context *Context, params *expression.ExpressionParameters, expected int, cache expression.ExpressionCache) {
	result, err := context.Matches(params, cache)
	assert.Equal(t, expected == ResError, err != nil, "Context matching (success vs. error): "+fmt.Sprintf("%+v, params %+v", context.Criteria, params))
	if err == nil {
		assert.Equal(t, expected == ResTrue, result, "Context matching: "+fmt.Sprintf("%+v, params %+v", context.Criteria, params))
	}
}

func matchContext(t *testing.T, context *Context, paramsMatch []*expression.ExpressionParameters, paramsDoesntMatch []*expression.ExpressionParameters) {
	// Evaluate with and without cache
	cache := expression.NewExpressionCache()
	for _, params := range paramsMatch {
		match(t, context, params, ResTrue, nil)
		match(t, context, params, ResTrue, cache)
	}
	for _, params := range paramsDoesntMatch {
		match(t, context, params, ResFalse, nil)
		match(t, context, params, ResFalse, cache)
	}
}

func evalKeys(t *testing.T, context *Context, params *template.TemplateParameters, expectedError bool, expected []string, cache template.TemplateCache) {
	keys, err := context.ResolveKeys(params, cache)
	assert.Equal(t, expectedError, err != nil, "Allocation key evaluation (success vs. error). Context: "+fmt.Sprintf("%+v, params %+v", context, params))
	if err == nil {
		assert.Equal(t, expected, keys, "Allocation key resolution: "+fmt.Sprintf("%+v, params %+v", context.Allocation, params))
	}
}

func TestServiceContextMatching(t *testing.T) {
	policy := LoadUnitTestsPolicy()

	// Test prod-high context
	context := policy.Contexts["prod-high"]

	// Params which result in matching
	paramsMatch := []*expression.ExpressionParameters{
		expression.NewExpressionParams(
			map[string]string{
				"dev":      "no",
				"prod":     "yes",
				"priority": "200",
			},
			nil,
		),
	}

	// Params which don't result in matching
	paramsDoesntMatch := []*expression.ExpressionParameters{
		expression.NewExpressionParams(
			map[string]string{
				"dev":         "no",
				"prod":        "yes",
				"priority":    "200",
				"nozookeeper": "true",
			},

			map[string]interface{}{
				"service": struct {
					Name string
				}{
					Name: "zookeeper",
				},
			},
		),

		expression.NewExpressionParams(
			map[string]string{
				"dev":      "no",
				"prod":     "yes",
				"priority": "10",
			},

			nil,
		),
	}

	matchContext(t, context, paramsMatch, paramsDoesntMatch)
}

func TestServiceContextRequireAnyFails(t *testing.T) {
	policy := LoadUnitTestsPolicy()
	context := policy.Contexts["special-not-matched"]
	paramsMatch := []*expression.ExpressionParameters{}
	paramsDoesntMatch := []*expression.ExpressionParameters{
		expression.NewExpressionParams(
			map[string]string{
				"never1": "a1",
				"never2": "a2",
			},

			nil,
		),
	}
	matchContext(t, context, paramsMatch, paramsDoesntMatch)
}

func TestServiceContextRequireAnyEmpty(t *testing.T) {
	policy := LoadUnitTestsPolicy()
	context := policy.Contexts["special-matched"]
	paramsMatch := []*expression.ExpressionParameters{
		expression.NewExpressionParams(
			map[string]string{
				"specialname": "specialvalue",
			},

			nil,
		),
	}
	paramsDoesntMatch := []*expression.ExpressionParameters{}
	matchContext(t, context, paramsMatch, paramsDoesntMatch)
}

func TestServiceContextKeyResolution(t *testing.T) {
	policy := LoadUnitTestsPolicy()

	// Test prod-high context
	context := policy.Contexts["prod-high"]

	// Params which result in successful key evaluation
	paramSuccess := template.NewTemplateParams(
		struct {
			User interface{}
		}{
			User: struct {
				Name string
			}{
				"actualvalue",
			},
		},
	)

	// Params which result in unsuccessful key evaluation
	paramFailure := template.NewTemplateParams(
		struct {
			User interface{}
		}{
			User: struct {
				Noname string
			}{
				"novalue",
			},
		},
	)

	// Evaluate with and without cache
	cache := template.NewTemplateCache()

	// Success
	evalKeys(t, context, paramSuccess, false, []string{"actualvalue"}, nil)
	evalKeys(t, context, paramSuccess, false, []string{"actualvalue"}, cache)

	// Failure
	evalKeys(t, context, paramFailure, true, nil, nil)
	evalKeys(t, context, paramFailure, true, nil, cache)
}
