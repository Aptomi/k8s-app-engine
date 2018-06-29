package registry

import (
	"sync"

	"github.com/Aptomi/aptomi/pkg/runtime/store"
)

// defaultRegistry is the generic store implementation that is the glue layer for saving
// different engine objects into the object store
type defaultRegistry struct {
	policyChangeLock sync.Mutex
	store            store.Interface
}

// New returns default implementation of generic store
func New(store store.Interface) Interface {
	return &defaultRegistry{
		store: store,
	}
}
