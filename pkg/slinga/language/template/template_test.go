package template

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func evaluate(templateStr string, params *TemplateParameters) (string, error) {
	t, err := NewTemplate(templateStr)
	if err != nil {
		return "", err
	}
	return t.Evaluate(params)
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

	result, err := evaluate("test-{{.User.Labels.team}}-{{.Labels.tagname}}", params)
	assert.Nil(t, err, "Template should evaluate without errors")
	assert.Equal(t, "test-platform_services-tagvalue", result, "Template should be evaluated correctly, user team parameter must be substituted with its value")

	result, err = evaluate("test-{{.User.MissingField}}-{{.MissingObject}}", params)
	assert.NotNil(t, err, "Template should not evaluate, because there is a missing field")

	result, err = evaluate("test-{{.User.Labels.missinglabel}}", params)
	assert.NotNil(t, err, "Template should not evaluate, because there is a missing label")
}
