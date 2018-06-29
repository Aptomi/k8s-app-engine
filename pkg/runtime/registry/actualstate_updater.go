package registry

import (
	"fmt"
	"sync"
	"time"

	"github.com/Aptomi/aptomi/pkg/engine/actual"
	"github.com/Aptomi/aptomi/pkg/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/runtime"
	"github.com/Aptomi/aptomi/pkg/runtime/store"
)

func (ds *defaultStore) NewActualStateUpdater(actualState *resolve.PolicyResolution) actual.StateUpdater {
	return &actualStateUpdater{
		store:       ds.store,
		actualState: actualState,
	}
}

type actualStateUpdater struct {
	store       registry.Generic
	mutex       sync.Mutex
	actualState *resolve.PolicyResolution
}

// GetComponentInstance returns component instance by key from the underlying store
func (updater *actualStateUpdater) GetComponentInstance(key string) *resolve.ComponentInstance {
	// make sure all changes to actual state are synchronized
	updater.mutex.Lock()
	defer updater.mutex.Unlock()

	return updater.actualState.ComponentInstanceMap[key]
}

// CreateComponentInstance creates a new component instance in the actual state, as well as in the underlying store (with appropriate synchronization)
func (updater *actualStateUpdater) CreateComponentInstance(instance *resolve.ComponentInstance) error {
	// make sure all changes to actual state are synchronized
	updater.mutex.Lock()
	defer updater.mutex.Unlock()

	// update timestamps
	instance.CreatedAt = time.Now()
	instance.UpdatedAt = time.Now()

	// save component instance in the actual state store
	err := updater.save(instance)
	if err != nil {
		return err
	}

	// move it over to the actual state
	updater.actualState.ComponentInstanceMap[instance.GetKey()] = instance

	return nil
}

// UpdateComponentInstance updates an existing component instance in the actual state, as well as in the underlying store by calling function makeChanges (with appropriate synchronization)
func (updater *actualStateUpdater) UpdateComponentInstance(key string, makeChanges func(instance *resolve.ComponentInstance)) error {
	// make sure all changes to actual state are synchronized
	updater.mutex.Lock()
	defer updater.mutex.Unlock()

	// load instance from the store
	instance, err := updater.loadComponentInstance(key)
	if err != nil {
		return err
	}

	// update timestamp
	instance.UpdatedAt = time.Now()

	// update component instance
	makeChanges(instance)

	// save component instance in the actual state store
	err = updater.save(instance)
	if err != nil {
		return err
	}

	// move it over to the actual state
	updater.actualState.ComponentInstanceMap[instance.GetKey()] = instance

	return nil
}

// DeleteComponentInstance deletes an existing component instance from the actual state, as well as from the underlying store (with appropriate synchronization)
func (updater *actualStateUpdater) DeleteComponentInstance(key string) error {
	// make sure all changes to actual state are synchronized
	updater.mutex.Lock()
	defer updater.mutex.Unlock()

	// delete an existing component from the actual state store
	err := updater.delete(storableKeyForComponent(key))
	if err != nil {
		return err
	}

	// delete an existing component from the actual state map
	delete(updater.actualState.ComponentInstanceMap, key)

	return nil
}

// GetUpdatedActualState returns the updated actual state
func (updater *actualStateUpdater) GetUpdatedActualState() *resolve.PolicyResolution {
	return updater.actualState
}

func storableKeyForComponent(componentKey string) string {
	return runtime.KeyFromParts(runtime.SystemNS, resolve.ComponentInstanceObject.Kind, componentKey)
}

func (updater *actualStateUpdater) loadComponentInstance(key string) (*resolve.ComponentInstance, error) {
	obj, err := updater.store.Get(storableKeyForComponent(key))
	if err != nil {
		return nil, err
	}
	instance, ok := obj.(*resolve.ComponentInstance)
	if !ok {
		return nil, fmt.Errorf("tried to load component instance from the store, but loaded %v", instance)
	}
	return instance, nil
}

func (updater *actualStateUpdater) save(obj runtime.Storable) error {
	if _, ok := obj.(*resolve.ComponentInstance); !ok {
		return fmt.Errorf("only ComponentInstances could be updated using actual.StateUpdater, not: %T", obj)
	}

	_, err := updater.store.Save(obj)
	return err
}

func (updater *actualStateUpdater) delete(key string) error {
	return updater.store.Delete(key)
}
