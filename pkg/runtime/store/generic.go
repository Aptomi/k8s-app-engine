package store

import (
	"github.com/Aptomi/aptomi/pkg/config"
	"github.com/Aptomi/aptomi/pkg/runtime"
)

// Generic is an interface which describes basic operations on storable objects in DB
type Generic interface {
	Open(config.DB) error
	Close() error

	Get(key string) (runtime.Storable, error)
	GetGen(key string, gen runtime.Generation) (runtime.Versioned, error)

	List(prefix string) ([]runtime.Storable, error)
	ListGenerations(key string) ([]runtime.Storable, error)

	// Save could create new object in db or create new generation for existing object
	// todo(slukjanov): think about improving saving objects
	Save(runtime.Storable) (updated bool, err error)
	// Update always updates existing object in db and not creating new generation even for versioned objects
	// todo(slukjanov): introduce "status" for objects and don't update version when only status changed
	Update(runtime.Storable) (updated bool, err error)

	Delete(key string) error
}
