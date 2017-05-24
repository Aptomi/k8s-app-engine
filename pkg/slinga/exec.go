package slinga

import (
	"bytes"
	"errors"
	"strings"
	"text/template"
)

// CodeExecutor is an interface that allows to create different executors for component allocation (e.g. helm, kube.libsonnet, etc)
type CodeExecutor interface {
	Install(key string, labels LabelSet, dependencies map[string]string) error
	Update(key string, labels LabelSet) error
	Destroy(key string) error
}

// GetCodeExecutor returns an executor based on code.Type
func (code *Code) GetCodeExecutor() (CodeExecutor, error) {
	switch code.Type {
	case "aptomi/code/kubernetes-helm", "kubernetes-helm":
		return HelmCodeExecutor{code}, nil
	case "aptomi/code/unittests", "unittests":
		return FakeCodeExecutor{code}, nil
	default:
		return nil, errors.New("CodeExecutor not found: " + code.Type)
	}
}

func (code *Code) processCodeContent(labels LabelSet, dependencies map[string]string) (map[string]map[string]string, error) {
	result := make(map[string]map[string]string)
	for section, params := range code.Content {
		result[section] = make(map[string]string)
		for key, value := range params {
			evaluatedParam, err := evaluateCodeParamTemplate(value, labels, dependencies)
			if err != nil {
				return nil, err
			}

			result[section][key] = evaluatedParam
		}
	}
	return result, nil
}

func evaluateCodeParamTemplate(templateStr string, labels LabelSet, dependencies map[string]string) (string, error) {
	type Parameters struct {
		Labels       map[string]string
		Dependencies map[string]string
	}
	param := Parameters{Labels: labels.Labels, Dependencies: dependencies}

	tmpl, err := template.New("").Parse(templateStr)
	if err != nil {
		return "", errors.New("Invalid template " + templateStr)
	}

	var doc bytes.Buffer
	err = tmpl.Execute(&doc, param)

	if err != nil {
		return "", errors.New("Cannot evaluate template " + templateStr)
	}

	result := doc.String()
	if strings.Contains(result, "<no value>") {
		return "", errors.New("Cannot evaluate template " + templateStr)
	}

	// TODO(slukjanov): it's temporary solution, fix it later
	return HelmName(doc.String()), nil
}
