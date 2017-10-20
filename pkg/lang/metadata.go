package lang

import (
	"github.com/Aptomi/aptomi/pkg/object"
)

// Metadata is an object metadata implementation (NS, Kind, Name, Generation) which works for all standard objects
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
