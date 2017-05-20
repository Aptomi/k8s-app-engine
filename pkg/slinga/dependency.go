package slinga

import (
	"github.com/golang/glog"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

/*
	This file declares all the necessary structures for Dependencies (User -> Service)
*/

type Dependency struct {
	UserId  string
	Service string
	Labels map[string]string
}

type GlobalDependencies struct {
	// dependencies <service> -> list of dependencies
	Dependencies map[string][]*Dependency
}

// Apply set of transformations to labels
func (dependency *Dependency) getLabelSet() LabelSet {
	return LabelSet{Labels: dependency.Labels}
}

// Loads users from YAML file
func LoadDependenciesFromDir(dir string) GlobalDependencies {
	dat, e := ioutil.ReadFile(dir + "/dependencies.yaml")
	if e != nil {
		glog.Fatalf("Unable to read file: %v", e)
	}
	t := []*Dependency{}
	e = yaml.Unmarshal([]byte(dat), &t)
	if e != nil {
		glog.Fatalf("Unable to unmarshal user: %v", e)
	}
	r := GlobalDependencies{Dependencies: make(map[string][]*Dependency)}
	for _, d := range t {
		r.Dependencies[d.Service] = append(r.Dependencies[d.Service], d)
	}
	return r
}
