package language

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLabelOperations(t *testing.T) {
	labelsBefore := NewLabelSet(map[string]string{"l1": "1", "l2": "2", "l3": "3"})

	ops := &LabelOperations{}
	(*ops)["set"] = map[string]string{"a": "b", "c": "d"}
	(*ops)["remove"] = map[string]string{"l1": ""}

	labelsAfter := labelsBefore.ApplyTransform(ops)

	assert.Equal(t, 4, len(labelsAfter.Labels), "Correct number of labels should be retained after transform")
	assert.Equal(t, "2", labelsAfter.Labels["l2"], "Label 'l2' should be retained")
	assert.Equal(t, "3", labelsAfter.Labels["l3"], "Label 'l3' should be retained")
	assert.Equal(t, "b", labelsAfter.Labels["a"], "Label 'a' should be added")
	assert.Equal(t, "d", labelsAfter.Labels["c"], "Label 'c' should be added")
	assert.Equal(t, "", labelsAfter.Labels["l1"], "Label 'l1' should not be present")

	labelsAfter = labelsAfter.ApplyTransform(nil)
	assert.Equal(t, 4, len(labelsAfter.Labels), "Correct number of labels should be retained after transform")
}
