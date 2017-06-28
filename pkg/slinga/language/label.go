package language

import "reflect"

// LabelSet defines the set of labels that will be manipulated
type LabelSet struct {
	Labels map[string]string
}

// Returns a set of user labels
func (user *User) GetLabelSet() LabelSet {
	return LabelSet{Labels: user.Labels}
}

// Returns a set of user secrets
func (user *User) GetSecretSet() LabelSet {
	return LabelSet{Labels: user.Secrets}
}

// Apply set of transformations to labels
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

// Merge two sets of labels
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

// Function to compare two labels sets. If one is nil and another one is empty, it will return true as well
func (src LabelSet) Equal(dst LabelSet) bool {
	if len(src.Labels) == 0 && len(dst.Labels) == 0 {
		return true
	}
	return reflect.DeepEqual(src.Labels, dst.Labels)
}

func (cluster *Cluster) GetLabelSet() LabelSet {
	return LabelSet{Labels: cluster.Labels}
}
