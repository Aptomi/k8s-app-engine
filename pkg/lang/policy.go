package lang

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/runtime"
	"strings"
	"sync"
)

// Policy describes the entire Aptomi policy.
//
// At the highest level, policy consists of namespaces. Namespaces provide isolation for policy objects and access to
// namespaces can be controlled via ACL rules. Thus, different users can have different access rights to different parts
// of Aptomi policy. Namespaces are useful in environments with many users, multiple teams and projects.
//
// Objects get stored in their corresponding namespaces. Names of objects must be unique within a namespace and a given
// object kind.
//
// Once policy is defined, it can be passed to the engine for policy resolution. Policy resolution translates a given
// policy (intent) into actual state (what services/components need to created/updated/deleted, how and where) and the
// corresponding set of actions.
type Policy struct {
	// Namespace is a map from namespace name into a PolicyNamespace
	Namespace map[string]*PolicyNamespace `validate:"dive"`

	// Access control rules for different policy namespaces
	once        sync.Once
	aclResolver *ACLResolver // lazily initialized value
}

// NewPolicy creates a new Policy
func NewPolicy() *Policy {
	return &Policy{Namespace: make(map[string]*PolicyNamespace)}
}

// View returns a policy view object, which allows to make all policy operations on behalf of a certain user
// Policy view object will enforce all ACLs, allowing the user to only perform actions which he is allowed to perform
// All ACL rules should be loaded and added to the policy before this method gets called
func (policy *Policy) View(user *User) *PolicyView {
	policy.once.Do(func() {
		systemNamespace := policy.Namespace[runtime.SystemNS]
		if systemNamespace != nil {
			policy.aclResolver = NewACLResolver(systemNamespace.ACLRules)
		} else {
			policy.aclResolver = NewACLResolver(make(map[string]*Rule))
		}
	})
	return NewPolicyView(policy, user)
}

// AddObject adds a given object into the policy. When you add objects to the policy, they get added to the corresponding
// Namespace. If error occurs (e.g. object has an unknown kind) then the error will be returned
func (policy *Policy) AddObject(obj Base) error {
	policyNamespace, ok := policy.Namespace[obj.GetNamespace()]
	if !ok {
		policyNamespace = NewPolicyNamespace(obj.GetNamespace())
		policy.Namespace[obj.GetNamespace()] = policyNamespace
	}
	return policyNamespace.addObject(obj)
}

// RemoveObject removes a given object from the policy. Returns true if removed and false if nothing got removed.
func (policy *Policy) RemoveObject(obj Base) bool {
	policyNamespace, ok := policy.Namespace[obj.GetNamespace()]
	if !ok {
		return false
	}

	return policyNamespace.removeObject(obj)
}

// GetObjectsByKind returns all objects in a policy with a given kind, across all namespaces
func (policy *Policy) GetObjectsByKind(kind string) []Base {
	result := []Base{}
	for _, policyNS := range policy.Namespace {
		result = append(result, policyNS.getObjectsByKind(kind)...)
	}
	return result
}

// GetObject looks up and returns an object from the policy, given its kind, locator ([namespace/]name), and current
// namespace relative to which the call is being made. It may return nil and no error, if an object hasn't been found in the policy.
// TODO: we need to fix semantics of this method, so that it either returns a non-nil object or an error
func (policy *Policy) GetObject(kind string, locator string, currentNs string) (runtime.Object, error) {
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
		return nil, fmt.Errorf("can't parse policy object locator for kind '%s': '%s' (parts = %d)", kind, locator, len(parts))
	}

	policyNS, ok := policy.Namespace[ns]
	if !ok {
		return nil, fmt.Errorf("namespace '%s' has no objects, but trying to look up '%s': '%s'", ns, kind, locator)
	}

	return policyNS.getObject(kind, name)
}

// Validate performs validation of the entire policy, making sure that all of its objects are well-formed.
// It also checks that all cross-object references are valid. If policy is malformed, then a list of errors is returned.
// Otherwise, if policy is correctly formed, then nil is returned.
// The resulting error can be caster to (validator.ValidationErrors) and iterated over, to get the full list of errors.
func (policy *Policy) Validate() error {
	return NewPolicyValidator(policy).Validate()
}
