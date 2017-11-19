package lang

import (
	"fmt"
)

// PolicyNamespace describes a specific namespace within Aptomi policy.
// All policy objects get placed in the appropriate maps and structs within PolicyNamespace.
type PolicyNamespace struct {
	Name         string               `validate:"identifier"`
	Services     map[string]*Service  `validate:"dive"`
	Contracts    map[string]*Contract `validate:"dive"`
	Clusters     map[string]*Cluster  `validate:"dive"`
	Rules        *GlobalRules         `validate:"required"`
	ACLRules     *GlobalRules         `validate:"required"`
	Dependencies *GlobalDependencies  `validate:"required"`
}

// NewPolicyNamespace creates a new PolicyNamespace
func NewPolicyNamespace(name string) *PolicyNamespace {
	return &PolicyNamespace{
		Name:         name,
		Services:     make(map[string]*Service),
		Contracts:    make(map[string]*Contract),
		Clusters:     make(map[string]*Cluster),
		Rules:        NewGlobalRules(),
		ACLRules:     NewGlobalRules(),
		Dependencies: NewGlobalDependencies(),
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
		policyNamespace.Rules.addRule(obj.(*Rule))
	case ACLRuleObject.Kind:
		policyNamespace.ACLRules.addRule(obj.(*Rule))
	case DependencyObject.Kind:
		policyNamespace.Dependencies.addDependency(obj.(*Dependency))
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
		return policyNamespace.Rules.removeRule(obj.(*Rule))
	case ACLRuleObject.Kind:
		return policyNamespace.ACLRules.removeRule(obj.(*Rule))
	case DependencyObject.Kind:
		return policyNamespace.Dependencies.removeDependency(obj.(*Dependency))
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
		for _, rule := range policyNamespace.Rules.Rules {
			result = append(result, rule)
		}
	case ACLRuleObject.Kind:
		for _, rule := range policyNamespace.ACLRules.Rules {
			result = append(result, rule)
		}
	case DependencyObject.Kind:
		for _, dependencyList := range policyNamespace.Dependencies.DependenciesByContract {
			for _, dependency := range dependencyList {
				result = append(result, dependency)
			}
		}
	default:
		panic(fmt.Sprintf("not supported by PolicyNamespace.getObjectsByKind(): unknown kind %s", kind))
	}
	return result
}

func (policyNamespace *PolicyNamespace) getObject(kind string, name string) (Base, error) {
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
		if result, ok = policyNamespace.Rules.RuleMap[name]; !ok {
			return nil, nil
		}
	case ACLRuleObject.Kind:
		if result, ok = policyNamespace.ACLRules.RuleMap[name]; !ok {
			return nil, nil
		}
	case DependencyObject.Kind:
		if result, ok = policyNamespace.Dependencies.DependencyMap[name]; !ok {
			return nil, nil
		}
	default:
		return nil, fmt.Errorf("not supported by PolicyNamespace.getObject(): unknown kind %s, %s", kind, name)
	}
	return result, nil
}
