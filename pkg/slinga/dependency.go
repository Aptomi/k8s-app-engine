package slinga

import (
	log "github.com/Sirupsen/logrus"
	"github.com/mattn/go-zglob"
	"sort"
)

/*
	This file declares all the necessary structures for Dependencies (User "wants" Service)
*/

// Dependency in a form <UserID> requested <Service> (and provided additional <Labels>)
type Dependency struct {
	ID       string
	UserID   string
	Service  string
	Labels   map[string]string
	Trace    bool
	Disabled bool

	// This field is populated when dependency gets resolved
	ResolvesTo string
}

// GlobalDependencies represents the list of global dependencies (see the definition above)
type GlobalDependencies struct {
	// dependencies <service> -> list of dependencies
	DependenciesByService map[string][]*Dependency

	// dependencies <id> -> dependency
	DependenciesByID map[string]*Dependency
}

func (src *GlobalDependencies) count() int {
	return countElements(src.DependenciesByID)
}

// NewGlobalDependencies creates and initializes a new empty list of global dependencies
func NewGlobalDependencies() GlobalDependencies {
	return GlobalDependencies{
		DependenciesByService: make(map[string][]*Dependency),
		DependenciesByID:      make(map[string]*Dependency),
	}
}

// Apply set of transformations to labels
func (dependency *Dependency) getLabelSet() LabelSet {
	return LabelSet{Labels: dependency.Labels}
}

// SetTrace enable tracing (detailed engine output) for all dependencies
func (src *GlobalDependencies) SetTrace(trace bool) {
	if trace {
		for _, d := range src.DependenciesByID {
			d.Trace = true
		}
	}
}

// Append a single dependency to an existing object
func (src GlobalDependencies) appendDependency(dependency *Dependency) {
	if len(dependency.ID) <= 0 {
		debug.WithFields(log.Fields{
			"dependency": dependency,
		}).Panic("Empty dependency ID")
	}
	src.DependenciesByService[dependency.Service] = append(src.DependenciesByService[dependency.Service], dependency)
	src.DependenciesByID[dependency.ID] = dependency
}

// Copy the whole structure with dependencies
func (src GlobalDependencies) makeCopy() GlobalDependencies {
	result := NewGlobalDependencies()
	for _, v := range src.DependenciesByID {
		result.appendDependency(v)
	}
	return result
}

// LoadDependenciesFromDir loads all dependencies from a given directory
func LoadDependenciesFromDir(baseDir string) GlobalDependencies {
	// read all services
	files, _ := zglob.Glob(GetAptomiObjectFilePatternYaml(baseDir, TypeDependencies))
	sort.Strings(files)
	result := NewGlobalDependencies()
	for _, fileName := range files {
		t := loadDependenciesFromFile(fileName)
		for _, d := range t {
			if d.Disabled {
				continue
			}
			result.appendDependency(d)
		}
	}
	return result
}
