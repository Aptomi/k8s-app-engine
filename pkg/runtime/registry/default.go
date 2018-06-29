package registry

import (
	"sync"

	"github.com/Aptomi/aptomi/pkg/engine/actual"
	"github.com/Aptomi/aptomi/pkg/engine/resolve"
)

// defaultStore is the generic store implementation that is the glue layer for saving
// different engine objects into the object store
type defaultStore struct {
	policyChangeLock sync.Mutex
	store            store.Generic
}

// NewStore returns default implementation of generic store
func NewStore(store store.Generic) store.Interface {
	return &defaultStore{
		store: store,
	}
}
