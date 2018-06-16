package lang

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLabelSetOperations(t *testing.T) {
	labels := NewLabelSet(map[string]string{"l1": "1", "l2": "2", "l3": "3"})

	ops := NewLabelOperations(
		map[string]string{"a": "b", "c": "d"},
		map[string]string{"l1": ""},
	)

	changed := labels.ApplyTransform(ops)
	assert.True(t, changed, "Labels should be changed")

	assert.Equal(t, 4, len(labels.Labels), "Correct number of labels should be retained after transform")
	assert.Equal(t, "2", labels.Labels["l2"], "Label 'l2' should be retained")
	assert.Equal(t, "3", labels.Labels["l3"], "Label 'l3' should be retained")
	assert.Equal(t, "b", labels.Labels["a"], "Label 'a' should be added")
	assert.Equal(t, "d", labels.Labels["c"], "Label 'c' should be added")
	assert.Equal(t, "", labels.Labels["l1"], "Label 'l1' should not be present")

	notChanged := labels.ApplyTransform(ops)
	assert.False(t, notChanged, "Labels should not be changed")

	assert.Equal(t, 4, len(labels.Labels), "Correct number of labels should be retained after transform")
}

func TestLabelSetOperationsSingleLabel(t *testing.T) {
	labels := NewLabelSet(map[string]string{"l1": "1", "l2": "2", "l3": "3"})
	ops := NewLabelOperationsSetSingleLabel("name", "value")

	changed := labels.ApplyTransform(ops)
	assert.True(t, changed, "Labels should be changed")

	labelsEqual := NewLabelSet(map[string]string{"l1": "1", "l2": "2", "l3": "3", "name": "value"})
	assert.True(t, labels.Equal(labelsEqual), "Label sets ops when setting a single label should work")
}

func TestLabelSetEquals(t *testing.T) {
	labels := NewLabelSet(map[string]string{"a": "1", "b": "2", "c": "3"})

	ops := NewLabelOperations(
		map[string]string{"d": "4", "e": "5", "a": "aValue"},
		map[string]string{"c": ""},
	)

	changed := labels.ApplyTransform(ops)
	assert.True(t, changed, "Labels should be changed")

	// check for equal
	labelsEqual := NewLabelSet(map[string]string{"a": "aValue", "b": "2", "d": "4", "e": "5"})
	assert.True(t, labels.Equal(labelsEqual), "Label sets should be equal")

	// check for not equal
	labelsNotEqual := NewLabelSet(map[string]string{"b": "2", "d": "someValue", "e": "5"})
	assert.False(t, labels.Equal(labelsNotEqual), "Label sets should not be equal")

	// check for equal (map with zero lengths and nil)
	labelsEmpty := NewLabelSet(map[string]string{})
	assert.True(t, labelsEmpty.Equal(NewLabelSet(nil)), "Empty label sets should be equal")
}

func TestLabelSetAdd(t *testing.T) {
	labels1 := NewLabelSet(map[string]string{"a": "1", "b": "2", "c": "3"})
	labels2 := NewLabelSet(map[string]string{"c": "4", "d": "5", "e": "6"})
	labels1.AddLabels(labels2.Labels)

	// check for equal
	labelsEqual := NewLabelSet(map[string]string{"a": "1", "b": "2", "c": "4", "d": "5", "e": "6"})
	assert.True(t, labels1.Equal(labelsEqual), "Label sets addition should work")
}

func TestLabelSetNotChanged(t *testing.T) {
	labels := NewLabelSet(map[string]string{"l1": "1", "l2": "2", "l3": "3"})

	ops := NewLabelOperations(
		map[string]string{"l1": "1", "l3": "3"},
		map[string]string{"l4": ""},
	)

	notChanged := labels.ApplyTransform(ops)
	assert.False(t, notChanged, "Labels should not be changed")
}
