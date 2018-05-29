package template

import (
	"bytes"
	"fmt"
	"github.com/Aptomi/aptomi/pkg/errors"
	"reflect"
	"strings"
	t "text/template"
)

// Template struct contains text template string as well as its compiled version
type Template struct {
	templateStr      string
	templateCompiled *t.Template
}

// Custom functions
var textFuncMap = t.FuncMap{
	"default": func(args ...interface{}) interface{} {
		if len(args) == 0 || len(args) > 2 {
			// will fail text template execution
			return nil
		}

		// if one argument, return it
		if len(args) == 1 {
			value := args[0]
			if value == nil {
				return ""
			}
			return value
		}

		// otherwise first argument is default value and the second is actual value
		arg := args[0]
		value := args[1]
		if value == nil {
			return arg
		}

		v := reflect.ValueOf(value)
		switch v.Kind() {
		case reflect.String, reflect.Slice, reflect.Array, reflect.Map:
			if v.Len() == 0 {
				return arg
			}
		case reflect.Bool:
			if !v.Bool() {
				return arg
			}
		}
		return value
	},
}

// NewTemplate compiles a text template and returns the result in Template struct
// Parameter templateStr must follow syntax defined by text/template
func NewTemplate(templateStr string) (*Template, error) {
	templateCompiled, err := t.New("").Funcs(textFuncMap).Parse(templateStr)
	if err != nil {
		return nil, fmt.Errorf("unable to compile template '%s': %s", templateStr, err)
	}
	return &Template{
		templateStr:      templateStr,
		templateCompiled: templateCompiled,
	}, nil
}

// Evaluate evaluates a compiled text template given a set named parameters
func (template *Template) Evaluate(params *Parameters) (string, error) {
	// Evaluate
	var doc bytes.Buffer

	// Multiple executions of the same template can execute safely in parallel
	err := template.templateCompiled.Execute(&doc, params.params)
	if err != nil {
		return "", errors.NewErrorWithDetails(
			fmt.Sprintf("unable to evaluate template '%s': %s", template.templateStr, err),
			errors.Details{
				"params": params,
			},
		)
	}

	// Convert result to bool
	result := doc.String()
	if strings.Contains(result, "<no value>") {
		return "", errors.NewErrorWithDetails(
			fmt.Sprintf("unable to evaluate template '%s': <no value>", template.templateStr),
			errors.Details{
				"params": params,
			},
		)
	}

	return result, nil
}
