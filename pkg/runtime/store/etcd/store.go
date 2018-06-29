package etcd

import (
	"github.com/Aptomi/aptomi/pkg/runtime"
	"github.com/Aptomi/aptomi/pkg/runtime/store"
)

type etcdStore struct {
}

func New( /* todo config */ ) store.Interface {
	// todo connect to db
	return &etcdStore{}
}

func (s *etcdStore) Close() error {
	panic("implement me")
}

func (s *etcdStore) Save(storable runtime.Storable, opts ...store.SaveOpt) error {
	panic("implement me")
}

func (s *etcdStore) Find(kind runtime.Kind, opts ...store.FindOpt) store.Finder {
	// todo next 3 lines could be extracted to the common code
	findOpts := &store.FindOpts{}
	for _, opt := range opts {
		opt(findOpts)
	}

	return &finder{s, kind, findOpts}
}

func (s *etcdStore) Delete(kind runtime.Kind, opts ...store.DeleteOpt) store.Deleter {
	panic("implement me")
}

type finder struct {
	*etcdStore
	kind runtime.Kind
	*store.FindOpts
}

func (f *finder) First(runtime.Storable) error {
	panic("implement me")
}

func (f *finder) Last(runtime.Storable) error {
	panic("implement me")
}

func (f *finder) List(interface{}) error {
	panic("implement me")
}
