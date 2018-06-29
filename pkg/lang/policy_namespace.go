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
	case BundleType.Kind:
		policyNamespace.Bundles[obj.GetName()] = obj.(*Bundle) // nolint: errcheck
	case ServiceObject.Kind:
		policyNamespace.Services[obj.GetName()] = obj.(*Service) // nolint: errcheck
	case ClusterObject.Kind:
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
	case RuleObject.Kind:
		policyNamespace.Rules[obj.GetName()] = obj.(*Rule) // nolint: errcheck
	case ACLRuleObject.Kind:
		policyNamespace.ACLRules[obj.GetName()] = obj.(*ACLRule) // nolint: errcheck
	case ClaimType.Kind:
		policyNamespace.Claims[obj.GetName()] = obj.(*Claim) // nolint: errcheck
	default:
		return fmt.Errorf("not supported by PolicyNamespace.addObject(): unknown kind %s", kind)
	}
	return nil
}

func (policyNamespace *PolicyNamespace) removeObject(obj Base) bool {
	switch kind := obj.GetKind(); kind {
	case BundleType.Kind:
		if _, exist := policyNamespace.Bundles[obj.GetName()]; exist {
			delete(policyNamespace.Bundles, obj.GetName())
			return true
		}
	case ServiceObject.Kind:
		if _, exist := policyNamespace.Services[obj.GetName()]; exist {
			delete(policyNamespace.Services, obj.GetName())
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
	case ClaimType.Kind:
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
	case BundleType.Kind:
		for _, bundle := range policyNamespace.Bundles {
			result = append(result, bundle)
		}
	case ServiceObject.Kind:
		for _, service := range policyNamespace.Services {
			result = append(result, service)
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
	case ClaimType.Kind:
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
	case BundleType.Kind:
		if result, ok = policyNamespace.Bundles[name]; !ok {
			return nil, nil
		}
	case ServiceObject.Kind:
		if result, ok = policyNamespace.Services[name]; !ok {
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
	case ClaimType.Kind:
		if result, ok = policyNamespace.Claims[name]; !ok {
			return nil, nil
		}
	default:
		return nil, fmt.Errorf("not supported by PolicyNamespace.getObject(): unknown kind %s, %s", kind, name)
	}
	return result, nil
}
