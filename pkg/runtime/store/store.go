package store

import (
	"github.com/Aptomi/aptomi/pkg/runtime"
)

type Interface interface {
	Close() error

	Save(storable runtime.Storable, opts ...SaveOpt) error
	Find(kind runtime.Kind, opts ...FindOpt) Finder
	Delete(kind runtime.Kind, opts ...DeleteOpt) Deleter
}

// Save

type SaveOpts struct {
	inPlace bool
	// todo could be used for saving RevisionData
	//forcedGen runtime.Generation
}

type SaveOpt func(opts *SaveOpts)

// Find

type Finder interface {
	First(runtime.Storable) error
	Last(runtime.Storable) error
	List([]runtime.Storable) error
}

type FindOpts struct {
	findGen   bool
	key       runtime.Key
	gen       runtime.Generation
	condition FieldEq
}

type FindOpt func(opts *FindOpts)

func WhereEq(name string, value interface{}) FindOpt {
	return func(opts *FindOpts) {
		opts.condition = FieldEq{name, value}
	}
}

type FieldEq struct {
	name  string
	value interface{}
}

// Delete

type Deleter interface {
	One() error
	All() (int, error)
}

type DeleteOpts struct {
}

type DeleteOpt func(opts *DeleteOpts)
