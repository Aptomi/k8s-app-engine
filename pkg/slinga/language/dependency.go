package language

import (
	"fmt"
	. "github.com/Aptomi/aptomi/pkg/slinga/object"
)

/*
	This file declares all the necessary structures for Dependencies (User "wants" Service)
*/

var DependencyObject = &Info{
	Kind:        "dependency",
	Constructor: func() Base { return &Dependency{} },
}

// Dependency in a form <UserID> requested <Service> (and provided additional <Labels>)
type Dependency struct {
	Metadata

	UserID  string
	Service string
	Labels  map[string]string

	// This fields are populated when dependency gets resolved
	Resolved   bool
	ServiceKey string
}

// GlobalDependencies represents the list of global dependencies (see the definition above)
// TODO: during serialization there is data duplication (as both fields get serialized). should prob avoid this
type GlobalDependencies struct {
	// dependencies <service> -> list of dependencies
	DependenciesByService map[string][]*Dependency

	// dependencies <id> -> dependency
	DependenciesByID map[string]*Dependency
}

// NewGlobalDependencies creates and initializes a new empty list of global dependencies
func NewGlobalDependencies() *GlobalDependencies {
	return &GlobalDependencies{
		DependenciesByService: make(map[string][]*Dependency),
		DependenciesByID:      make(map[string]*Dependency),
	}
}

// GetLabelSet applies set of transformations to labels
func (dependency *Dependency) GetLabelSet() LabelSet {
	return NewLabelSet(dependency.Labels)
}

// AddDependency appends a single dependency to an existing object
func (src GlobalDependencies) AddDependency(dependency *Dependency) {
	if len(dependency.GetID()) <= 0 {
		panic(fmt.Sprintf("Empty dependency ID: %+v", dependency))
	}
	src.DependenciesByService[dependency.Service] = append(src.DependenciesByService[dependency.Service], dependency)
	src.DependenciesByID[dependency.GetID()] = dependency
}

// TODO: added temporary method to deal with existing dependency IDs. Once we implement namespaces, may be this has to be re-thinked
func (dependency *Dependency) GetID() string {
	return dependency.Name
}
