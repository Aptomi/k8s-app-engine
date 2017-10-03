package object

// Catalog contains a map of objects informational structures by their kind
type Catalog struct {
	Kinds map[string]*Info
}

// Constructor function definition to create flavors of base objects
type Constructor func() Base

// Info is an informational structure for an object, which defines its
// - kind
// - versioned
// - constructor
type Info struct {
	Kind        string
	Versioned   bool
	Constructor Constructor
}

// New creates a new instance of the object, given its properties defined by the informational structure
func (info *Info) New() Base {
	return info.Constructor()
}

// NewCatalog creates a new Catalog
func NewCatalog(infos ...*Info) *Catalog {
	catalog := &Catalog{
		make(map[string]*Info),
	}
	for _, info := range infos {
		catalog.Kinds[info.Kind] = info
	}
	return catalog
}

// Get looks up object informational structure given its kind
func (catalog *Catalog) Get(kind string) *Info {
	// todo return error if not found?
	// todo support default Kind?
	return catalog.Kinds[kind]
}
