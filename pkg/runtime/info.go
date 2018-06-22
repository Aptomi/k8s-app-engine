package runtime

// Info represents list of additional characteristics of the runtime object
type Info struct {
	Kind        Kind
	Storable    bool
	Versioned   bool
	Deletable   bool
	Constructor Constructor
}

// Constructor is a function to get instance of the specific object
type Constructor func() Object

// New creates a new instance of the specific object defined in Info
func (info *Info) New() Object {
	return info.Constructor()
}

// GetTypeKind returns TypeKind instance for the object described by info
func (info *Info) GetTypeKind() TypeKind {
	return TypeKind{Kind: info.Kind}
}
