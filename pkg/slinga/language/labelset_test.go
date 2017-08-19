package language

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLabelSetOperations(t *testing.T) {
	labelsBefore := NewLabelSet(map[string]string{"l1": "1", "l2": "2", "l3": "3"})

	ops := NewLabelOperations(
		map[string]string{"a": "b", "c": "d"},
		map[string]string{"l1": ""},
	)

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

func TestLabelSetOperationsSingleLabel(t *testing.T) {
	labelsBefore := NewLabelSet(map[string]string{"l1": "1", "l2": "2", "l3": "3"})
	ops := NewLabelOperationsSetSingleLabel("name", "value")

	labelsAfter := labelsBefore.ApplyTransform(ops)
	labelsEqual := NewLabelSet(map[string]string{"l1": "1", "l2": "2", "l3": "3", "name": "value"})
	assert.True(t, labelsAfter.Equal(labelsEqual), "Label sets ops when setting a single label should work")
}

func TestLabelSetEquals(t *testing.T) {
	labelsBefore := NewLabelSet(map[string]string{"a": "1", "b": "2", "c": "3"})

	ops := NewLabelOperations(
		map[string]string{"d": "4", "e": "5", "a": "aValue"},
		map[string]string{"c": ""},
	)

	labelsAfter := labelsBefore.ApplyTransform(ops)

	// check for equal
	labelsEqual := NewLabelSet(map[string]string{"a": "aValue", "b": "2", "d": "4", "e": "5"})
	assert.True(t, labelsAfter.Equal(labelsEqual), "Label sets should be equal")

	// check for not equal
	labelsNotEqual := NewLabelSet(map[string]string{"b": "2", "d": "someValue", "e": "5"})
	assert.False(t, labelsAfter.Equal(labelsNotEqual), "Label sets should not be equal")

	// check for equal (map with zero lengths and nil)
	labelsEmpty := NewLabelSet(map[string]string{})
	assert.True(t, labelsEmpty.Equal(NewLabelSet(nil)), "Empty label sets should be equal")
}

func TestLabelSetAdd(t *testing.T) {
	labels1 := NewLabelSet(map[string]string{"a": "1", "b": "2", "c": "3"})
	labels2 := NewLabelSet(map[string]string{"c": "4", "d": "5", "e": "6"})
	labelsAfter := labels1.AddLabels(labels2)

	// check for equal
	labelsEqual := NewLabelSet(map[string]string{"a": "1", "b": "2", "c": "4", "d": "5", "e": "6"})
	assert.True(t, labelsAfter.Equal(labelsEqual), "Label sets addition should work")
}

