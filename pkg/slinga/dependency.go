package slinga

import (
	"github.com/golang/glog"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"github.com/mattn/go-zglob"
	"sort"
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

func NewGlobalDependencies() GlobalDependencies {
	return GlobalDependencies{Dependencies: make(map[string][]*Dependency)}
}

// Apply set of transformations to labels
func (dependency *Dependency) getLabelSet() LabelSet {
	return LabelSet{Labels: dependency.Labels}
}

// Merge
func (src GlobalDependencies) appendDependencies(ops GlobalDependencies) GlobalDependencies {
	result := NewGlobalDependencies()
	for k, v := range src.Dependencies {
		result.Dependencies[k] = append(result.Dependencies[k], v...)
	}
	for k, v := range ops.Dependencies {
		result.Dependencies[k] = append(result.Dependencies[k], v...)
	}
	return result
}

// Loads dependencies from directory
func LoadDependenciesFromDir(dir string) GlobalDependencies {
	// read all services
	files, _ := zglob.Glob(dir + "/**/dependencies.*.yaml")
	sort.Strings(files)
	r := NewGlobalDependencies()
	for _, f := range files {
		glog.Infof("Loading dependencies from %s", f)
		dependencies := LoadDependenciesFromFile(f)
		r = r.appendDependencies(dependencies)
	}
	return r
}

// Loads dependencies from file
func LoadDependenciesFromFile(filename string) GlobalDependencies {
	dat, e := ioutil.ReadFile(filename)
	if e != nil {
		glog.Fatalf("Unable to read file: %v", e)
	}
	t := []*Dependency{}
	e = yaml.Unmarshal([]byte(dat), &t)
	if e != nil {
		glog.Fatalf("Unable to unmarshal dependencies: %v", e)
	}
	r := NewGlobalDependencies()
	for _, d := range t {
		r.Dependencies[d.Service] = append(r.Dependencies[d.Service], d)
	}
	return r
}
