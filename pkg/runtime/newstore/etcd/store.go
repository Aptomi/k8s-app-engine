package etcd

import (
	"github.com/Aptomi/aptomi/pkg/runtime/newstore"
)

type store struct {
}

func New( /* todo config */ ) newstore.Interface {
	// todo connect to db
	return &store{}
}

func (s *store) Save(storable newstore.Storable, opts ...newstore.SaveOpt) error {
	panic("implement me")
}

func (s *store) Find(kind newstore.Kind, opts ...newstore.FindOpt) newstore.Finder {
	// todo next 3 lines could be extracted to the common code
	findOpts := &newstore.FindOpts{}
	for _, opt := range opts {
		opt(findOpts)
	}

	return &finder{s, kind, findOpts}
}

func (s *store) Delete(kind newstore.Kind, opts ...newstore.DeleteOpt) newstore.Deleter {
	panic("implement me")
}

type finder struct {
	*store
	kind newstore.Kind
	*newstore.FindOpts
}

func (f *finder) First(newstore.Storable) error {
	panic("implement me")
}

func (f *finder) Last(newstore.Storable) error {
	panic("implement me")
}

func (f *finder) List([]newstore.Storable) error {
	panic("implement me")
}
