package slinga

import (
	"log"
	"io/ioutil"
	"gopkg.in/yaml.v2"
)

/*
 	This file declares all the necessary structures for Dependencies (User -> Service)
  */

type Dependency struct {
	UserId       string
	Service     string
}

type GlobalDependencies struct {
	// dependencies <service> -> list of users
	Dependencies map[string][]string
}

// Loads users from YAML file
func LoadDependenciesFromDir(dir string) GlobalDependencies {
	dat, e := ioutil.ReadFile(dir + "/dependencies.yaml")
	if e != nil {
		log.Fatalf("Unable to read file: %v", e)
	}
	t := []Dependency{}
	e = yaml.Unmarshal([]byte(dat), &t)
	if e != nil {
		log.Fatalf("Unable to unmarshal user: %v", e)
	}
	r := GlobalDependencies{Dependencies: make(map[string][]string)}
	for _, d := range t {
		r.Dependencies[d.Service] = append(r.Dependencies[d.Service], d.UserId)
	}
	return r
}
