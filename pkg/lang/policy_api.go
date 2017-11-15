package lang

type APIPolicy struct {
	Namespace map[string]*PolicyNamespace
}

type APIPolicyNamespace struct {
	Services     map[string]*Service
	Contracts    map[string]*Contract
	Clusters     map[string]*Cluster
	Rules        map[string]*Rule
	ACLRules     map[string]*Rule
	Dependencies map[string]*Dependency
}

func (view *PolicyView) APIPolicy() *APIPolicy {
	// if we're changing data in any map, we should copy map as well
	// don't change existing object, make copy of them

	return &APIPolicy{}
}
