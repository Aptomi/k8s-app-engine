package lang

import (
	"github.com/Aptomi/aptomi/pkg/runtime"
)

// Metadata is an object metadata implementation (Namespace, Kind, Name, Generation) which works for all standard objects.
// Namespace defines in which namespace the object is defined. An object always gets placed in only one namespace.
// Kind describes type of the object (e.g. Bundle, Service, Cluster, etc)
// Name is a user-provided string identifier of an object. Names are usually human readable and must be unique across
// objects within the same namespace and the same object kind.
type Metadata struct {
	Namespace  string             `yaml:",omitempty" validate:"identifier"`
	Name       string             `yaml:",omitempty" validate:"identifier"`
	Generation runtime.Generation `yaml:",omitempty"`
	Deleted    bool               `yaml:",omitempty"`
}

// GetNamespace returns object namespace
func (meta *Metadata) GetNamespace() string {
	return meta.Namespace
}

// GetName returns object name
func (meta *Metadata) GetName() string {
	return meta.Name
}

// GetGeneration returns object generation
func (meta *Metadata) GetGeneration() runtime.Generation {
	return meta.Generation
}

// SetGeneration sets object generation
func (meta *Metadata) SetGeneration(generation runtime.Generation) {
	meta.Generation = generation
}

// IsDeleted returns if object deleted or not
func (meta *Metadata) IsDeleted() bool {
	return meta.Deleted
}

// SetDeleted sets object deleted flag
func (meta *Metadata) SetDeleted(deleted bool) {
	meta.Deleted = deleted
}
