package language

import (
	"fmt"
	. "github.com/Aptomi/aptomi/pkg/slinga/object"
)

/*
	This file declares all the necessary structures for Slinga
*/

var PolicyNamespaceDataObject = &Info{
	Kind:        "policy",
	Constructor: func() Base { return &PolicyNamespaceData{} },
}

type PolicyNamespaceData struct {
	Metadata

	Objects map[string]Key
}

// PolicyNamespace is a global policy object with services and contexts
type PolicyNamespace struct {
	Services     map[string]*Service
	Contexts     map[string]*Context
	Clusters     map[string]*Cluster
	Rules        *GlobalRules
	Dependencies *GlobalDependencies
}

func NewPolicyNamespace() *PolicyNamespace {
	return &PolicyNamespace{
		Services:     make(map[string]*Service),
		Contexts:     make(map[string]*Context),
		Clusters:     make(map[string]*Cluster),
		Rules:        NewGlobalRules(),
		Dependencies: NewGlobalDependencies(),
	}
}

// TODO: deal with namespaces
func (policy *PolicyNamespace) AddObject(object Base) {
	switch kind := object.GetKind(); kind {
	case ServiceObject.Kind:
		policy.Services[object.GetName()] = object.(*Service)
	case ContextObject.Kind:
		policy.Contexts[object.GetName()] = object.(*Context)
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
