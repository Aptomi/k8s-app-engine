package lang

import (
	"github.com/Aptomi/aptomi/pkg/slinga/object"
	"strings"
)

// Metadata is an object metadata implementation (NS, Kind, Name, Generation) which works for all standard objects
type Metadata struct {
	Namespace  string
	Kind       string
	Name       string
	Generation object.Generation
}

// GetKey returns object key as NS:Kind:Name:Generation
func (meta *Metadata) GetKey() string {
	return strings.Join([]string{meta.Namespace, meta.Kind, meta.Name}, object.KeySeparator)
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
