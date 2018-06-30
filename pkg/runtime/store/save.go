package store

import (
	"github.com/Aptomi/aptomi/pkg/runtime"
)

type SaveOpt func(opts *SaveOpts)

type SaveOpts struct {
	replace  bool
	forceGen runtime.Generation
}

func (opts *SaveOpts) IsReplace() bool {
	return opts.replace
}

func (opts *SaveOpts) GetForceGen() runtime.Generation {
	return opts.forceGen
}

func NewSaveOpts(opts []SaveOpt) *SaveOpts {
	saveOpts := &SaveOpts{}
	for _, opt := range opts {
		opt(saveOpts)
	}

	return saveOpts
}

func WithReplace() SaveOpt {
	return func(opts *SaveOpts) {
		if opts.forceGen != 0 {
			panic("can't use WithReplace when WithForceGen already used (as it's implicitly allows replacement)")
		}
		if opts.replace {
			panic("can't use WithReplace more then one time")
		}

		opts.replace = true
	}
}

func WithForceGen(gen runtime.Generation) SaveOpt {
	return func(opts *SaveOpts) {
		if opts.replace {
			panic("can't use WithForceGen when WithReplace already used (as it's implicitly allows replacement)")
		}
		if opts.forceGen != 0 {
			panic("can't use WithForceGen more then one time")
		}

		opts.forceGen = gen
	}
}
