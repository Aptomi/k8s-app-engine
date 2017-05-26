package slinga

import (
	"github.com/golang/glog"
	"github.com/mattn/go-zglob"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"sort"
)

/*
	This file declares all the necessary structures for Slinga YAML file to be successfully parsed
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
	Discovery	 interface{}
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

// Policy is a global policy object with services and contexts
type Policy struct {
	Services map[string]*Service
	Contexts map[string][]*Context
}

// LoadPolicyFromDir loads policy from a directory, recursively processing all files
func LoadPolicyFromDir(dir string) Policy {
	s := Policy{
		Services: make(map[string]*Service),
		Contexts: make(map[string][]*Context),
	}

	// read all services
	files, _ := zglob.Glob(dir + "/**/service.*.yaml")
	sort.Strings(files)
	for _, f := range files {
		glog.Infof("Loading service from %s", f)
		service := loadServiceFromFile(f)
		s.Services[service.Name] = service
	}

	// read all contexts
	files, _ = zglob.Glob(dir + "/**/context.*.yaml")
	sort.Strings(files)
	for _, f := range files {
		glog.Infof("Loading context from %s", f)
		context := loadContextFromFile(f)
		s.Contexts[context.Service] = append(s.Contexts[context.Service], context)
	}

	return s
}

// Loads service from YAML file
func loadServiceFromFile(filename string) *Service {
	dat, e := ioutil.ReadFile(filename)
	if e != nil {
		glog.Fatalf("Unable to read file: %v", e)
	}
	t := Service{}
	e = yaml.Unmarshal([]byte(dat), &t)
	if e != nil {
		glog.Fatalf("Unable to unmarshal service: %v", e)
	}
	return &t
}

// Loads context from YAML file
func loadContextFromFile(filename string) *Context {
	dat, e := ioutil.ReadFile(filename)
	if e != nil {
		glog.Fatalf("Unable to read file: %v", e)
	}
	t := Context{}
	e = yaml.Unmarshal([]byte(dat), &t)
	if e != nil {
		glog.Fatalf("Unable to unmarshal context: %v", e)
	}
	return &t
}

// Serialize object into YAML
func serializeObject(t interface{}) string {
	d, e := yaml.Marshal(&t)
	if e != nil {
		glog.Fatalf("error: %v", e)
	}
	return string(d)
}

// Prints slinga object onto screen
//noinspection GoUnusedFunction
func printObject(t interface{}) {
	glog.Infof("--- dump:\n%s\n", serializeObject(t))

	m := make(map[interface{}]interface{})
	e := yaml.Unmarshal([]byte(serializeObject(t)), &m)
	if e != nil {
		glog.Fatalf("error: %v", e)
	}
	glog.Infof("%v\n\n", m)
}
