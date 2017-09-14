package object

type Catalog struct {
	Kinds map[string]*Info
}

type Constructor func() Base

type Info struct {
	Kind        string
	Versioned   bool
	Constructor Constructor
}

func NewObjectCatalog(infos ...*Info) *Catalog {
	catalog := &Catalog{
		make(map[string]*Info),
	}
	for _, info := range infos {
		catalog.Kinds[info.Kind] = info
	}
	return catalog
}

func (catalog *Catalog) Get(kind string) *Info { // todo return error if not found?
	// todo support default Kind?
	return catalog.Kinds[kind]
}

func (info *Info) New() Base {
	return info.Constructor()
}
