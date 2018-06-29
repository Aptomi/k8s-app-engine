package newstore

import (
	"github.com/Aptomi/aptomi/pkg/runtime"
)

//  todo key should be kind/namespace/name to support requests like all objects for namespace of specific kind?????

type Key = runtime.Key
type Kind = runtime.Kind
type Generation = runtime.Generation

type Storable interface {
	GetKind() Kind
	GetName() string
	GetNamespace() string
	GetGeneration() Generation    // if it returns 0 - treat as non-versioned object == no index generated for it, direct load
	SetGeneration(gen Generation) // for non-versioned should panic
}

type Interface interface {
	// todo add Close()

	Save(storable Storable, opts ...SaveOpt) error
	Find(kind Kind, opts ...FindOpt) Finder
	Delete(kind Kind, opts ...DeleteOpt) Deleter
}

// Save

type SaveOpts struct {
	inPlace   bool
	forcedGen Generation
}

type SaveOpt func(opts *SaveOpts)

// Find

type Finder interface {
	First(Storable) error
	Last(Storable) error
	List([]Storable) error
}

type FindOpts struct {
	findGen   bool
	key       Key
	gen       Generation
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
