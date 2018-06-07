package actual

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/engine/resolve"
	"sync"
	"time"
)

// NewNoOpActionStateUpdater creates a mock state updater for unit tests, which does not have an underlying object store to save changes to
func NewNoOpActionStateUpdater() StateUpdater {
	return &noOpActualStateUpdater{}
}

type noOpActualStateUpdater struct {
	mutex sync.Mutex
}

func (updater *noOpActualStateUpdater) CreateComponentInstance(instance *resolve.ComponentInstance, actualState *resolve.PolicyResolution) error {
	// make sure all changes to actual state are synchronized
	updater.mutex.Lock()
	defer updater.mutex.Unlock()

	// update timestamps
	instance.CreatedAt = time.Now()
	instance.UpdatedAt = time.Now()

	// move it over to the actual state
	actualState.ComponentInstanceMap[instance.GetKey()] = instance

	return nil
}

func (updater *noOpActualStateUpdater) UpdateComponentInstance(key string, actualState *resolve.PolicyResolution, makeChanges func(instance *resolve.ComponentInstance)) error {
	// make sure all changes to actual state are synchronized
	updater.mutex.Lock()
	defer updater.mutex.Unlock()

	// load instance from the store
	instance, ok := actualState.ComponentInstanceMap[key]
	if !ok {
		return fmt.Errorf("instance not found in actual state")
	}

	// update timestamp
	instance.UpdatedAt = time.Now()

	// update component instance
	makeChanges(instance)

	return nil
}

func (updater *noOpActualStateUpdater) DeleteComponentInstance(key string, actualState *resolve.PolicyResolution) error {
	// make sure all changes to actual state are synchronized
	updater.mutex.Lock()
	defer updater.mutex.Unlock()

	// delete an existing component from the actual state map
	delete(actualState.ComponentInstanceMap, key)

	return nil
}
