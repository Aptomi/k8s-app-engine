package language

// Allocation defines within a Context for a given service
type Allocation struct {
	Name     string
	Labels   *LabelOperations

	// Evaluated field (when parameters in name are substituted with real values)
	NameResolved string
}

// ResolveName resolves name for an allocation
func (allocation *Allocation) ResolveName(user *User, labels LabelSet) error {
	result, err := evaluateTemplate(allocation.Name, user, labels)
	allocation.NameResolved = result
	return err
}
