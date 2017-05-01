package slinga

import (
	"fmt"
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
}

type GlobalState struct {
	Services []Service
	Contexts map[string][]Context
}

// Loads state from a directory
func loadGlobalStateFromDir(dir string) GlobalState {
	s := GlobalState{Contexts: make(map[string][]Context)}

	// read all services
	files, _ := filepath.Glob(dir + "service.*.yaml")
	sort.Strings(files)
	for _, f := range files {
		fmt.Println("Loading service from: ", f)
		s.Services = append(s.Services, loadServiceFromFile(f))
	}

	// read all contexts
	files, _ = filepath.Glob(dir + "context.*.yaml")
	sort.Strings(files)
	for _, f := range files {
		fmt.Println("Loading context from: ", f)
		context := loadContextFromFile(f)
		s.Contexts[context.Service] = append(s.Contexts[context.Service], context)
	}

	return s
}

// Loads service from YAML file
func loadServiceFromFile(filename string) Service {
	dat, e := ioutil.ReadFile(filename)
	if e != nil {
		panic(e)
	}
	t := Service{}
	e = yaml.Unmarshal([]byte(dat), &t)
	if e != nil {
		log.Fatalf("error: %v", e)
	}
	return t
}

// Loads context from YAML file
func loadContextFromFile(filename string) Context {
	dat, e := ioutil.ReadFile(filename)
	if e != nil {
		panic(e)
	}
	t := Context{}
	e = yaml.Unmarshal([]byte(dat), &t)
	if e != nil {
		log.Fatalf("error: %v", e)
	}
	return t
}

// Prints slinga object onto screen
//noinspection GoUnusedFunction
func printSlingaObject(t interface{}) {
	d, e := yaml.Marshal(&t)
	if e != nil {
		log.Fatalf("error: %v", e)
	}
	fmt.Printf("--- dump:\n%s\n", string(d))

	m := make(map[interface{}]interface{})
	e = yaml.Unmarshal([]byte(string(d)), &m)
	if e != nil {
		log.Fatalf("error: %v", e)
	}
	fmt.Printf("%v\n\n", m)
}

