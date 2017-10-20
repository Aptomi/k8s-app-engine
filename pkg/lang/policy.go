package lang

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/object"
	"strings"
	"sync"
)

// Policy describes the entire aptomi policy, consisting of multiple namespaces
type Policy struct {
	Namespace map[string]*PolicyNamespace

	once        sync.Once
	aclResolver *ACLResolver // lazily initialized value
}

// NewPolicy creates a new Policy
func NewPolicy() *Policy {
	return &Policy{
		Namespace: make(map[string]*PolicyNamespace),
	}
}

// View returns a policy view object, which allows to make all policy operations on behalf of a certain user
// Policy view object will enforce all ACLs, allowing the user to only perform actions which he is allowed to perform
// All ACL rules should be loaded and added to the policy before this method gets called
func (policy *Policy) View(user *User) *PolicyView {
	policy.once.Do(func() {
		systemNamespace := policy.Namespace[object.SystemNS]
		if systemNamespace != nil {
			policy.aclResolver = NewACLResolver(systemNamespace.ACLRules)
		} else {
			policy.aclResolver = NewACLResolver(NewGlobalRules())
		}
	})
	return NewPolicyView(policy, user)
}

// AddObject adds an object into the policy, putting it into the corresponding namespace
func (policy *Policy) AddObject(obj object.Base) {
	policyNamespace, ok := policy.Namespace[obj.GetNamespace()]
	if !ok {
		policyNamespace = NewPolicyNamespace(obj.GetNamespace())
		policy.Namespace[obj.GetNamespace()] = policyNamespace
	}
	policyNamespace.addObject(obj)
}

// GetObjectsByKind returns all objects in a policy with a given kind, across all namespaces
func (policy *Policy) GetObjectsByKind(kind string) []object.Base {
	result := []object.Base{}
	for _, policyNS := range policy.Namespace {
		result = append(result, policyNS.getObjectsByKind(kind)...)
	}
	return result
}

// GetObject looks up and returns an object from the policy, given its kind, locator ([namespace/]name), and namespace relative to which the call is being made
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
		return nil, fmt.Errorf("can't parse policy object locator: '%s'", locator)
	}

	policyNS, ok := policy.Namespace[ns]
	if !ok {
		return nil, fmt.Errorf("namespace '%s' doesn't exist, but referenced in locator '%s'", ns, locator)
	}

	return policyNS.getObject(kind, name)
}
