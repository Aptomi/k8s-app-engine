package lang

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/runtime"
)

// PolicyNamespace describes a specific namespace within Aptomi policy.
// All policy objects get placed in the appropriate maps and structs within PolicyNamespace.
type PolicyNamespace struct {
	Name         string                 `validate:"identifier"`
	Services     map[string]*Service    `validate:"dive"`
	Contracts    map[string]*Contract   `validate:"dive"`
	Clusters     map[string]*Cluster    `validate:"dive"`
	Rules        map[string]*Rule       `validate:"dive"`
	ACLRules     map[string]*Rule       `validate:"dive"`
	Dependencies map[string]*Dependency `validate:"dive"`
}

// NewPolicyNamespace creates a new PolicyNamespace
func NewPolicyNamespace(name string) *PolicyNamespace {
	return &PolicyNamespace{
		Name:         name,
		Services:     make(map[string]*Service),
		Contracts:    make(map[string]*Contract),
		Clusters:     make(map[string]*Cluster),
		Rules:        make(map[string]*Rule),
		ACLRules:     make(map[string]*Rule),
		Dependencies: make(map[string]*Dependency),
	}
}

func (policyNamespace *PolicyNamespace) addObject(obj Base) error {
	switch kind := obj.GetKind(); kind {
	case ServiceObject.Kind:
		policyNamespace.Services[obj.GetName()] = obj.(*Service)
	case ContractObject.Kind:
		policyNamespace.Contracts[obj.GetName()] = obj.(*Contract)
	case ClusterObject.Kind:
		policyNamespace.Clusters[obj.GetName()] = obj.(*Cluster)
	case RuleObject.Kind:
		policyNamespace.Rules[obj.GetName()] = obj.(*Rule)
	case ACLRuleObject.Kind:
		policyNamespace.ACLRules[obj.GetName()] = obj.(*Rule)
	case DependencyObject.Kind:
		policyNamespace.Dependencies[obj.GetName()] = obj.(*Dependency)
	default:
		return fmt.Errorf("not supported by PolicyNamespace.addObject(): unknown kind %s", kind)
	}
	return nil
}

func (policyNamespace *PolicyNamespace) removeObject(obj Base) bool {
	switch kind := obj.GetKind(); kind {
	case ServiceObject.Kind:
		if _, exist := policyNamespace.Services[obj.GetName()]; exist {
			delete(policyNamespace.Services, obj.GetName())
			return true
		}
	case ContractObject.Kind:
		if _, exist := policyNamespace.Contracts[obj.GetName()]; exist {
			delete(policyNamespace.Contracts, obj.GetName())
			return true
		}
	case ClusterObject.Kind:
		if _, exist := policyNamespace.Clusters[obj.GetName()]; exist {
			delete(policyNamespace.Clusters, obj.GetName())
			return true
		}
	case RuleObject.Kind:
		if _, exist := policyNamespace.Rules[obj.GetName()]; exist {
			delete(policyNamespace.Rules, obj.GetName())
			return true
		}
	case ACLRuleObject.Kind:
		if _, exist := policyNamespace.ACLRules[obj.GetName()]; exist {
			delete(policyNamespace.ACLRules, obj.GetName())
			return true
		}
	case DependencyObject.Kind:
		if _, exist := policyNamespace.Dependencies[obj.GetName()]; exist {
			delete(policyNamespace.Dependencies, obj.GetName())
			return true
		}
	}

	return false
}

func (policyNamespace *PolicyNamespace) getObjectsByKind(kind string) []Base {
	var result []Base
	switch kind {
	case ServiceObject.Kind:
		for _, service := range policyNamespace.Services {
			result = append(result, service)
		}
	case ContractObject.Kind:
		for _, contract := range policyNamespace.Contracts {
			result = append(result, contract)
		}
	case ClusterObject.Kind:
		for _, cluster := range policyNamespace.Clusters {
			result = append(result, cluster)
		}
	case RuleObject.Kind:
		for _, rule := range policyNamespace.Rules {
			result = append(result, rule)
		}
	case ACLRuleObject.Kind:
		for _, rule := range policyNamespace.ACLRules {
			result = append(result, rule)
		}
	case DependencyObject.Kind:
		for _, dependency := range policyNamespace.Dependencies {
			result = append(result, dependency)
		}
	default:
		panic(fmt.Sprintf("not supported by PolicyNamespace.getObjectsByKind(): unknown kind %s", kind))
	}
	return result
}

func (policyNamespace *PolicyNamespace) getObject(kind string, name string) (runtime.Object, error) {
	var ok bool
	var result Base
	switch kind {
	case ServiceObject.Kind:
		if result, ok = policyNamespace.Services[name]; !ok {
			return nil, nil
		}
	case ContractObject.Kind:
		if result, ok = policyNamespace.Contracts[name]; !ok {
			return nil, nil
		}
	case ClusterObject.Kind:
		if result, ok = policyNamespace.Clusters[name]; !ok {
			return nil, nil
		}
	case RuleObject.Kind:
		if result, ok = policyNamespace.Rules[name]; !ok {
			return nil, nil
		}
	case ACLRuleObject.Kind:
		if result, ok = policyNamespace.ACLRules[name]; !ok {
			return nil, nil
		}
	case DependencyObject.Kind:
		if result, ok = policyNamespace.Dependencies[name]; !ok {
			return nil, nil
		}
	default:
		return nil, fmt.Errorf("not supported by PolicyNamespace.getObject(): unknown kind %s, %s", kind, name)
	}
	return result, nil
}
