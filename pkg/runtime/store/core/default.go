package core

import (
	"github.com/Aptomi/aptomi/pkg/engine/actual"
	"github.com/Aptomi/aptomi/pkg/runtime/store"
	"sync"
)

// defaultStore is the generic store implementation that is the glue layer for saving
// different engine objects into the object store
type defaultStore struct {
	policyChangeLock   sync.Mutex
	store              store.Generic
	actualStateUpdater actual.StateUpdater
}

// NewStore returns default implementation of generic store
func NewStore(store store.Generic) store.Core {
	return &defaultStore{
		store:              store,
		actualStateUpdater: &actualStateUpdater{store: store},
	}
}

func (ds *defaultStore) GetActualStateUpdater() actual.StateUpdater {
	return ds.actualStateUpdater
}
