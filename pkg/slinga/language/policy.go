package language

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/slinga/object"
)

/*
	This file declares all the necessary structures for Slinga
*/

// Policy describes the entire policy with all namespaces included
type Policy = PolicyNamespace

func NewPolicy() *Policy {
	return NewPolicyNamespace()
}

// PolicyNamespace describes a specific namespace in a policy (services, contracts, clusters, rules and dependencies, etc)
type PolicyNamespace struct {
	Services     map[string]*Service
	Contracts    map[string]*Contract
	Clusters     map[string]*Cluster
	Rules        *GlobalRules
	Dependencies *GlobalDependencies
}

func NewPolicyNamespace() *PolicyNamespace {
	return &PolicyNamespace{
		Services:     make(map[string]*Service),
		Contracts:    make(map[string]*Contract),
		Clusters:     make(map[string]*Cluster),
		Rules:        NewGlobalRules(),
		Dependencies: NewGlobalDependencies(),
	}
}

// TODO: deal with namespaces
func (policy *PolicyNamespace) AddObject(object object.Base) {
	switch kind := object.GetKind(); kind {
	case ServiceObject.Kind:
		policy.Services[object.GetName()] = object.(*Service)
	case ContractObject.Kind:
		policy.Contracts[object.GetName()] = object.(*Contract)
	case ClusterObject.Kind:
		policy.Clusters[object.GetName()] = object.(*Cluster)
	case RuleObject.Kind:
		policy.Rules.addRule(object.(*Rule))
	case DependencyObject.Kind:
		policy.Dependencies.AddDependency(object.(*Dependency))
	default:
		panic(fmt.Sprintf("Can't add object to policy: %v", object))
	}
}

func (policy *PolicyNamespace) GetClusterByLabels(labels *LabelSet) (*Cluster, error) {
	var cluster *Cluster
	if clusterName, ok := labels.Labels["cluster"]; ok {
		if cluster, ok = policy.Clusters[clusterName]; !ok {
			return nil, fmt.Errorf("Cluster '%s' is not defined in policy", clusterName)
		}
	}
	return cluster, nil
}
