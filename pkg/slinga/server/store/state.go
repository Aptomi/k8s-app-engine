package store

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/actual"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/slinga/object"
	"github.com/Aptomi/aptomi/pkg/slinga/object/store"
)

func (s *defaultStore) GetActualState() (*resolve.PolicyResolution, error) {
	// todo empty state temporarily
	actualState := resolve.NewPolicyResolution()

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
