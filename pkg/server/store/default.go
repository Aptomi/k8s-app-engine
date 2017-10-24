package store

import (
	"github.com/Aptomi/aptomi/pkg/object/store"
	"sync"
)

// DefaultStore is the ServerStore implementation that is the glue layer for saving
// different engine objects into the object store
type DefaultStore struct {
	policyUpdate sync.Mutex
	store        store.ObjectStore
}

// New returns default implementation of ServerStore
func New(store store.ObjectStore) ServerStore {
	return &DefaultStore{sync.Mutex{}, store}
}

// Object returns backing ObjectStore instance to work directly with it
func (s *DefaultStore) Object() store.ObjectStore {
	return s.store
}
