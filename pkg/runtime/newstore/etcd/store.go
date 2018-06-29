package etcd

import (
	"github.com/Aptomi/aptomi/pkg/runtime/newstore"
)

type store struct {
}

func New( /* todo config */ ) newstore.Interface {
	return &store{}
}

func (s *store) Save(storable newstore.Storable, opts ...newstore.SaveOpt) error {
	panic("implement me")
}

func (s *store) Find(kind newstore.Kind, opts ...newstore.FindOpt) newstore.Finder {
	panic("implement me")
}

func (s *store) Delete(kind newstore.Kind, opts ...newstore.DeleteOpt) newstore.Deleter {
	panic("implement me")
}
