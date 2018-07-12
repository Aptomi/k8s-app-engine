package store

// SaveOpt is a function that changes object save process options
type SaveOpt func(opts *SaveOpts)

// SaveOpts is a list of object save process options
type SaveOpts struct {
	replaceOrForceGen bool
}

// IsReplaceOrForceGen returns true if an existing object should be replaced or it should be saved with specific revision
func (opts *SaveOpts) IsReplaceOrForceGen() bool {
	return opts.replaceOrForceGen
}

// NewSaveOpts creates SaveOpts (object save process config) from list of SaveOpt (object save process config modifiers)
func NewSaveOpts(opts []SaveOpt) *SaveOpts {
	saveOpts := &SaveOpts{}
	for _, opt := range opts {
		opt(saveOpts)
	}

	return saveOpts
}

// WithReplaceOrForceGen is object save process modifier for an existing object to be replaced or be saved with specific revision
func WithReplaceOrForceGen() SaveOpt {
	return func(opts *SaveOpts) {
		if opts.replaceOrForceGen {
			panic("can't use WithReplaceOrForceGen more then one time")
		}

		opts.replaceOrForceGen = true
	}
}
