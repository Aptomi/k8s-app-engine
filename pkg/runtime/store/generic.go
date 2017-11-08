package store

import (
	"github.com/Aptomi/aptomi/pkg/config"
	"github.com/Aptomi/aptomi/pkg/runtime"
)

// Generic is an interface which describes basic operations on storable objects in DB
type Generic interface {
	Open(config.DB) error
	Close() error

	// todo replace string with Key?
	Get(key string) (runtime.Storable, error)
	GetGen(key string, gen runtime.Generation) (runtime.Versioned, error)

	List(prefix string) ([]runtime.Storable, error)

	// todo should it accept key as well?
	Save(runtime.Storable) (updated bool, err error)

	Delete(key string) error
}
