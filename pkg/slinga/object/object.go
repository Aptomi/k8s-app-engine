// Package object includes all unified Object stuff and ways to persist it
package object

// Kind represents object name like Cluster, Service, Context, etc.
type Kind string

// BaseObject interface represents unified object that could be stored in DB, accessed through API, etc.
type BaseObject interface {
	GetKind() Kind
	GetKey() Key
	GetUID() UID
	GetGeneration() Generation
	GetName() string
	GetNamespace() string
}

// Metadata represents standard metadata for unified objects.
// It implements BaseObject interface and it's enough to include it into any struct to make object DB and API
// layers compatible.
type Metadata struct {
	Kind       Kind
	UID        UID
	Generation Generation
	Name       string
	Namespace  string
	// TODO(slukjanov): do we need CreatedAt string? I think yes
	// TODO(slukjanov): should any object have owner?
}

// GetKind returns object's Kind
func (meta *Metadata) GetKind() Kind {
	return meta.Kind
}

// GetUID returns object's UID
func (meta *Metadata) GetUID() UID {
	return meta.UID
}

// GetGeneration returns object's Generation ("version")
func (meta *Metadata) GetGeneration() Generation {
	return meta.Generation
}

// GetKey returns object's Key
// TODO(slukjanov): should we only store key or uid / gen? or cache key inside metadata?
func (meta *Metadata) GetKey() Key {
	return KeyFromParts(meta.UID, meta.Generation)
}

// GetNamespace returns object's Namespace
func (meta *Metadata) GetNamespace() string {
	return meta.Namespace
}

// GetName returns object's Name
func (meta *Metadata) GetName() string {
	return meta.Name
}
