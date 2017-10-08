package store

import (
	"github.com/Aptomi/aptomi/pkg/slinga/object/store"
	"sync"
)

type defaultStore struct {
	policyUpdate sync.Mutex
	store        store.ObjectStore
}

func New(store store.ObjectStore) ServerStore {
	return &defaultStore{sync.Mutex{}, store}
}

func (s *defaultStore) Object() store.ObjectStore {
	return s.store
}
