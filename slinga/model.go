package slinga

/*
 	This file declares all utility structures and methods required for Slinga processing
  */

// Set of labels that will be manipulated
type LabelSet struct {
	Labels map[string]string
}

// Apply set of transformations to labels
func (user *User) getLabelSet() LabelSet {
	return LabelSet{Labels: user.Labels}
}
// Apply set of transformations to labels
func (src *LabelSet) applyTransform(ops LabelOperations) LabelSet {
	result := LabelSet{Labels: make(map[string]string)}

	// copy original labels
	for k, v := range src.Labels {
		result.Labels[k] = v;
	}

	// set labels
	for k, v := range ops["set"] {
		result.Labels[k] = v;
	}

	// remove labels
	for k, _ := range ops["remove"] {
		delete(result.Labels, k);
	}

	return result
}

// Check if context criteria is satisfied
func (context *Context) matches(labels LabelSet) bool {
	for _, c := range context.Criteria {
		if evaluate(c, labels) {
			return true
		}
	}
	return false
}

// Check if allocation criteria is satisfied
func (allocation *Allocation) matches(labels LabelSet) bool {
	for _, c := range allocation.Criteria {
		if evaluate(c, labels) {
			return true
		}
	}
	return false
}
