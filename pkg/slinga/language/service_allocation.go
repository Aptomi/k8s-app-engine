package language

// Allocation defines within a Context for a given service
type Allocation struct {
	Name     string
	Criteria *Criteria
	Labels   *LabelOperations

	// Evaluated field (when parameters in name are substituted with real values)
	NameResolved string
}

// Matches checks if allocation criteria is satisfied
func (allocation *Allocation) Matches(labels LabelSet) bool {
	return allocation.Criteria == nil || allocation.Criteria.allows(labels)
}

// ResolveName resolves name for an allocation
func (allocation *Allocation) ResolveName(user *User, labels LabelSet) error {
	result, err := evaluateTemplate(allocation.Name, user, labels)
	allocation.NameResolved = result
	return err
}
