package slinga

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/mattn/go-zglob"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"sort"
	"strings"
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

// AddObjectsToPolicy adds an object to the policy (basically, copies one or more files to Aptomi DB directory)
func AddObjectsToPolicy(aptFilter AptomiOject, args ...string) {
	for _, v := range args {
		stat, err := os.Stat(v)
		if err == nil {
			if stat.IsDir() {
				// if it's a directory, process all yaml files in it
				files, _ := zglob.Glob(v + "/**/*.*")
				for _, f := range files {
					AddObjectsToPolicy(aptFilter, f)
				}
			} else {
				// if it's a file, copy it over
				name := stat.Name()
				idx := strings.Index(name, ".")
				if idx >= 0 {
					objectType := name[0:idx]
					apt, ok := AptomiObjectsCanBeAdded[objectType]

					// Charts don't have to start with object prefix
					if aptFilter == Charts && strings.HasSuffix(name, ".tgz") {
						ok = true
						apt = aptFilter
					}

					if ok {
						if apt == aptFilter {
							err := copyFile(v, GetAptomiObjectDir(GetAptomiBaseDir(), apt)+"/"+name)
							if err != nil {
								fmt.Printf("Unable to add %s to aptomi policy: %s\n", objectType, name)
								debug.WithFields(log.Fields{
									"fileName":   v,
									"name":       name,
									"objectType": objectType,
									"error":      err,
								}).Fatal("Unable to add object to aptomi policy")
							} else {
								fmt.Printf("Adding %s to aptomi policy: %s\n", objectType, name)
							}
						}
					} else {
						debug.WithFields(log.Fields{
							"fileName":   v,
							"name":       name,
							"objectType": objectType,
						}).Warning("Invalid object type. Must be within defined Aptomi object types")
					}
				} else {
					debug.WithFields(log.Fields{
						"fileName": v,
						"name":     name,
					}).Warning("File name must be prefixed with object type")
				}
			}
		}
	}
}

// LoadPolicyFromDir loads policy from a directory, recursively processing all files
func LoadPolicyFromDir(baseDir string) Policy {
	s := Policy{
		Services: make(map[string]*Service),
		Contexts: make(map[string][]*Context),
		Clusters: make(map[string]*Cluster),
	}

	// read all clusters
	files, _ := zglob.Glob(GetAptomiObjectDir(baseDir, Clusters) + "/**/cluster.*.yaml")
	sort.Strings(files)
	for _, f := range files {
		cluster := loadClusterFromFile(f)
		s.Clusters[cluster.Name] = cluster
	}

	// read all services
	files, _ = zglob.Glob(GetAptomiObjectDir(baseDir, Services) + "/**/service.*.yaml")
	sort.Strings(files)
	for _, f := range files {
		service := loadServiceFromFile(f)
		s.Services[service.Name] = service
	}

	// read all contexts
	files, _ = zglob.Glob(GetAptomiObjectDir(baseDir, Contexts) + "/**/context.*.yaml")
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

// ResetAptomiState fully resets aptomi state by deleting all file from its database. Including policy, logs, etc
func ResetAptomiState() {
	debug.WithFields(log.Fields{
		"baseDir": GetAptomiBaseDir(),
	}).Info("Resetting aptomi state")

	files, _ := zglob.Glob(GetAptomiBaseDir() + "/**/*.*")
	for _, f := range files {
		err := os.Remove(f)
		if err != nil {
			debug.WithFields(log.Fields{
				"file":  f,
				"error": err,
			}).Fatal("Unable to remove file")
		}
	}

	if len(files) > 0 {
		fmt.Printf("Aptomi state is now empty. Deleted %d objects\n", len(files))
	} else {
		fmt.Println("Aptomi state is empty")
	}
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
