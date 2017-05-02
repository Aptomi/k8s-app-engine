package slinga

/*
 	This file declares all the structures and methods required for Slinga processing (which don't exist in YAML)
  */

// Set of labels that will be manipulated
type LabelSet struct {
	Labels map[string]string
}

// Apply set of transformations to labels
func (src LabelSet) applyTransform(ops LabelOperations) LabelSet {
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