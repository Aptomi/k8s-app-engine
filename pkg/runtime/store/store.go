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
	replace     bool
	forceNewGen bool
	forceGen    runtime.Generation
}

func (opts *SaveOpts) IsReplace() bool {
	return opts.replace
}

func (opts *SaveOpts) IsForceNewGen() bool {
	return opts.forceNewGen
}

func (opts *SaveOpts) GetForceGen() runtime.Generation {
	return opts.forceGen
}

type SaveOpt func(opts *SaveOpts)

func NewSaveOpts(opts []SaveOpt) *SaveOpts {
	saveOpts := &SaveOpts{}
	for _, opt := range opts {
		opt(saveOpts)
	}

	return saveOpts
}

func WithReplace() SaveOpt {
	return func(opts *SaveOpts) {
		if opts.forceNewGen {
			panic("can't use WithReplace when WithForceNewGen already used")
		}
		if opts.forceGen != 0 {
			panic("can't use WithReplace when WithForceGen already used (as it's implicitly allows replacement)")
		}
		if opts.replace {
			panic("can't use WithReplace more then one time")
		}

		opts.replace = true
	}
}

func WithForceNewGen() SaveOpt {
	return func(opts *SaveOpts) {
		if opts.replace {
			panic("can't use WithForceNewGen when WithReplace already used")
		}
		if opts.forceGen != 0 {
			panic("can't use WithForceNewGen when WithForceGen already used")
		}
		if opts.forceNewGen {
			panic("can't use WithForceNewGen more then one time")
		}

		opts.forceNewGen = true
	}
}

func WithForceGen(gen runtime.Generation) SaveOpt {
	return func(opts *SaveOpts) {
		if opts.replace {
			panic("can't use WithForceGen when WithReplace already used (as it's implicitly allows replacement)")
		}
		if opts.forceNewGen {
			panic("can't use WithNewGen when WithForceGen already used")
		}
		if opts.forceGen != 0 {
			panic("can't use WithForceGen more then one time")
		}

		opts.forceGen = gen
	}
}

// Find

type Finder interface {
	One(runtime.Storable) error
	List(interface{}) error
}

type FindOpts struct {
	key           runtime.Key
	gen           runtime.Generation
	fieldEqName   string
	fieldEqValues []interface{}
	getLast       bool
	getFirst      bool
}

func (opts *FindOpts) GetKey() runtime.Key {
	return opts.key
}

func (opts *FindOpts) GetGen() runtime.Generation {
	return opts.gen
}

func (opts *FindOpts) GetFieldEqName() string {
	return opts.fieldEqName
}

func (opts *FindOpts) GetFieldEqValues() []interface{} {
	return opts.fieldEqValues
}

func (opts *FindOpts) IsGetFirst() bool {
	return opts.getFirst
}

func (opts *FindOpts) IsGetLast() bool {
	return opts.getLast
}

type FindOpt func(opts *FindOpts)

func NewFindOpts(opts []FindOpt) *FindOpts {
	findOpts := &FindOpts{}
	for _, opt := range opts {
		opt(findOpts)
	}

	return findOpts
}

func WithKey(key runtime.Key) FindOpt {
	return func(opts *FindOpts) {
		if opts.key != "" {
			panic("can't use WithKey more then one time")
		}

		opts.key = key
	}
}

func WithGen(gen runtime.Generation) FindOpt {
	return func(opts *FindOpts) {
		if opts.key == "" {
			panic("can't use WithGen without WithKey (key isn't set)")
		}
		if opts.gen != 0 {
			panic("can't use WithGen more then one time")
		}

		opts.gen = gen
	}
}

func WithWhereEq(name string, values ...interface{}) FindOpt {
	return func(opts *FindOpts) {
		if name == "" {
			panic("can't use WithWhereEq with empty field name")
		}
		if len(values) == 0 {
			panic("can't use WithWhereEq without at least single value")
		}
		if opts.fieldEqName != "" {
			panic("can't use WithWhereEq more then one time")
		}

		opts.fieldEqName = name
		opts.fieldEqValues = values
	}
}

func WithGetFirst() FindOpt {
	return func(opts *FindOpts) {
		if opts.key == "" {
			panic("can't use WithGetFirst without WithKey (key isn't set)")
		}
		if opts.gen != 0 {
			panic("can't use WithGetFirst when WithGen already used")
		}
		if opts.getLast {
			panic("can't use WithGetFirst when WithGetLast already used")
		}
		if opts.getFirst {
			panic("can't use WithGetFirst more then one time")
		}

		opts.getFirst = true
	}
}

func WithGetLast() FindOpt {
	return func(opts *FindOpts) {
		if opts.key == "" {
			panic("can't use WithGetLast without WithKey (key isn't set)")
		}
		if opts.gen != 0 {
			panic("can't use WithGetLast when WithGen already used")
		}
		if opts.getFirst {
			panic("can't use WithGetLast when WithGetFirst already used")
		}
		if opts.getLast {
			panic("can't use WithGetLast more then one time")
		}

		opts.getLast = true
	}
}
