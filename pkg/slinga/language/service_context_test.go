package language

import (
	"testing"
	"github.com/Aptomi/aptomi/pkg/slinga/language/expression"
	"github.com/stretchr/testify/assert"
	"fmt"
	"github.com/Aptomi/aptomi/pkg/slinga/language/template"
)

func match(t *testing.T, context *Context, params *expression.ExpressionParameters, expected bool, cache expression.ExpressionCache) {
	assert.Equal(t, expected, context.Matches(params, cache), "Context matching: "+fmt.Sprintf("%+v, params %+v", context.Criteria, params))
}

func evalKeys(t *testing.T, context *Context, params *template.TemplateParameters, expectedError bool, expected []string, cache template.TemplateCache) {
	keys, err := context.ResolveKeys(params, cache)
	assert.Equal(t, expectedError, err != nil, "Allocation key evaluation (success vs. error). Context: "+fmt.Sprintf("%+v, params %+v", context, params))
	if err == nil {
		assert.Equal(t, expected, keys, "Allocation key resolution: "+fmt.Sprintf("%+v, params %+v", context.Allocation, params))
	}
}

func TestServiceContextMatching(t *testing.T) {
	policy := loadUnitTestsPolicy()

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
					Metadata map[string]string
				}{
					map[string]string{
						"Name": "zookeeper",
					},
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

	// Evaluate with and without cache
	cache := expression.NewExpressionCache()
	for _, params := range paramsMatch {
		match(t, context, params, true, nil)
		match(t, context, params, true, cache)
	}
	for _, params := range paramsDoesntMatch {
		match(t, context, params, false, nil)
		match(t, context, params, false, cache)
	}
}

func TestServiceContextKeyResolution(t *testing.T) {
	policy := loadUnitTestsPolicy()

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
