package store

import (
	. "github.com/Aptomi/aptomi/pkg/slinga/object"
)

type ObjectStore interface {
	Open(connection string) error
	Close() error

	Save(BaseObject) error

	// + SaveMany (in one tx)
	// + GetManyByKeys
	// + Find(namespace, kind, name, rand, generation) - if some == "" or 0 don't match by it

	GetByKey(Key) (BaseObject, error)
}
