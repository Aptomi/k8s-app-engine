package slinga

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLabelOperations(t *testing.T) {
	labelsBefore := LabelSet{Labels: map[string]string{"l1": "1", "l2": "2", "l3": "3"}}

	ops := &LabelOperations{}
	(*ops)["set"] = map[string]string{"a": "b", "c": "d"}
	(*ops)["remove"] = map[string]string{"l1": ""}

	labelsAfter := labelsBefore.applyTransform(ops)

	assert.Equal(t, 4, len(labelsAfter.Labels), "Correct number of labels should be retained after transform")
	assert.Equal(t, "2", labelsAfter.Labels["l2"], "Label 'l2' should be retained")
	assert.Equal(t, "3", labelsAfter.Labels["l3"], "Label 'l3' should be retained")
	assert.Equal(t, "b", labelsAfter.Labels["a"], "Label 'a' should be added")
	assert.Equal(t, "d", labelsAfter.Labels["c"], "Label 'c' should be added")
	assert.Equal(t, "", labelsAfter.Labels["l1"], "Label 'l1' should not be present")

	labelsAfter = labelsAfter.applyTransform(nil)
	assert.Equal(t, 4, len(labelsAfter.Labels), "Correct number of labels should be retained after transform")
}

func TestTemplateEvaluation(t *testing.T) {
	alice := LoadUserByIDFromDir("testdata/unittests", "1")

	result, err := evaluateTemplate("test-{{.User.Labels.team}}", alice)
	assert.Nil(t, err, "Template should evaluate without errors")
	assert.Equal(t, "test-platform_services", result, "Template should be evaluated correctly, user team parameter must be substituted with its value")

	result, err = evaluateTemplate("test-{{.User.MissingField}}-{{.MissingObject}}", alice)
	assert.NotNil(t, err, "Template should not evaluate, because there is a missing field")

	result, err = evaluateTemplate("test-{{.User.Labels.missinglabel}}", alice)
	assert.NotNil(t, err, "Template should not evaluate, because there is a missing label")
}
