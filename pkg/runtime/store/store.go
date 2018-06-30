package store

import (
	"github.com/Aptomi/aptomi/pkg/runtime"
)

type Interface interface {
	Close() error

	Save(storable runtime.Storable, opts ...SaveOpt) error
	Find(kind runtime.Kind, opts ...FindOpt) Finder
	Delete(kind runtime.Kind, key runtime.Key) error
}

type Finder interface {
	One(runtime.Storable) error
	List(interface{}) error
}
