package lang

import (
	"github.com/Aptomi/aptomi/pkg/object"
)

// Metadata is an object metadata implementation (Namespace, Kind, Name, Generation) which works for all standard objects.
// Namespace defines in which namespace the object is defined. An object always gets placed in only one namespace.
// Kind describes type of the object (e.g. Service, Contract, Cluster, etc)
// Name is a user-provided string identifier of an object. Names are usually human readable and must be unique across
// objects within the same namespace and the same object kind.
type Metadata struct {
	Namespace  string `validate:"identifier"`
	Kind       string `validate:"identifier"`
	Name       string `validate:"identifier"`
	Generation object.Generation
}

// GetNamespace returns object namespace
func (meta *Metadata) GetNamespace() string {
	return meta.Namespace
}

// GetKind returns object kind
func (meta *Metadata) GetKind() string {
	return meta.Kind
}

// GetName returns object name
func (meta *Metadata) GetName() string {
	return meta.Name
}

// GetGeneration returns object generation
func (meta *Metadata) GetGeneration() object.Generation {
	return meta.Generation
}

// SetGeneration sets object generation
func (meta *Metadata) SetGeneration(generation object.Generation) {
	meta.Generation = generation
}
