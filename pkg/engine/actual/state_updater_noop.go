package actual

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/engine/resolve"
	"sync"
	"time"
)

type noOpActualStateUpdater struct {
	mutex       sync.Mutex
	actualState *resolve.PolicyResolution
}

// NewNoOpActionStateUpdater creates a mock state updater for unit tests, which does not have an underlying object store to save changes to
func NewNoOpActionStateUpdater(actualState *resolve.PolicyResolution) StateUpdater {
	return &noOpActualStateUpdater{
		actualState: actualState,
	}
}

// GetComponentInstance returns component instance by key
func (updater *noOpActualStateUpdater) GetComponentInstance(key string) *resolve.ComponentInstance {
	// make sure all changes to actual state are synchronized
	updater.mutex.Lock()
	defer updater.mutex.Unlock()

	return updater.actualState.ComponentInstanceMap[key]
}

// CreateComponentInstance creates a new component instance in the actual state
func (updater *noOpActualStateUpdater) CreateComponentInstance(instance *resolve.ComponentInstance) error {
	// make sure all changes to actual state are synchronized
	updater.mutex.Lock()
	defer updater.mutex.Unlock()

	// update timestamps
	instance.CreatedAt = time.Now()
	instance.UpdatedAt = time.Now()

	// move it over to the actual state
	updater.actualState.ComponentInstanceMap[instance.GetKey()] = instance

	return nil
}

// UpdateComponentInstance updates an existing component instance in the actual state
func (updater *noOpActualStateUpdater) UpdateComponentInstance(key string, makeChanges func(instance *resolve.ComponentInstance)) error {
	// make sure all changes to actual state are synchronized
	updater.mutex.Lock()
	defer updater.mutex.Unlock()

	// load instance from the store
	instance, ok := updater.actualState.ComponentInstanceMap[key]
	if !ok {
		return fmt.Errorf("instance not found in actual state")
	}

	// update timestamp
	instance.UpdatedAt = time.Now()

	// update component instance
	makeChanges(instance)

	return nil
}

// DeleteComponentInstance deletes an existing component instance from the actual state
func (updater *noOpActualStateUpdater) DeleteComponentInstance(key string) error {
	// make sure all changes to actual state are synchronized
	updater.mutex.Lock()
	defer updater.mutex.Unlock()

	// delete an existing component from the actual state map
	delete(updater.actualState.ComponentInstanceMap, key)

	return nil
}

// GetUpdatedActualState returns the updated actual state
func (updater *noOpActualStateUpdater) GetUpdatedActualState() *resolve.PolicyResolution {
	return updater.actualState
}
