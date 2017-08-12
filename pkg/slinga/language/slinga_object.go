package language

type SlingaObjectType string

const (
	// Everything which is an "input" policy object (cluster, context, rule, service)
	TypePolicy SlingaObjectType = "policy"
)

// SlingaObject is a universal object
type SlingaObject struct {
	// Kind is a type of an object
	Kind string

	// Metadata is a set of pre-defined fields that every object has
	Metadata SlingaObjectMetadata

	// Spec will be parsed using a parser, which is specific for a given object kind
	Spec interface{}
}

// Metadata is a set of pre-defined fields that every object has
type SlingaObjectMetadata struct {
	// Name of a namespace within aptomi
	Namespace string

	// Name of an object (unique within a namespace for a given object kind)
	Name string
}

// SlingaObjectInterface defines common methods on SlingaObject
type SlingaObjectInterface interface {
	GetNamespace() string
	GetName() string
	GetObjectType() SlingaObjectType

	// Returns unique object key (by default: kind -> namespace -> name, standard implementation in SlingaObject)
	// GetKey() string

	// Returns diff between two objects (by default: standard implementation in SlingaObject)
	// GetDiff() string
}

func (object *SlingaObject) GetNamespace() string {
	return object.Metadata.Namespace
}

func (object *SlingaObject) GetName() string {
	return object.Metadata.Name
}
