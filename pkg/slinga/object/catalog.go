package object

type ObjectCatalog struct {
	Kinds map[string]*ObjectInfo
}

type Constructor func() BaseObject

type ObjectInfo struct {
	Kind        string
	Constructor Constructor
}

func NewObjectCatalog() *ObjectCatalog {
	return &ObjectCatalog{
		make(map[string]*ObjectInfo),
	}
}

func (catalog *ObjectCatalog) Add(info *ObjectInfo) {
	catalog.Kinds[info.Kind] = info
}

func (catalog *ObjectCatalog) Get(kind string) *ObjectInfo { // todo return error if not found?
	// todo support default Kind?
	return catalog.Kinds[kind]
}

func (info *ObjectInfo) New() BaseObject {
	return info.Constructor()
}
