package store

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/engine/actual"
	"github.com/Aptomi/aptomi/pkg/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/object"
	"github.com/Aptomi/aptomi/pkg/object/store"
)

func (s *defaultStore) GetActualState() (*resolve.PolicyResolution, error) {
	// todo empty state temporarily
	actualState := resolve.NewPolicyResolution()

	instances, err := s.store.GetAll(object.SystemNS, resolve.ComponentInstanceObject.Kind)
	if err != nil {
		return nil, fmt.Errorf("Error while getting all component instances: %s", err)
	}

	for _, instanceObj := range instances {
		if instance, ok := instanceObj.(*resolve.ComponentInstance); ok {
			key := instance.GetKey()
			actualState.ComponentInstanceMap[key] = instance
			actualState.ComponentProcessingOrder = append(actualState.ComponentProcessingOrder, key)
		}
	}

	return actualState, nil
}

func (s *defaultStore) ActualStateUpdater() actual.StateUpdater {
	return &defaultStateUpdater{s.store}
}

type defaultStateUpdater struct {
	store store.ObjectStore
}

func (u *defaultStateUpdater) Create(obj object.Base) error {
	return u.Update(obj)
}

func (u *defaultStateUpdater) Update(obj object.Base) error {
	if _, ok := obj.(*resolve.ComponentInstance); !ok {
		return fmt.Errorf("Only ComponentInstances could be updated using actual.StateUpdater, not: %T", obj)
	}

	_, err := u.store.Save(obj)
	return err
}

func (u *defaultStateUpdater) Delete(string) error {
	// todo
	panic("not implemented: defaultStateUpdater.Delete")
}
