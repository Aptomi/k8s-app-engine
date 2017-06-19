package slinga

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/mattn/go-zglob"
	"gopkg.in/yaml.v2"
	"io/ioutil"
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
	Type     string
	Metadata map[string]string
	Params   interface{}
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

	// Lazily evaluated field (not serialized). Use via getter
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

// Loads service from YAML file
func loadServiceFromFile(fileName string) *Service {
	debug.WithFields(log.Fields{
		"file": fileName,
	}).Info("Loading service")

	dat, e := ioutil.ReadFile(fileName)
	if e != nil {
		debug.WithFields(log.Fields{
			"file":  fileName,
			"error": e,
		}).Fatal("Unable to read file")
	}
	t := Service{}
	e = yaml.Unmarshal([]byte(dat), &t)
	if e != nil {
		debug.WithFields(log.Fields{
			"file":  fileName,
			"error": e,
		}).Fatal("Unable to unmarshal service")
	}
	return &t
}

// Loads context from YAML file
func loadContextFromFile(fileName string) *Context {
	debug.WithFields(log.Fields{
		"file": fileName,
	}).Info("Loading context")

	dat, e := ioutil.ReadFile(fileName)
	if e != nil {
		debug.WithFields(log.Fields{
			"file":  fileName,
			"error": e,
		}).Fatal("Unable to read file")
	}
	t := Context{}
	e = yaml.Unmarshal([]byte(dat), &t)
	if e != nil {
		debug.WithFields(log.Fields{
			"file":  fileName,
			"error": e,
		}).Fatal("Unable to unmarshal context")
	}
	return &t
}

// Loads cluster from YAML file
func loadClusterFromFile(fileName string) *Cluster {
	debug.WithFields(log.Fields{
		"file": fileName,
	}).Info("Loading cluster")

	dat, e := ioutil.ReadFile(fileName)
	if e != nil {
		debug.WithFields(log.Fields{
			"file":  fileName,
			"error": e,
		}).Fatal("Unable to read file")
	}
	t := Cluster{}
	e = yaml.Unmarshal([]byte(dat), &t)
	if e != nil {
		debug.WithFields(log.Fields{
			"file":  fileName,
			"error": e,
		}).Fatal("Unable to unmarshal cluster")
	}
	return &t
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
		}).Fatal("Directory contents can't be deleted")
	}

	fmt.Println("Aptomi state is now empty. Deleted all objects")
}

// Serialize object into YAML
func serializeObject(t interface{}) string {
	d, e := yaml.Marshal(&t)
	if e != nil {
		debug.WithFields(log.Fields{
			"object": t,
			"error":  e,
		}).Fatal("Can't serialize object", e)
	}
	return string(d)
}

// Prints slinga object onto screen
//noinspection GoUnusedFunction
func printObject(t interface{}) {
	fmt.Printf("--- dump:\n%s\n", serializeObject(t))

	m := make(map[interface{}]interface{})
	e := yaml.Unmarshal([]byte(serializeObject(t)), &m)
	if e != nil {
		fmt.Printf("error: %v", e)
	}
	fmt.Printf("%v\n\n", m)
}
