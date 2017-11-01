package lang

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/object"
	"gopkg.in/go-playground/validator.v9"
)

// PolicyNamespace describes a specific namespace within Aptomi policy.
// All policy objects get placed in the appropriate maps and structs within PolicyNamespace.
type PolicyNamespace struct {
	Name         string
	Services     map[string]*Service
	Contracts    map[string]*Contract
	Clusters     map[string]*Cluster
	Rules        *GlobalRules
	ACLRules     *GlobalRules
	Dependencies *GlobalDependencies

	// Validator for policy objects
	validator *validator.Validate
}

// NewPolicyNamespace creates a new PolicyNamespace
func NewPolicyNamespace(name string, validator *validator.Validate) *PolicyNamespace {
	return &PolicyNamespace{
		Name:         name,
		Services:     make(map[string]*Service),
		Contracts:    make(map[string]*Contract),
		Clusters:     make(map[string]*Cluster),
		Rules:        NewGlobalRules(),
		ACLRules:     NewGlobalRules(),
		Dependencies: NewGlobalDependencies(),
		validator:    validator,
	}
}

func (policyNamespace *PolicyNamespace) addObject(obj object.Base) error {
	// validate object before adding it
	err := policyNamespace.validator.Struct(obj)
	if err != nil {
		return err
	}

	// add object
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
