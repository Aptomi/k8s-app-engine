package language

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/slinga/object"
	"strings"
)

// Policy describes the entire policy with all namespaces
type Policy struct {
	Namespace map[string]*PolicyNamespace
}

func NewPolicy() *Policy {
	return &Policy{
		Namespace: make(map[string]*PolicyNamespace),
	}
}

func (policy *Policy) AddObject(object object.Base) {
	policyNamespace, ok := policy.Namespace[object.GetNamespace()]
	if !ok {
		policyNamespace = NewPolicyNamespace(object.GetNamespace())
		policy.Namespace[object.GetNamespace()] = policyNamespace
	}
	policyNamespace.addObject(object)
}

func (policy *Policy) GetObjectsByKind(kind string) []object.Base {
	result := []object.Base{}
	for _, policyNS := range policy.Namespace {
		result = append(result, policyNS.getObjectsByKind(kind)...)
	}
	return result
}

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
		return nil, fmt.Errorf("Can't parse reference to a policy object: ", locator)
	}

	policyNS, ok := policy.Namespace[ns]
	if !ok {
		return nil, fmt.Errorf("Namespace %s doesn't exist, but referenced from ref %s", ns, locator)
	}

	return policyNS.getObject(kind, name), nil
}
