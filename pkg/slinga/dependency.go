package slinga

import (
	"github.com/mattn/go-zglob"
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
	Trace   bool

	// This field is populated when dependency gets resolved
	ResolvesTo string
}

// GlobalDependencies represents the list of global dependencies (see the definition above)
type GlobalDependencies struct {
	// dependencies <service> -> list of dependencies
	Dependencies map[string][]*Dependency
}

func (src *GlobalDependencies) count() int {
	return countElements(src.Dependencies)
}

// NewGlobalDependencies creates and initializes a new empty list of global dependencies
func NewGlobalDependencies() GlobalDependencies {
	return GlobalDependencies{Dependencies: make(map[string][]*Dependency)}
}

// LoadDependenciesFromFile loads all dependencies from a given file
func LoadDependenciesFromFile(fileName string) GlobalDependencies {
	r := NewGlobalDependencies()
	t := loadDependenciesFromFile(fileName)
	for _, d := range t {
		r.Dependencies[d.Service] = append(r.Dependencies[d.Service], d)
	}
	return r
}

// Apply set of transformations to labels
func (dependency *Dependency) getLabelSet() LabelSet {
	return LabelSet{Labels: dependency.Labels}
}

// SetTrace enable tracing (detailed engine output) for all dependencies
func (src *GlobalDependencies) SetTrace(trace bool) {
	if trace {
		for _, serviceDeps := range src.Dependencies {
			for _, v := range serviceDeps {
				v.Trace = true
			}
		}
	}
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

// Merge
func (src GlobalDependencies) appendDependency(ops *Dependency) GlobalDependencies {
	result := NewGlobalDependencies()
	for k, v := range src.Dependencies {
		result.Dependencies[k] = append(result.Dependencies[k], v...)
	}
	result.Dependencies[ops.Service] = append(result.Dependencies[ops.Service], ops)
	return result
}

// LoadDependenciesFromDir loads all dependencies from a given directory
func LoadDependenciesFromDir(baseDir string) GlobalDependencies {
	// read all services
	files, _ := zglob.Glob(GetAptomiObjectFilePatternYaml(baseDir, TypeDependencies))
	sort.Strings(files)
	r := NewGlobalDependencies()
	for _, f := range files {
		dependencies := LoadDependenciesFromFile(f)
		r = r.appendDependencies(dependencies)
	}
	return r
}
