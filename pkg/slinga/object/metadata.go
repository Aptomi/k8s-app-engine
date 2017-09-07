package object

import "strings"

type Metadata struct {
	Namespace  string
	Kind       string
	Name       string
	Generation Generation
	// TODO(slukjanov): do we need CreatedAt string? I think yes
	// TODO(slukjanov): should any object have owner?
}

func (meta *Metadata) GetKey() string {
	return strings.Join([]string{meta.Namespace, meta.Kind, meta.Name}, KeySeparator)
}

// GetNamespace returns object's Namespace
func (meta *Metadata) GetNamespace() string {
	return meta.Namespace
}

// GetKind returns object's Kind
func (meta *Metadata) GetKind() string {
	return meta.Kind
}

// GetName returns object's Name
func (meta *Metadata) GetName() string {
	return meta.Name
}

// GetGeneration returns object's Generation ("version")
func (meta *Metadata) GetGeneration() Generation {
	return meta.Generation
}
