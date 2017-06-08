package slinga

// LabelSet defines the set of labels that will be manipulated
type LabelSet struct {
	Labels map[string]string
}

// Apply set of transformations to labels
func (user *User) getLabelSet() LabelSet {
	return LabelSet{Labels: user.Labels}
}

// Apply set of transformations to labels
func (src *LabelSet) applyTransform(ops *LabelOperations) LabelSet {
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
func (src LabelSet) addLabels(ops LabelSet) LabelSet {
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
