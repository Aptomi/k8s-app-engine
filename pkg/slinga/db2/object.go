package db2

// ObjectKind represents object name like Cluster, Service, Context, etc.
type ObjectKind string

// Object interface represents unified object that could be stored in DB, accessed through API, etc.
type Object interface {
	GetKind() ObjectKind
	GetKey() Key
	GetUID() UID
	GetGeneration() Generation
	GetName() string
	GetNamespace() string
	GetSpec() interface{}

	// TODO(slukjanov): should any object have owner?
}

// BaseObject provides basic implementation of the Object interface
type BaseObject struct {
	Kind     ObjectKind
	Metadata BaseObjectMetadata
	Spec     interface{}
}

// BaseObjectMetadata represents standard metadata for unified objects
type BaseObjectMetadata struct {
	UID        UID
	Generation Generation
	Name       string
	Namespace  string
	// TODO(slukjanov): do we need CreatedAt string? I think yes
}

// GetKind returns object's ObjectKind
func (object *BaseObject) GetKind() ObjectKind {
	return object.Kind
}

// GetUID returns object's UID
func (object *BaseObject) GetUID() UID {
	return object.Metadata.UID
}

// GetGeneration returns object's Generation ("version")
func (object *BaseObject) GetGeneration() Generation {
	return object.Metadata.Generation
}

// GetKey returns object's Key
// TODO(slukjanov): should we only store key or uid / gen? or cache key inside metadata?
func (object *BaseObject) GetKey() Key {
	return KeyFromParts(object.Metadata.UID, object.Metadata.Generation)
}

// GetNamespace returns object's Namespace
func (object *BaseObject) GetNamespace() string {
	return object.Metadata.Namespace
}

// GetName returns object's Name
func (object *BaseObject) GetName() string {
	return object.Metadata.Name
}

// GetSpec returns object's Spec
func (object *BaseObject) GetSpec() interface{} {
	return object.Spec
}
