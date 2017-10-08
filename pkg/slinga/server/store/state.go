package store

import (
	"github.com/Aptomi/aptomi/pkg/slinga/engine/actual"
	"github.com/Aptomi/aptomi/pkg/slinga/object"
	"github.com/Aptomi/aptomi/pkg/slinga/object/store"
)

type defaultStateUpdater struct {
	store store.ObjectStore
}

func (s *defaultStore) ActualStateUpdater() actual.StateUpdater {
	return &defaultStateUpdater{s.store}
}

func (u *defaultStateUpdater) Create(obj object.Base) error {
	// todo allowed only for component instances == actual state
	//todo impl
	return nil
}

func (u *defaultStateUpdater) Update(obj object.Base) error {
	// todo allowed only for component instances == actual state
	//todo impl
	return nil
}

func (u *defaultStateUpdater) Delete(string) error {
	// todo allowed only for component instances == actual state
	//todo impl
	return nil
}
