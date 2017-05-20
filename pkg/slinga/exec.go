package slinga

import (
	"bytes"
	"errors"
	"strings"
	"text/template"
)

type CodeExecutor interface {
	Install(key string, labels LabelSet) error
	Update(key string, labels LabelSet) error
	Destroy(key string) error
}

func (code *Code) GetCodeExecutor() (CodeExecutor, error) {
	switch code.Type {
	case "aptomi/code/kubernetes-helm", "kubernetes-helm":
		return HelmCodeExecutor{code}, nil
	case "aptomi/code/fake", "fake":
		return FakeCodeExecutor{code}, nil
	default:
		return nil, errors.New("CodeExecutor not found: " + code.Type)
	}
}

func (code *Code) processCodeContent(labels LabelSet) (map[string]map[string]string, error) {
	result := make(map[string]map[string]string)
	for section, params := range code.Content {
		result[section] = make(map[string]string)
		for key, value := range params {
			evaluatedParam, err := evaluateCodeParamTemplate(value, labels)
			if err != nil {
				return nil, err
			}

			result[section][key] = evaluatedParam
		}
	}
	return result, nil
}

func evaluateCodeParamTemplate(templateStr string, labels LabelSet) (string, error) {
	type Parameters struct {
		Labels map[string]string
	}
	param := Parameters{Labels: labels.Labels}

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

	return doc.String(), nil
}
