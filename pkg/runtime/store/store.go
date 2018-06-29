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

// Save

type SaveOpts struct {
	inPlace bool
	// todo could be used for saving RevisionData
	//forcedGen runtime.Generation
}

type SaveOpt func(opts *SaveOpts)

// Find

type Finder interface {
	One(runtime.Storable) error
	List(interface{}) error
}

type FindOpts struct {
	findGen   bool
	key       runtime.Key
	gen       runtime.Generation
	condition FieldEq
}

type FindOpt func(opts *FindOpts)

func WithKey(key runtime.Key) FindOpt {
	return func(opts *FindOpts) {
		opts.key = key
	}
}

func WithGen(gen runtime.Generation) FindOpt {
	return func(opts *FindOpts) {
		opts.gen = gen
	}
}

func WithWhereEq(name string, values ...interface{}) FindOpt {
	return func(opts *FindOpts) {
		opts.condition = FieldEq{name, values}
	}
}

func WithGetLast() FindOpt {
	return func(opts *FindOpts) {
		// todo
	}
}

func WithGetFirst() FindOpt {
	return func(opts *FindOpts) {
		// todo
	}
}

type FieldEq struct {
	name   string
	values []interface{}
}
