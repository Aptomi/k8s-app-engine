package language

import (
	"fmt"
	"reflect"
)

/*
	This file declares all the necessary structures for Slinga
*/

// PolicyNamespace is a global policy object with services and contexts
type PolicyNamespace struct {
	Services     map[string]*Service
	Contexts     map[string]*Context
	Clusters     map[string]*Cluster
	Rules        *GlobalRules
	Dependencies *GlobalDependencies
}

func NewPolicy() *PolicyNamespace {
	return &PolicyNamespace{
		Services:     make(map[string]*Service),
		Contexts:     make(map[string]*Context),
		Clusters:     make(map[string]*Cluster),
		Rules:        NewGlobalRules(),
		Dependencies: NewGlobalDependencies(),
	}
}

// TODO: deal with namespaces
func (policy *PolicyNamespace) addObject(object SlingaObjectInterface) {
	if object.GetObjectType() == TypePolicy {
		p := reflect.ValueOf(object).Interface()

		switch v := p.(type) {
		case *Service:
			policy.Services[v.GetName()] = v
		case *Context:
			policy.Contexts[v.GetName()] = v
		case *Cluster:
			policy.Clusters[v.GetName()] = v
		case *Rule:
			policy.Rules.addRule(v)
		case *Dependency:
			policy.Dependencies.AddDependency(v)
		default:
			panic(fmt.Sprintf("Can't add object to policy: %v", object))
		}
	}
}
