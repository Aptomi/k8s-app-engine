package slinga

/*
	This file declares all utility structures and methods required for Slinga processing
*/

// ParameterMultiMap allows to do [string][string]...[string] -> value
type NestedParameterMap map[string]interface{}

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

// Check if context criteria is satisfied
func (context *Context) matches(labels LabelSet) bool {
	return context.Criteria == nil || context.Criteria.allows(labels)
}

// Check if allocation criteria is satisfied
func (allocation *Allocation) matches(labels LabelSet) bool {
	return allocation.Criteria == nil || allocation.Criteria.allows(labels)
}

// Resolve name for an allocation
func (allocation *Allocation) resolveName(user User, labels LabelSet) error {
	result, err := evaluateTemplate(allocation.Name, user, labels)
	allocation.NameResolved = result
	return err
}

// Whether criteria evaluates to "true" for a given set of labels or not
func (criteria *Criteria) allows(labels LabelSet) bool {
	// If one of the reject criterias matches, then it's not allowed
	for _, reject := range criteria.Reject {
		if evaluate(reject, labels) {
			return false
		}
	}

	// If one of the accept criterias matches, then it's allowed
	for _, reject := range criteria.Accept {
		if evaluate(reject, labels) {
			return true
		}
	}

	// If the accept section is empty, return true
	if len(criteria.Accept) == 0 {
		return true
	}

	return false
}

// Lazily initializes and returns a map of name -> component
func (service *Service) getComponentsMap() map[string]*ServiceComponent {
	if service.componentsMap == nil {
		// Put all components into map
		service.componentsMap = make(map[string]*ServiceComponent)
		for _, c := range service.Components {
			service.componentsMap[c.Name] = c
		}
	}
	return service.componentsMap
}

// Makes of copy of parameter structure
func (src NestedParameterMap) makeCopy() NestedParameterMap {
	result := NestedParameterMap{}
	for k, v := range src {
		result[k] = v
	}
	return result
}

// Gets nested parameter map
func (src NestedParameterMap) getNestedMap(key string) NestedParameterMap {
	return src[key].(NestedParameterMap)
}
