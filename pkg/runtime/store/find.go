package store

import (
	"github.com/Aptomi/aptomi/pkg/runtime"
)

// FindOpt is a function that changes object find process options
type FindOpt func(opts *FindOpts)

// FindOpts is a list of object find process options
type FindOpts struct {
	keyPrefix     runtime.Key
	key           runtime.Key
	gen           runtime.Generation
	fieldEqName   string
	fieldEqValues []interface{}
	getLast       bool
	getFirst      bool
}

// GetKeyPrefix returns key prefix to find objects with keys prefixed by it
func (opts *FindOpts) GetKeyPrefix() runtime.Key {
	return opts.keyPrefix
}

// GetKey returns key to find objects with it
func (opts *FindOpts) GetKey() runtime.Key {
	return opts.key
}

// GetGen returns generation for find objects with specified key and this generation
func (opts *FindOpts) GetGen() runtime.Generation {
	return opts.gen
}

// GetFieldEqName returns name of the field to find object with this field equal to some value
func (opts *FindOpts) GetFieldEqName() string {
	return opts.fieldEqName
}

// GetFieldEqValues returns values for the specified field to find object with field equal to at least one of this values
func (opts *FindOpts) GetFieldEqValues() []interface{} {
	return opts.fieldEqValues
}

// IsGetFirst returns true if first result should be returned
func (opts *FindOpts) IsGetFirst() bool {
	return opts.getFirst
}

// IsGetLast returns true if last result should be returned
func (opts *FindOpts) IsGetLast() bool {
	return opts.getLast
}

// NewFindOpts creates FindOpts (object find process config) from list of FindOpt (object find process config modifiers)
func NewFindOpts(opts []FindOpt) *FindOpts {
	findOpts := &FindOpts{}
	for _, opt := range opts {
		opt(findOpts)
	}

	return findOpts
}

// WithKey defines key to find objects with it
func WithKey(key runtime.Key) FindOpt {
	return func(opts *FindOpts) {
		if opts.key != "" {
			panic("can't use WithKey more then one time")
		}

		opts.key = key
	}
}

// WithKeyPrefix defines key prefix to find objects with keys prefixed with it
func WithKeyPrefix(keyPrefix runtime.Key) FindOpt {
	return func(opts *FindOpts) {
		if opts.key != "" {
			panic("can't use WithKeyPrefix with key specified")
		}
		if opts.keyPrefix != "" {
			panic("can't use WithKeyPrefix more then one time")
		}

		opts.keyPrefix = keyPrefix
	}
}

// WithGen defines generation to find object with it
func WithGen(gen runtime.Generation) FindOpt {
	return func(opts *FindOpts) {
		if opts.key == "" {
			panic("can't use WithGen without WithKey (key isn't set)")
		}
		if opts.keyPrefix != "" {
			panic("can't use WithGen with key prefix specified")
		}
		if opts.gen != 0 {
			panic("can't use WithGen more then one time")
		}

		opts.gen = gen
	}
}

// WithWhereEq defines field name and values to find objects with this field equals to at least one of the specified values
func WithWhereEq(name string, values ...interface{}) FindOpt {
	return func(opts *FindOpts) {
		if name == "" {
			panic("can't use WithWhereEq with empty field name")
		}
		if len(values) == 0 {
			panic("can't use WithWhereEq without at least single value")
		}
		if opts.key == "" {
			panic("can't use WithWhereEq without specified key (it's only for searching generations now)")
		}
		if opts.keyPrefix != "" {
			panic("can't use WithWhereEq with key prefix specified (it's only for searching generations now)")
		}
		if opts.fieldEqName != "" {
			panic("can't use WithWhereEq more then one time")
		}

		opts.fieldEqName = name
		opts.fieldEqValues = values
	}
}

// WithGetFirst defines that first result should be returned
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

// WithGetLast defines that last result should be returned
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
