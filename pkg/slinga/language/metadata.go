package language

import (
	"github.com/Aptomi/aptomi/pkg/slinga/object"
	"strings"
)

type Metadata struct {
	Namespace  string
	Kind       string
	Name       string
	Generation object.Generation
}

func (meta *Metadata) GetKey() string {
	// todo cache key?
	return strings.Join([]string{meta.Namespace, meta.Kind, meta.Name}, object.KeySeparator)
}

func (meta *Metadata) GetNamespace() string {
	return meta.Namespace
}

func (meta *Metadata) GetKind() string {
	return meta.Kind
}

func (meta *Metadata) GetName() string {
	return meta.Name
}

func (meta *Metadata) GetGeneration() object.Generation {
	return meta.Generation
}

func (meta *Metadata) SetGeneration(generation object.Generation) {
	meta.Generation = generation
}
