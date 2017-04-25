package main

import (
	"fmt"
	"log"
	"io/ioutil"
	"gopkg.in/yaml.v2"
)

type Context struct {
	Labels map[string]string
	Instances []string
}

type Service struct {
	Name string
	Code string
	Labels map[string]string
	Contexts map[string]Context
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

// Dump slinga service onto screen
func dump(t Service) {
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
	filename := "/Users/ralekseenkov/go/aptomi-workspace/src/aptomi/slinga/example-1/service.kafka.yaml"
	t := loadService(filename)
	dump(t)
}
