package template

import (
	"bytes"
	"fmt"
	"strings"
	t "text/template"
)

type Template struct {
	templateStr      string
	templateCompiled *t.Template
}

func NewTemplate(templateStr string) (*Template, error) {
	templateCompiled, err := t.New("").Parse(templateStr)
	if err != nil {
		return nil, fmt.Errorf("Cannot compile template %s: %s", templateStr, err.Error())
	}
	return &Template{
		templateStr:      templateStr,
		templateCompiled: templateCompiled,
	}, nil
}

// Evaluate an expression, given a set of labels
func (template *Template) Evaluate(params *TemplateParameters) (string, error) {
	// Evaluate
	var doc bytes.Buffer
	err := template.templateCompiled.Execute(&doc, params.params)
	if err != nil {
		return "", fmt.Errorf("Cannot evaluate template %s: %s", template.templateStr, err.Error())
	}

	// Convert result to bool
	result := doc.String()
	if strings.Contains(result, "<no value>") {
		return "", fmt.Errorf("Cannot evaluate template %s: <no value>", template.templateStr)
	}

	return result, nil
}
