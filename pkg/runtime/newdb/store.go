package newdb

type Store interface {
	Close() error

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
