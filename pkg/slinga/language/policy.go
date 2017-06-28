package language

import (
	. "github.com/Frostman/aptomi/pkg/slinga/fileio"
	. "github.com/Frostman/aptomi/pkg/slinga/util"
	"github.com/mattn/go-zglob"
	"sort"
)

/*
	This file declares all the necessary structures for Slinga
*/

// Policy is a global policy object with services and contexts
type Policy struct {
	Services map[string]*Service
	Contexts map[string][]*Context
	Clusters map[string]*Cluster
	Rules    GlobalRules
}

// CountServices returns number of services in the policy
func (policy *Policy) CountServices() int {
	return CountElements(policy.Services)
}

// CountContexts returns number of contexts in the policy
func (policy *Policy) CountContexts() int {
	return CountElements(policy.Contexts)
}

// CountClusters returns number of clusters in the policy
func (policy *Policy) CountClusters() int {
	return CountElements(policy.Clusters)
}

// LoadPolicyFromDir loads policy from a directory, recursively processing all files
func LoadPolicyFromDir(baseDir string) Policy {
	s := Policy{
		Services: make(map[string]*Service),
		Contexts: make(map[string][]*Context),
		Clusters: make(map[string]*Cluster),
		Rules:    NewGlobalRules(),
	}

	// read all clusters
	files, _ := zglob.Glob(GetAptomiObjectFilePatternYaml(baseDir, TypeCluster))
	sort.Strings(files)
	for _, f := range files {
		cluster := loadClusterFromFile(f)
		s.Clusters[cluster.Name] = cluster
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
		s.Services[service.Name] = service
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
			if !service.Enabled || disabled[service.Name] {
				if !disabled[service.Name] {
					changed = true
				}
				disabled[service.Name] = true
				for _, componentName := range toDisable[service.Name] {
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
		if s.Services[context.Service] != nil {
			s.Contexts[context.Service] = append(s.Contexts[context.Service], context)
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

