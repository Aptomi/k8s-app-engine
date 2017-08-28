// Package object includes all unified Object stuff and ways to persist it
package object

// Metadata represents standard metadata for unified objects.
// It implements BaseObject interface and it's enough to include it into any struct to make object DB and API
// layers compatible.
type Metadata struct {
	Namespace  string
	Kind       string
	Name       string
	RandAddon  string
	Generation Generation
	// TODO(slukjanov): do we need CreatedAt string? I think yes
	// TODO(slukjanov): should any object have owner?
}

// GetKey returns object's Key
func (meta *Metadata) GetKey() Key {
	return KeyFromParts("", meta.Namespace, meta.Kind, meta.Name, meta.RandAddon, meta.Generation)
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

func (meta *Metadata) GetRandAddon() string {
	return meta.RandAddon
}

// GetGeneration returns object's Generation ("version")
func (meta *Metadata) GetGeneration() Generation {
	return meta.Generation
}
