package slinga

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTemplateEvaluation(t *testing.T) {
	alice := LoadUserByIDFromDir("testdata/unittests", "1")
	labels := LabelSet{Labels: map[string]string{"tagname": "tagvalue"}}

	result, err := evaluateTemplate("test-{{.User.Labels.team}}-{{.Labels.tagname}}", alice, labels)
	assert.Nil(t, err, "Template should evaluate without errors")
	assert.Equal(t, "test-platform_services-tagvalue", result, "Template should be evaluated correctly, user team parameter must be substituted with its value")

	result, err = evaluateTemplate("test-{{.User.MissingField}}-{{.MissingObject}}", alice, labels)
	assert.NotNil(t, err, "Template should not evaluate, because there is a missing field")

	result, err = evaluateTemplate("test-{{.User.Labels.missinglabel}}", alice, labels)
	assert.NotNil(t, err, "Template should not evaluate, because there is a missing label")
}
