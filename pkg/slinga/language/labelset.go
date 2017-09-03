package language

import "reflect"

// LabelSet defines the set of labels that will be manipulated
type LabelSet struct {
	Labels map[string]string
}

// NewLabelSet creates a new LabelSet from a given set of labels
func NewLabelSet(labels map[string]string) *LabelSet {
	result := &LabelSet{Labels: make(map[string]string)}
	result.AddLabels(labels)
	return result
}

// AddLabels adds new labels to the current set of labels
func (src *LabelSet) AddLabels(addMap map[string]string) {
	for k, v := range addMap {
		src.Labels[k] = v
	}
}

// ApplyTransform applies set of transformations to labels
// Returns true if changes have been made
func (src *LabelSet) ApplyTransform(ops LabelOperations) bool {
	changed := false
	if ops != nil {
		// set labels
		for k, v := range ops["set"] {
			if src.Labels[k] != v {
				src.Labels[k] = v
				changed = true
			}
		}

		// remove labels
		for k := range ops["remove"] {
			if _, exists := src.Labels[k]; exists {
				delete(src.Labels, k)
				changed = true
			}
		}
	}
	return changed
}

// Equal compares two labels sets. If one is nil and another one is empty, it will return true as well
// This method ignores IsSecret for now
func (src *LabelSet) Equal(dst *LabelSet) bool {
	if len(src.Labels) == 0 && len(dst.Labels) == 0 {
		return true
	}
	return reflect.DeepEqual(src.Labels, dst.Labels)
}
