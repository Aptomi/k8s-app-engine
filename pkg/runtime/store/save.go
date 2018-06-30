package store

type SaveOpt func(opts *SaveOpts)

type SaveOpts struct {
	replaceOrForceGen bool
}

func (opts *SaveOpts) IsReplaceOrForceGen() bool {
	return opts.replaceOrForceGen
}

func NewSaveOpts(opts []SaveOpt) *SaveOpts {
	saveOpts := &SaveOpts{}
	for _, opt := range opts {
		opt(saveOpts)
	}

	return saveOpts
}

func WithReplaceOrForceGen() SaveOpt {
	return func(opts *SaveOpts) {
		if opts.replaceOrForceGen {
			panic("can't use WithReplaceOrForceGen more then one time")
		}

		opts.replaceOrForceGen = true
	}
}
