package language

import (
	. "github.com/Aptomi/aptomi/pkg/slinga/db"
	"github.com/mattn/go-zglob"
	"sort"
	"fmt"
	"reflect"
)

/*
	This file declares all the necessary structures for Slinga
*/

// Policy is a global policy object with services and contexts
type Policy struct {
	Services     map[string]*Service
	Contexts     map[string]*Context
	Clusters     map[string]*Cluster
	Rules        *GlobalRules
	Dependencies *GlobalDependencies
}

func NewPolicy() *Policy {
	return &Policy{
		Services: make(map[string]*Service),
		Contexts: make(map[string]*Context),
		Clusters: make(map[string]*Cluster),
		Rules:    NewGlobalRules(),
		Dependencies: NewGlobalDependencies(),
	}
}

// TODO: deal with namespaces
func (policy *Policy) addObject(object SlingaObjectInterface) {
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

// LoadPolicyFromDir loads policy from a directory, recursively processing all files
func LoadPolicyFromDir(baseDir string) Policy {
	s := Policy{
		Services: make(map[string]*Service),
		Contexts: make(map[string]*Context),
		Clusters: make(map[string]*Cluster),
		Rules:    NewGlobalRules(),
	}

	// read all clusters
	files, _ := zglob.Glob(GetAptomiObjectFilePatternYaml(baseDir, TypeCluster))
	sort.Strings(files)
	for _, f := range files {
		cluster := loadClusterFromFile(f)
		s.Clusters[cluster.GetName()] = cluster
	}

	// read all services
	/*
		files, _ = zglob.Glob(GetAptomiObjectFilePatternYaml(baseDir, TypeService))
		sort.Strings(files)
		for _, f := range files {
			service := loadServiceFromFile(f)
			if service.Enabled {
				s.Services[service.Name] = service
			}
		}
	*/

	// TODO: remove later - it's a temporary hack for the demo, so we can enable/disable groups of services
	files, _ = zglob.Glob(GetAptomiObjectFilePatternYaml(baseDir, TypeService))
	sort.Strings(files)
	for _, f := range files {
		service := loadServiceFromFile(f)
		s.Services[service.GetName()] = service
	}
	toDisable := map[string][]string{
		"analytics_pipeline": {"hdfs", "kafka", "spark", "zookeeper"},
		"twitter_stats":      {"istio"},
	}
	disabled := make(map[string]bool)
	changed := true
	for changed {
		changed = false
		for _, service := range s.Services {
			if !service.Enabled || disabled[service.GetName()] {
				if !disabled[service.GetName()] {
					changed = true
				}
				disabled[service.GetName()] = true
				for _, componentName := range toDisable[service.GetName()] {
					if !disabled[componentName] {
						changed = true
					}
					disabled[componentName] = true
				}
			}
		}
	}
	for serviceName := range disabled {
		delete(s.Services, serviceName)
	}

	// read all contexts
	files, _ = zglob.Glob(GetAptomiObjectFilePatternYaml(baseDir, TypeContext))
	sort.Strings(files)
	for _, f := range files {
		context := loadContextFromFile(f)
		if s.Services[context.GetName()] != nil {
			s.Contexts[context.GetName()] = context
		}
	}

	// read all rules
	files, _ = zglob.Glob(GetAptomiObjectFilePatternYaml(baseDir, TypeRules))
	sort.Strings(files)
	for _, f := range files {
		rules := loadRulesFromFile(f)
		for _, rule := range rules {
			if rule.Enabled {
				s.Rules.addRule(rule)
			}
		}
	}

	return s
}
