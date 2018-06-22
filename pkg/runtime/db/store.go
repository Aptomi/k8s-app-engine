package db

type Store interface {
	Close() error

	// hm, choose needed from Insert, Update, UpdateMatching, Upsert
	Update(storable Storable, inPlace bool) error
	// using key if provided or getting key from obj
	Delete(storable Storable, key string) error
	// do we need DeleteMatching(dataType Storable, query *Query) error
	Get(result Storable, key string) error
	List(result []Storable, query ...*Query) error
}

type Query struct {
}
