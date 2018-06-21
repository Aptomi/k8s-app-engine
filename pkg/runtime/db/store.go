package db

type Store interface {
	Get(result Storable, key string) error
	List(result []Storable, query ...*Query) error

	// using key if provided or getting key from obj
	Delete(obj Storable, key string) error

	// do we need DeleteMatching(dataType Storable, query *Query) error
	// hm, choose needed from Insert, Update, UpdateMatching, Upsert

	Update(obj Storable, inPlace bool) error
}

type Query struct {
}
