package lang

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/object"
)

// PolicyNamespace describes a specific namespace in aptomi policy (services, contracts, clusters, rules, dependencies, etc)
type PolicyNamespace struct {
	Name         string
	Services     map[string]*Service
	Contracts    map[string]*Contract
	Clusters     map[string]*Cluster
	Rules        *GlobalRules
	Dependencies *GlobalDependencies
}

// NewPolicyNamespace creates a new PolicyNamespace
func NewPolicyNamespace(name string) *PolicyNamespace {
	return &PolicyNamespace{
		Name:         name,
		Services:     make(map[string]*Service),
		Contracts:    make(map[string]*Contract),
		Clusters:     make(map[string]*Cluster),
		Rules:        NewGlobalRules(),
		Dependencies: NewGlobalDependencies(),
	}
}

func (policyNamespace *PolicyNamespace) addObject(obj object.Base) {
	switch kind := obj.GetKind(); kind {
	case ServiceObject.Kind:
		policyNamespace.Services[obj.GetName()] = obj.(*Service)
	case ContractObject.Kind:
		policyNamespace.Contracts[obj.GetName()] = obj.(*Contract)
	case ClusterObject.Kind:
		if obj.GetNamespace() != object.SystemNS {
			panic(fmt.Sprintf("Adding cluster '%s' into a non-system namespace '%s'", obj.GetName(), obj.GetNamespace()))
		}
		policyNamespace.Clusters[obj.GetName()] = obj.(*Cluster)
	case RuleObject.Kind:
		policyNamespace.Rules.addRule(obj.(*Rule))
	case DependencyObject.Kind:
		policyNamespace.Dependencies.AddDependency(obj.(*Dependency))
	default:
		panic(fmt.Sprintf("Can't add object to policy namespace: %v", obj))
	}
}

func (policyNamespace *PolicyNamespace) getObjectsByKind(kind string) []object.Base {
	result := []object.Base{}
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
	case DependencyObject.Kind:
		for _, dependencyList := range policyNamespace.Dependencies.DependenciesByContract {
			for _, dependency := range dependencyList {
				result = append(result, dependency)
			}
		}
	default:
		panic(fmt.Sprintf("Can't get objects by kind: %s", kind))
	}
	return result
}

func (policyNamespace *PolicyNamespace) getObject(kind string, name string) (object.Base, error) {
	var ok bool
	var result object.Base
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
		return nil, fmt.Errorf("Rule not supported by PolicyNamespace.getObject(): %s, %s", kind, name)
	case DependencyObject.Kind:
		return nil, fmt.Errorf("Dependency not supported by PolicyNamespace.getObject(): %s, %s", kind, name)
	default:
		return nil, fmt.Errorf("Unknown object kind in PolicyNamespace.getObject(): %s, %s", kind, name)
	}
	return result, nil
}
