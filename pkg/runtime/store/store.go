package store

import (
	"github.com/Aptomi/aptomi/pkg/runtime"
)

// Interface represents API of the object storage
type Interface interface {
	Close() error

	Save(storable runtime.Storable, opts ...SaveOpt) (bool, error)
	Find(kind runtime.Kind, result interface{}, opts ...FindOpt) error
	Delete(kind runtime.Kind, key runtime.Key) error
}
