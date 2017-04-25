package main

import (
	"fmt"
	"log"
	"io/ioutil"
	"gopkg.in/yaml.v2"
)

type Allocation struct {
	Name     	string
	Criteria	[]string
	Labels   	map[string]map[string]string
}

type Context struct {
	Name     	string
	Service     string
	Criteria	[]string
	Labels   	map[string]map[string]string

	Allocations []Allocation
}

type Service struct {
	Name     	string
	Code     	string
	Labels   	map[string]map[string]string
}

// Load slinga service from file
func loadService(filename string) Service {
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

// Load slinga service from file
func loadContext(filename string) Context {
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

// Dump slinga service onto screen
func dumpService(t Service) {
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

// Dump slinga context onto screen
func dumpContext(t Context) {
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

func main() {
	dir := "/Users/ralekseenkov/go/src/aptomi-workspace/src/aptomi/slinga/example-1/"

	filenameSvc := dir + "service.kafka.yaml"
	s := loadService(filenameSvc)
	dumpService(s)

	filenameCtx := dir + "context.test.kafka.yaml"
	c := loadContext(filenameCtx)
	dumpContext(c)
}
