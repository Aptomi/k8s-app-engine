package lang

import (
	"fmt"

	"github.com/Aptomi/aptomi/pkg/runtime"
	"gopkg.in/yaml.v2"
)

// PolicyNamespace describes a specific namespace within Aptomi policy.
// All policy objects get placed in the appropriate maps and structs within PolicyNamespace.
type PolicyNamespace struct {
	Name     string              `validate:"identifier"`
	Bundles  map[string]*Bundle  `validate:"dive"`
	Services map[string]*Service `validate:"dive"`
	Clusters map[string]*Cluster `validate:"dive"`
	Rules    map[string]*Rule    `validate:"dive"`
	ACLRules map[string]*ACLRule `validate:"dive"`
	Claims   map[string]*Claim   `validate:"dive"`
}

// NewPolicyNamespace creates a new PolicyNamespace
func NewPolicyNamespace(name string) *PolicyNamespace {
	return &PolicyNamespace{
		Name:     name,
		Bundles:  make(map[string]*Bundle),
		Services: make(map[string]*Service),
		Clusters: make(map[string]*Cluster),
		Rules:    make(map[string]*Rule),
		ACLRules: make(map[string]*ACLRule),
		Claims:   make(map[string]*Claim),
	}
}

func (policyNamespace *PolicyNamespace) addObject(obj Base) error {
	switch kind := obj.GetKind(); kind {
	case TypeBundle.Kind:
		policyNamespace.Bundles[obj.GetName()] = obj.(*Bundle) // nolint: errcheck
	case TypeService.Kind:
		policyNamespace.Services[obj.GetName()] = obj.(*Service) // nolint: errcheck
	case TypeCluster.Kind:
		// cluster is a special object, which we don't allow to update certain parts of (e.g. type and config)
		clusterUpdated := obj.(*Cluster) // nolint: errcheck
		clusterExisting, present := policyNamespace.Clusters[obj.GetName()]
		if present {
			// we can't really use reflect.DeepEqual here, because it treats nil and empty maps differently
			configExisting, err := yaml.Marshal(clusterExisting.Config)
			if err != nil {
				return err
			}
			configUpdated, err := yaml.Marshal(clusterUpdated.Config)
			if err != nil {
				return err
			}

			// we can't change type or config
			if clusterUpdated.Type != clusterExisting.Type || string(configUpdated) != string(configExisting) {
				return fmt.Errorf("modification of cluster type or config is not allowed: %s needs to be deleted first", obj.GetName())
			}
		}
		policyNamespace.Clusters[obj.GetName()] = obj.(*Cluster) // nolint: errcheck
	case TypeRule.Kind:
		policyNamespace.Rules[obj.GetName()] = obj.(*Rule) // nolint: errcheck
	case TypeACLRule.Kind:
		policyNamespace.ACLRules[obj.GetName()] = obj.(*ACLRule) // nolint: errcheck
	case TypeClaim.Kind:
		policyNamespace.Claims[obj.GetName()] = obj.(*Claim) // nolint: errcheck
	default:
		return fmt.Errorf("not supported by PolicyNamespace.addObject(): unknown kind %s", kind)
	}
	return nil
}

func (policyNamespace *PolicyNamespace) removeObject(obj Base) bool {
	switch kind := obj.GetKind(); kind {
	case TypeBundle.Kind:
		if _, exist := policyNamespace.Bundles[obj.GetName()]; exist {
			delete(policyNamespace.Bundles, obj.GetName())
			return true
		}
	case TypeService.Kind:
		if _, exist := policyNamespace.Services[obj.GetName()]; exist {
			delete(policyNamespace.Services, obj.GetName())
			return true
		}
	case TypeCluster.Kind:
		if _, exist := policyNamespace.Clusters[obj.GetName()]; exist {
			delete(policyNamespace.Clusters, obj.GetName())
			return true
		}
	case TypeRule.Kind:
		if _, exist := policyNamespace.Rules[obj.GetName()]; exist {
			delete(policyNamespace.Rules, obj.GetName())
			return true
		}
	case TypeACLRule.Kind:
		if _, exist := policyNamespace.ACLRules[obj.GetName()]; exist {
			delete(policyNamespace.ACLRules, obj.GetName())
			return true
		}
	case TypeClaim.Kind:
		if _, exist := policyNamespace.Claims[obj.GetName()]; exist {
			delete(policyNamespace.Claims, obj.GetName())
			return true
		}
	}

	return false
}

func (policyNamespace *PolicyNamespace) getObjectsByKind(kind string) []Base {
	var result []Base
	switch kind {
	case TypeBundle.Kind:
		for _, bundle := range policyNamespace.Bundles {
			result = append(result, bundle)
		}
	case TypeService.Kind:
		for _, service := range policyNamespace.Services {
			result = append(result, service)
		}
	case TypeCluster.Kind:
		for _, cluster := range policyNamespace.Clusters {
			result = append(result, cluster)
		}
	case TypeRule.Kind:
		for _, rule := range policyNamespace.Rules {
			result = append(result, rule)
		}
	case TypeACLRule.Kind:
		for _, rule := range policyNamespace.ACLRules {
			result = append(result, rule)
		}
	case TypeClaim.Kind:
		for _, claim := range policyNamespace.Claims {
			result = append(result, claim)
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
	case TypeBundle.Kind:
		if result, ok = policyNamespace.Bundles[name]; !ok {
			return nil, nil
		}
	case TypeService.Kind:
		if result, ok = policyNamespace.Services[name]; !ok {
			return nil, nil
		}
	case TypeCluster.Kind:
		if result, ok = policyNamespace.Clusters[name]; !ok {
			return nil, nil
		}
	case TypeRule.Kind:
		if result, ok = policyNamespace.Rules[name]; !ok {
			return nil, nil
		}
	case TypeACLRule.Kind:
		if result, ok = policyNamespace.ACLRules[name]; !ok {
			return nil, nil
		}
	case TypeClaim.Kind:
		if result, ok = policyNamespace.Claims[name]; !ok {
			return nil, nil
		}
	default:
		return nil, fmt.Errorf("not supported by PolicyNamespace.getObject(): unknown kind %s, %s", kind, name)
	}
	return result, nil
}
