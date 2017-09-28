package store

import (
	"github.com/Aptomi/aptomi/pkg/slinga/object"
	"io"
)

type ObjectStore interface {
	Open(connection string) error
	Close() error

	Save(object.Base) (updated bool, err error)

	// + SaveMany (in one tx)
	// + GetManyByKeys
	// + Find(namespace, kind, name, rand, generation) - if some == "" or 0 don't match by it

	GetByName(namespace string, kind string, name string, gen object.Generation) (object.Base, error)

	Dump(writer io.Writer) error
}
