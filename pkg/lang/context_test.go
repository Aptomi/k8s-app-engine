package lang

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/lang/expression"
	"github.com/Aptomi/aptomi/pkg/lang/template"
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	ResTrue  = iota
	ResFalse = iota
	ResError = iota
)

func match(t *testing.T, context *Context, params *expression.Parameters, expected int, cache *expression.Cache) {
	t.Helper()
	result, err := context.Matches(params, cache)
	assert.Equal(t, expected == ResError, err != nil, "Context matching (success vs. error): "+fmt.Sprintf("%+v, params %+v", context.Criteria, params))
	if err == nil {
		assert.Equal(t, expected == ResTrue, result, "Context matching: "+fmt.Sprintf("%+v, params %+v", context.Criteria, params))
	}
}

func matchContext(t *testing.T, context *Context, paramsMatch []*expression.Parameters, paramsDoesntMatch []*expression.Parameters, paramsError []*expression.Parameters) {
	// Evaluate with and without cache
	t.Helper()
	cache := expression.NewCache()
	for _, params := range paramsMatch {
		match(t, context, params, ResTrue, nil)
		match(t, context, params, ResTrue, cache)
	}
	for _, params := range paramsDoesntMatch {
		match(t, context, params, ResFalse, nil)
		match(t, context, params, ResFalse, cache)
	}
	for _, params := range paramsError {
		match(t, context, params, ResError, nil)
		match(t, context, params, ResError, cache)
	}
}

func evalKeys(t *testing.T, context *Context, params *template.Parameters, expectedError bool, expected []string, cache *template.Cache) {
	t.Helper()
	keys, err := context.ResolveKeys(params, cache)
	assert.Equal(t, expectedError, err != nil, "Allocation key evaluation (success vs. error). Context: "+fmt.Sprintf("%+v, params %+v", context, params))
	if err == nil {
		assert.Equal(t, expected, keys, "Allocation key resolution: "+fmt.Sprintf("%+v, params %+v", context.Allocation, params))
	}
}

func TestServiceContextMatching(t *testing.T) {
	context := &Context{
		Name: "context",
		Criteria: &Criteria{
			RequireAll: []string{"dev == 'no' && prod == 'yes' && priority >= 200"},
			RequireAny: []string{
				"priority > 0",
				"prod == 'yes'",
				"dev == 'no'",
			},
		},
	}

	// Params which result in matching
	paramsMatch := []*expression.Parameters{
		expression.NewParams(
			map[string]string{
				"dev":      "no",
				"prod":     "yes",
				"priority": "200",
			},
			nil,
		),
	}

	// Params which don't result in matching
	paramsDoesntMatch := []*expression.Parameters{
		expression.NewParams(
			map[string]string{
				"dev":      "yes",
				"prod":     "no",
				"priority": "500",
			},

			map[string]interface{}{
				"pname": struct {
					Name string
				}{
					Name: "pvalue",
				},
			},
		),

		expression.NewParams(
			map[string]string{
				"dev":      "no",
				"prod":     "yes",
				"priority": "10",
			},

			nil,
		),
	}

	matchContext(t, context, paramsMatch, paramsDoesntMatch, nil)
}

func TestServiceContextRequireAnyFails(t *testing.T) {
	context := &Context{
		Name: "special-not-matched",
		Criteria: &Criteria{
			RequireAll: []string{"true"},
			RequireAny: []string{
				"never1 == 'unbeliveable_value_1'",
				"never2 == 'unbeliveable_value_2'",
				"never3 == 'unbeliveable_value_3'",
			},
			RequireNone: []string{"false"},
		},
	}
	paramsDoesntMatch := []*expression.Parameters{
		expression.NewParams(
			map[string]string{
				"never1": "a1",
				"never2": "a2",
			},

			nil,
		),
	}
	matchContext(t, context, nil, paramsDoesntMatch, nil)
}

func TestServiceContextRequireAnyEmpty(t *testing.T) {
	context := &Context{
		Name: "special-matched",
		Criteria: &Criteria{
			RequireAll:  []string{"specialname == 'specialvalue'"},
			RequireNone: []string{"false"},
		},
	}
	paramsMatch := []*expression.Parameters{
		expression.NewParams(
			map[string]string{
				"specialname": "specialvalue",
			},

			nil,
		),
	}
	matchContext(t, context, paramsMatch, nil, nil)
}

func TestServiceContextEmptyCriteria(t *testing.T) {
	context := &Context{}
	paramsMatch := []*expression.Parameters{
		expression.NewParams(
			map[string]string{
				"somename": "somevalue",
			},

			nil,
		),
	}
	matchContext(t, context, paramsMatch, nil, nil)
}

func makeInvalidContexts() []*Context {
	return []*Context{
		{
			Name: "special-invalid-context-require-all",
			Criteria: &Criteria{
				RequireAll: []string{"specialname + '123')((("},
			},
		},
		{

			Name: "special-invalid-context-require-any",
			Criteria: &Criteria{
				RequireAny: []string{"specialname + '456')((("},
			},
		},
		{
			Name: "special-invalid-context-require-none",
			Criteria: &Criteria{
				RequireNone: []string{"specialname + '789')((("},
			},
		},
	}
}

func TestServiceContextInvalidCriteria(t *testing.T) {
	contexts := makeInvalidContexts()
	paramsError := []*expression.Parameters{
		expression.NewParams(
			map[string]string{
				"specialname": "specialvalue",
			},

			nil,
		),
	}
	for _, context := range contexts {
		matchContext(t, context, nil, nil, paramsError)
	}
}

func TestServiceContextKeyResolution(t *testing.T) {
	context := &Context{
		Name: "context",
		Criteria: &Criteria{
			RequireAll: []string{"true"},
		},
		Allocation: &Allocation{
			Service: "test",
			Keys: []string{
				"{{.User.Name}}",
			},
		},
	}

	// Params which result in successful key evaluation
	paramSuccess := template.NewParams(
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
	paramFailure := template.NewParams(
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
	cache := template.NewCache()

	// Success
	evalKeys(t, context, paramSuccess, false, []string{"actualvalue"}, nil)
	evalKeys(t, context, paramSuccess, false, []string{"actualvalue"}, cache)

	// Failure
	evalKeys(t, context, paramFailure, true, nil, nil)
	evalKeys(t, context, paramFailure, true, nil, cache)
}
