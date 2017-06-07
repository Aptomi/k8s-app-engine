package slinga

import (
	"fmt"
	"bytes"
	"text/template"
	"strings"
	"errors"
)

// Evaluates a template
func evaluateTemplate(templateStr string, user User, labels LabelSet) (string, error) {
	type Parameters struct {
		User   User
		Labels map[string]string
	}
	param := Parameters{User: user, Labels: labels.Labels}

	tmpl, err := template.New("").Parse(templateStr)
	if err != nil {
		return "", fmt.Errorf("Invalid template %s: %s", templateStr, err.Error())
	}

	var doc bytes.Buffer
	err = tmpl.Execute(&doc, param)

	if err != nil {
		return "", fmt.Errorf("Cannot evaluate template %s: %s", templateStr, err.Error())
	}

	result := doc.String()
	if strings.Contains(result, "<no value>") {
		return "", fmt.Errorf("Cannot evaluate template %s: <no value>", templateStr)
	}

	return doc.String(), nil
}

func (component *ServiceComponent) processTemplateParams(template interface{}, componentKey string, labels LabelSet, user User, discoveryTree NestedParameterMap, templateType string, depth int) (NestedParameterMap, error) {
	if template == nil {
		return nil, nil
	}
	tracing.Printf(depth+1, "Component: %s (%s)", component.Name, templateType)

	// Create a copy of discovery tree, so we can add our own instance into it
	discoveryTreeCopy := discoveryTree.makeCopy()
	discoveryTreeCopy["instance"] = EscapeName(componentKey)

	templateData := templateData{
		Labels:    labels.Labels,
		Discovery: discoveryTreeCopy,
		User:      user}

	var evalParamsInterface func(params interface{}) (interface{}, error)

	// TODO: this method needs to be fixed to use less interface{} :-)
	evalParamsInterface = func(params interface{}) (interface{}, error) {
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
			evaluatedValue, err := evaluateCodeParamTemplate(paramsStr, templateData)
			tracing.Printf(depth+2, "Parameter '%s': %s", paramsStr, evaluatedValue)
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

func evaluateCodeParamTemplate(templateStr string, templateData templateData) (string, error) {
	tmpl, err := template.New("").Parse(templateStr)
	if err != nil {
		return "", errors.New("Invalid template " + templateStr)
	}

	var doc bytes.Buffer
	err = tmpl.Execute(&doc, templateData)

	if err != nil {
		return "", errors.New("Cannot evaluate template " + templateStr)
	}

	result := doc.String()
	if strings.Contains(result, "<no value>") {
		return "", errors.New("Cannot evaluate template " + templateStr)
	}

	return EscapeName(doc.String()), nil
}
