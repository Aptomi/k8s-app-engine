package store

import (
	"github.com/Aptomi/aptomi/pkg/runtime"
)

type FindOpt func(opts *FindOpts)

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
