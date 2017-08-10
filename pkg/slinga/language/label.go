package language

import "reflect"

// LabelSet defines the set of labels that will be manipulated
type LabelSet struct {
	Labels map[string]string
}

// GetLabelSet returns a set of user labels
func (user *User) GetLabelSet() LabelSet {
	return LabelSet{Labels: user.Labels}
}

// GetSecretSet returns a set of user secrets
func (user *User) GetSecretSet() LabelSet {
	return LabelSet{Labels: user.Secrets}
}

// ApplyTransform applies set of transformations to labels
func (src *LabelSet) ApplyTransform(ops *LabelOperations) LabelSet {
	result := LabelSet{Labels: make(map[string]string)}

	// copy original labels
	for k, v := range src.Labels {
		result.Labels[k] = v
	}

	if ops != nil {
		// set labels
		for k, v := range (*ops)["set"] {
			result.Labels[k] = v
		}

		// remove labels
		for k := range (*ops)["remove"] {
			delete(result.Labels, k)
		}
	}

	return result
}

// AddLabels merges two sets of labels and returns a new merged set
func (src LabelSet) AddLabels(ops LabelSet) LabelSet {
	result := LabelSet{Labels: make(map[string]string)}

	// copy original labels
	for k, v := range src.Labels {
		result.Labels[k] = v
	}

	// put new labels
	for k, v := range ops.Labels {
		result.Labels[k] = v
	}

	return result
}

// Equal compares two labels sets. If one is nil and another one is empty, it will return true as well
func (src LabelSet) Equal(dst LabelSet) bool {
	if len(src.Labels) == 0 && len(dst.Labels) == 0 {
		return true
	}
	return reflect.DeepEqual(src.Labels, dst.Labels)
}

// GetLabelSet returns a set of cluster labels
func (cluster *Cluster) GetLabelSet() LabelSet {
	return LabelSet{Labels: cluster.Labels}
}

