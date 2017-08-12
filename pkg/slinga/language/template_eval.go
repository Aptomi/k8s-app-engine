package language

import (
	"errors"
	. "github.com/Aptomi/aptomi/pkg/slinga/util"
	"github.com/Aptomi/aptomi/pkg/slinga/language/template"
)

type templateData struct {
	Labels    map[string]string
	User      *User
	Discovery NestedParameterMap
}

// Evaluates a template
func evaluateTemplate(templateStr string, user *User, labels LabelSet) (string, error) {
	type Parameters struct {
		User   *User
		Labels map[string]string
	}
	params := Parameters{User: user, Labels: labels.Labels}
	t, err := template.NewTemplate(templateStr)
	if err != nil {
		return "", err
	}
	return t.Evaluate(template.NewTemplateParams(params))
}

func evaluateTemplateAndEscape(templateStr string, tData templateData) (string, error) {
	t, err := template.NewTemplate(templateStr)
	if err != nil {
		return "", err
	}

	result, err := t.Evaluate(template.NewTemplateParams(tData))
	if err != nil {
		return "", err
	}
	return EscapeName(result), nil
}

// ProcessTemplateParams processes template params
func ProcessTemplateParams(template ParameterTree, componentKey string, labels LabelSet, user *User, discoveryTree NestedParameterMap) (NestedParameterMap, error) {
	if template == nil {
		return nil, nil
	}

	// Create a copy of discovery tree, so we can add our own instance into it
	discoveryTreeCopy := discoveryTree.MakeCopy()
	discoveryTreeCopy["instance"] = EscapeName(componentKey)

	tData := templateData{
		Labels:    labels.Labels,
		Discovery: discoveryTreeCopy,
		User:      user}

	var evalParamsInterface func(params ParameterTree) (interface{}, error)

	// TODO: this method needs to be fixed to use less interface{} :-)
	evalParamsInterface = func(params ParameterTree) (interface{}, error) {
		if params == nil {
			return "", nil
		} else if paramsMap, ok := params.(map[interface{}]interface{}); ok {
			resultMap := make(map[interface{}]interface{})

			for key, value := range paramsMap {
				evaluatedValue, err := evalParamsInterface(value)
				if err != nil {
					return nil, err
				}
				resultMap[key] = evaluatedValue
			}

			return resultMap, nil
		} else if paramsStr, ok := params.(string); ok {
			evaluatedValue, err := evaluateTemplateAndEscape(paramsStr, tData)
			// TODO: we may want to debug paramsStr -> evaluatedValue here
			if err != nil {
				return nil, err
			}
			return evaluatedValue, nil
		} else if paramsInt, ok := params.(int); ok {
			return paramsInt, nil
		} else if paramsBool, ok := params.(bool); ok {
			return paramsBool, nil
		}

		return nil, errors.New("There should be map[string]interface{} or string")
	}

	resultMap, err := evalParamsInterface(template)
	if err != nil {
		return nil, err
	}

	result := NestedParameterMap{}
	for k, v := range resultMap.(map[interface{}]interface{}) {
		result[k.(string)] = v
	}
	return result, err
}
