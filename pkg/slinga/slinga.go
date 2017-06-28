package slinga

import (
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/mattn/go-zglob"
	"sort"
	. "github.com/Frostman/aptomi/pkg/slinga/maputil"
)

/*
	This file declares all the necessary structures for Slinga
*/

// LabelOperations defines the set of label manipulations (e.g. set/remove)
type LabelOperations map[string]map[string]string

// Criteria defines a structure with criteria accept/reject syntax
type Criteria struct {
	Accept []string
	Reject []string
}

// Allocation defines within a Context for a given service
type Allocation struct {
	Name     string
	Criteria *Criteria
	Labels   *LabelOperations

	// Evaluated field (when parameters in name are substituted with real values)
	NameResolved string
}

// Context for a given service
type Context struct {
	Name        string
	Service     string
	Criteria    *Criteria
	Labels      *LabelOperations
	Allocations []*Allocation
}

// ParameterTree is a special type alias defined for freeform blocks with parameters
type ParameterTree interface{}

// Code with type and parameters, used to instantiate/update/delete component instances
type Code struct {
	Type   string
	Params ParameterTree
}

// ServiceComponent defines component within a service
type ServiceComponent struct {
	Name         string
	Service      string
	Code         *Code
	Discovery    ParameterTree
	Dependencies []string
	Labels       *LabelOperations
}

// Service defines individual service
type Service struct {
	Enabled    bool
	Name       string
	Owner      string
	Labels     *LabelOperations
	Components []*ServiceComponent

	// Lazily evaluated field (all components topologically sorted). Use via getter
	componentsOrdered []*ServiceComponent

	// Lazily evaluated field. Use via getter
	componentsMap map[string]*ServiceComponent
}

// UnmarshalYAML is a custom unmarshaller for Service, which sets Enabled to True before unmarshalling
func (s *Service) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type Alias Service
	instance := Alias{Enabled: true}
	if err := unmarshal(&instance); err != nil {
		return err
	}
	*s = Service(instance)
	return nil
}

// Cluster defines individual K8s cluster and way to access it
type Cluster struct {
	Name   string
	Type   string
	Labels map[string]string
	Metadata struct {
		KubeContext     string
		TillerNamespace string
		Namespace       string

		// store local proxy address when connection established
		tillerHost string

		// store kube external address
		kubeExternalAddress string

		// store istio svc name
		istioSvc string
	}
}

// Policy is a global policy object with services and contexts
type Policy struct {
	Services map[string]*Service
	Contexts map[string][]*Context
	Clusters map[string]*Cluster
	Rules    GlobalRules
}

func (policy *Policy) countServices() int {
	return CountElements(policy.Services)
}

func (policy *Policy) countContexts() int {
	return CountElements(policy.Contexts)
}

func (policy *Policy) countClusters() int {
	return CountElements(policy.Clusters)
}

// LoadPolicyFromDir loads policy from a directory, recursively processing all files
func LoadPolicyFromDir(baseDir string) Policy {
	s := Policy{
		Services: make(map[string]*Service),
		Contexts: make(map[string][]*Context),
		Clusters: make(map[string]*Cluster),
		Rules: NewGlobalRules(),
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
	toDisable := map[string][]string {
		"analytics_pipeline": {"hdfs", "kafka", "spark", "zookeeper"},
		"twitter_stats": {"istio"},
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

// ResetAptomiState fully resets aptomi state by deleting all files and directories from its database
// That includes all revisions of policy, resolution data, logs, etc
func ResetAptomiState() {
	baseDir := GetAptomiBaseDir()
	debug.WithFields(log.Fields{
		"baseDir": baseDir,
	}).Info("Resetting aptomi state")

	err := deleteDirectoryContents(baseDir)
	if err != nil {
		debug.WithFields(log.Fields{
			"directory": baseDir,
			"error":     err,
		}).Panic("Directory contents can't be deleted")
	}

	fmt.Println("Aptomi state is now empty. Deleted all objects")
}

// MarshalJSON marshals service component into a structure without freeform parameters, so UI doesn't fail
// See http://choly.ca/post/go-json-marshalling/
func (u *ServiceComponent) MarshalJSON() ([]byte, error) {
	type Alias ServiceComponent
	return json.Marshal(&struct {
		Code      *Code
		Discovery ParameterTree
		*Alias
	}{
		Code:      nil,
		Discovery: nil,
		Alias:     (*Alias)(u),
	})
}
