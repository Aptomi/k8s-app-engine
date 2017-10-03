package lang

import (
	"github.com/Aptomi/aptomi/pkg/slinga/object"
)

/*
	This file declares all the necessary structures for Dependencies (User "wants" Service)
*/

var DependencyObject = &object.Info{
	Kind:        "dependency",
	Versioned:   true,
	Constructor: func() object.Base { return &Dependency{} },
}

// Dependency in a form <UserID> requested <Service> (and provided additional <Labels>)
type Dependency struct {
	Metadata

	UserID   string
	Contract string
	Labels   map[string]string
}

// GlobalDependencies represents the list of global dependencies (see the definition above)
type GlobalDependencies struct {
	// dependencies <service> -> list of dependencies
	DependenciesByContract map[string][]*Dependency
}

// NewGlobalDependencies creates and initializes a new empty list of global dependencies
func NewGlobalDependencies() *GlobalDependencies {
	return &GlobalDependencies{
		DependenciesByContract: make(map[string][]*Dependency),
	}
}

// AddDependency appends a single dependency to an existing object
func (src GlobalDependencies) AddDependency(dependency *Dependency) {
	src.DependenciesByContract[dependency.Contract] = append(src.DependenciesByContract[dependency.Contract], dependency)
}
