package object

type ObjectCatalog struct {
	Kinds map[string]*ObjectInfo
}

type Constructor func() BaseObject

type ObjectInfo struct {
	Kind        string
	Constructor Constructor
}

func NewObjectCatalog(infos ...*ObjectInfo) *ObjectCatalog {
	catalog := &ObjectCatalog{
		make(map[string]*ObjectInfo),
	}
	for _, info := range infos {
		catalog.Kinds[info.Kind] = info
	}
	return catalog
}

func (catalog *ObjectCatalog) Get(kind string) *ObjectInfo { // todo return error if not found?
	// todo support default Kind?
	return catalog.Kinds[kind]
}

func (info *ObjectInfo) New() BaseObject {
	return info.Constructor()
}
