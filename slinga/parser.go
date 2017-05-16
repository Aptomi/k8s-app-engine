package slinga

import (
	"log"
	"io/ioutil"
	"gopkg.in/yaml.v2"
	"path/filepath"
	"sort"
)

/*
 	This file declares all the necessary structures for Slinga YAML file to be successfully parsed
  */

type LabelOperations map[string]map[string]string

type Allocation struct {
	Name     string
	Criteria []string
	Labels   LabelOperations

	// Evaluated field (when parameters in name are substituted with real values)
	NameResolved string
}

type Context struct {
	Name        string
	Service     string
	Criteria    []string
	Labels      LabelOperations
	Allocations []Allocation
}

type ServiceComponent struct {
	Name         string
	Service      string
	Code         string
	Dependencies []string
	Labels       LabelOperations
}

type Service struct {
	Name       string
	Labels     LabelOperations
	Components []ServiceComponent

	// Evaluated field (when components are sorted in instantiation order)
	ComponentsMap     map[string]ServiceComponent
	ComponentsOrdered []ServiceComponent
}

type Policy struct {
	Services map[string]Service
	Contexts map[string][]Context
}

// Loads policy from a directory
func LoadPolicyFromDir(dir string) Policy {
	s := Policy{
		Services: make(map[string]Service),
		Contexts: make(map[string][]Context),
	}

	// read all services
	files, _ := filepath.Glob(dir + "/policy/service.*.yaml")
	sort.Strings(files)
	for _, f := range files {
		log.Printf("Loading service from %s", f)
		service := loadServiceFromFile(f)
		s.Services[service.Name] = service
	}

	// read all contexts
	files, _ = filepath.Glob(dir + "/policy/context.*.yaml")
	sort.Strings(files)
	for _, f := range files {
		log.Printf("Loading context from %s", f)
		context := loadContextFromFile(f)
		s.Contexts[context.Service] = append(s.Contexts[context.Service], context)
	}

	return s
}

// Loads service from YAML file
func loadServiceFromFile(filename string) Service {
	dat, e := ioutil.ReadFile(filename)
	if e != nil {
		log.Fatalf("Unable to read file: %v", e)
	}
	t := Service{}
	e = yaml.Unmarshal([]byte(dat), &t)
	if e != nil {
		log.Fatalf("Unable to unmarshal service: %v", e)
	}
	return t
}

// Loads context from YAML file
func loadContextFromFile(filename string) Context {
	dat, e := ioutil.ReadFile(filename)
	if e != nil {
		log.Fatalf("Unable to read file: %v", e)
	}
	t := Context{}
	e = yaml.Unmarshal([]byte(dat), &t)
	if e != nil {
		log.Fatalf("Unable to unmarshal context: %v", e)
	}
	return t
}

// Serialize object into YAML
func serializeObject(t interface{}) string {
	d, e := yaml.Marshal(&t)
	if e != nil {
		log.Fatalf("error: %v", e)
	}
	return string(d)
}

// Prints slinga object onto screen
//noinspection GoUnusedFunction
func printObject(t interface{}) {
	log.Printf("--- dump:\n%s\n", serializeObject(t))

	m := make(map[interface{}]interface{})
	e := yaml.Unmarshal([]byte(serializeObject(t)), &m)
	if e != nil {
		log.Fatalf("error: %v", e)
	}
	log.Printf("%v\n\n", m)
}
