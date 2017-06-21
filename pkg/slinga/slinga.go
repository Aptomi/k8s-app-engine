package slinga

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/mattn/go-zglob"
	"sort"
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

// Code with type and parameters, used to instantiate/update/delete component instances
type Code struct {
	Type   string
	Params interface{}
}

// ServiceComponent defines component within a service
type ServiceComponent struct {
	Name         string
	Service      string
	Code         *Code
	Discovery    interface{}
	Dependencies []string
	Labels       *LabelOperations
}

// Service defines individual service
type Service struct {
	Name       string
	Labels     *LabelOperations
	Components []*ServiceComponent

	// Lazily evaluated field (all components topologically sorted). Use via getter
	componentsOrdered []*ServiceComponent

	// Lazily evaluated field. Use via getter
	componentsMap map[string]*ServiceComponent
}

// Cluster defines individual K8s cluster and way to access it
type Cluster struct {
	Name     string
	Type     string
	Labels   map[string]string
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
	return countElements(policy.Services)
}

func (policy *Policy) countContexts() int {
	return countElements(policy.Contexts)
}

func (policy *Policy) countClusters() int {
	return countElements(policy.Clusters)
}

// LoadPolicyFromDir loads policy from a directory, recursively processing all files
func LoadPolicyFromDir(baseDir string) Policy {
	s := Policy{
		Services: make(map[string]*Service),
		Contexts: make(map[string][]*Context),
		Clusters: make(map[string]*Cluster),
	}

	// read all clusters
	files, _ := zglob.Glob(GetAptomiObjectFilePatternYaml(baseDir, TypeCluster))
	sort.Strings(files)
	for _, f := range files {
		cluster := loadClusterFromFile(f)
		s.Clusters[cluster.Name] = cluster
	}

	// read all services
	files, _ = zglob.Glob(GetAptomiObjectFilePatternYaml(baseDir, TypeService))
	sort.Strings(files)
	for _, f := range files {
		service := loadServiceFromFile(f)
		s.Services[service.Name] = service
	}

	// read all contexts
	files, _ = zglob.Glob(GetAptomiObjectFilePatternYaml(baseDir, TypeContext))
	sort.Strings(files)
	for _, f := range files {
		context := loadContextFromFile(f)
		s.Contexts[context.Service] = append(s.Contexts[context.Service], context)
	}

	// read all rules
	s.Rules = LoadRulesFromDir(baseDir)

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
