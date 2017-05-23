package slinga

import (
	"github.com/golang/glog"
	"github.com/mattn/go-zglob"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"sort"
)

/*
	This file declares all the necessary structures for Dependencies (User -> Service)
*/

// Dependency in a form <UserID> requested <Service> (and provided additional <Labels>)
type Dependency struct {
	UserID  string
	Service string
	Labels  map[string]string

	// This field is populated when dependency gets resolved
	ResolvesTo string
}

// GlobalDependencies represents the list of global dependencies (see the definition above)
type GlobalDependencies struct {
	// dependencies <service> -> list of dependencies
	Dependencies map[string][]*Dependency
}

// NewGlobalDependencies creates and initializes a new empty list of global dependencies
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

// LoadDependenciesFromDir loads all dependencies from a given directory
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

// LoadDependenciesFromFile loads all dependencies from a given file
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
