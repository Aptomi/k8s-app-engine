package object

type ObjectCatalog struct {
	Infos map[Kind]*ObjectInfo
}

type Constructor func() BaseObject

type ObjectInfo struct {
	Kind        Kind
	Constructor Constructor
}

func NewObjectCatalog() *ObjectCatalog {
	return &ObjectCatalog{
		make(map[Kind]*ObjectInfo, 0),
	}
}

func (catalog *ObjectCatalog) Add(info *ObjectInfo) {
	catalog.Infos[info.Kind] = info
}

func (catalog *ObjectCatalog) Get(kind Kind) *ObjectInfo { // todo return error if not found?
	// todo support default Kind?
	return catalog.Infos[kind]
}

func (info *ObjectInfo) New() BaseObject {
	return info.Constructor()
}
