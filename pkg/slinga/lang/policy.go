package lang

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/slinga/object"
	"strings"
)

// Policy describes the entire aptomi policy, consisting of multiple namespaces
type Policy struct {
	Namespace map[string]*PolicyNamespace
}

// NewPolicy creates a new Policy
func NewPolicy() *Policy {
	return &Policy{
		Namespace: make(map[string]*PolicyNamespace),
	}
}

// AddObject adds an object into the policy, putting it into the corresponding namespace
func (policy *Policy) AddObject(object object.Base) {
	policyNamespace, ok := policy.Namespace[object.GetNamespace()]
	if !ok {
		policyNamespace = NewPolicyNamespace(object.GetNamespace())
		policy.Namespace[object.GetNamespace()] = policyNamespace
	}
	policyNamespace.addObject(object)
}

// GetObjectsByKind returns all objects in a policy with a given kind, across all namespaces
func (policy *Policy) GetObjectsByKind(kind string) []object.Base {
	result := []object.Base{}
	for _, policyNS := range policy.Namespace {
		result = append(result, policyNS.getObjectsByKind(kind)...)
	}
	return result
}

// GetObject looks up and returns an objects from the policy, given its kind, locator ([namespace/]name), and namespace relative to which the call is being made
func (policy *Policy) GetObject(kind string, locator string, currentNs string) (object.Base, error) {
	// parse locator: [namespace/]name. we might add [domain/] in the future
	parts := strings.Split(locator, "/")
	var ns, name string
	if len(parts) == 1 {
		ns = currentNs
		name = parts[0]
	} else if len(parts) == 2 {
		ns = parts[0]
		name = parts[1]
	} else {
		return nil, fmt.Errorf("Can't parse policy object locator: '%s'", locator)
	}

	policyNS, ok := policy.Namespace[ns]
	if !ok {
		return nil, fmt.Errorf("Namespace '%s' doesn't exist, but referenced in locator '%s'", ns, locator)
	}

	return policyNS.getObject(kind, name)
}
