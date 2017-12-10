package core

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/engine/actual"
	"github.com/Aptomi/aptomi/pkg/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/runtime"
	"github.com/Aptomi/aptomi/pkg/runtime/store"
)

func (ds *defaultStore) GetActualState() (*resolve.PolicyResolution, error) {
	actualState := resolve.NewPolicyResolution(false)

	instances, err := ds.store.List(runtime.KeyFromParts(runtime.SystemNS, resolve.ComponentInstanceObject.Kind, ""))
	if err != nil {
		return nil, fmt.Errorf("error while getting all component instances: %s", err)
	}

	for _, instanceObj := range instances {
		if instance, ok := instanceObj.(*resolve.ComponentInstance); ok {
			key := instance.GetKey()
			actualState.ComponentInstanceMap[key] = instance
		}
	}

	return actualState, nil
}

func (ds *defaultStore) GetActualStateUpdater() actual.StateUpdater {
	return &actualStateUpdater{ds.store}
}

type actualStateUpdater struct {
	store store.Generic
}

func (updater *actualStateUpdater) Save(obj runtime.Storable) error {
	if _, ok := obj.(*resolve.ComponentInstance); !ok {
		return fmt.Errorf("only ComponentInstances could be updated using actual.StateUpdater, not: %T", obj)
	}

	_, err := updater.store.Save(obj)
	return err
}

// Delete is used for reacting on object delete event (not supported for now)
func (updater *actualStateUpdater) Delete(key string) error {
	return updater.store.Delete(key)
}

func (ds *defaultStore) ResetActualState() error {
	instances, err := ds.store.List(runtime.KeyFromParts(runtime.SystemNS, resolve.ComponentInstanceObject.Kind, ""))
	if err != nil {
		return fmt.Errorf("error while getting all component instances: %s", err)
	}

	for _, instanceObj := range instances {
		if instance, ok := instanceObj.(*resolve.ComponentInstance); ok {
			key := runtime.KeyForStorable(instance)
			deleteErr := ds.store.Delete(key)
			if deleteErr != nil {
				return deleteErr
			}
		}
	}

	return nil
}
