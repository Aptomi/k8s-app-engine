package etcd

import (
	"fmt"
	"time"

	"github.com/Aptomi/aptomi/pkg/runtime"
	"github.com/Aptomi/aptomi/pkg/runtime/store"
	etcd "github.com/coreos/etcd/clientv3"
)

type etcdStore struct {
	client *etcd.Client
	codec  store.Codec
}

func New( /* todo config */ codec store.Codec) (store.Interface, error) {
	client, err := etcd.New(etcd.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("error while connecting to etcd: %s", err)
	}

	// todo run compactor?

	return &etcdStore{
		client: client,
		codec:  codec,
	}, nil
}

func (s *etcdStore) Close() error {
	return s.client.Close()
}

// todo need to rework keys to not include kind or to start with kind at least

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
