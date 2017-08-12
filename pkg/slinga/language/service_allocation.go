package language

// Allocation defines within a Context for a given service
type Allocation struct {
	Name         string
	ChangeLabels *LabelOperations `yaml:"change-labels"`
}

// ResolveName resolves name for an allocation
func (allocation *Allocation) ResolveName(user *User, labels LabelSet) (string, error) {
	return evaluateTemplate(allocation.Name, user, labels)
}
