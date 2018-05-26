package actual

import (
	"github.com/Aptomi/aptomi/pkg/runtime"
)

// NewNoOpActionStateUpdater creates a mock state updater for unit tests, which does not have an underlying object store to save changes to
func NewNoOpActionStateUpdater() StateUpdater {
	return &noOpActualStateUpdater{}
}

type noOpActualStateUpdater struct {
}

func (*noOpActualStateUpdater) Save(obj runtime.Storable) error {
	return nil
}

func (*noOpActualStateUpdater) Delete(string) error {
	return nil
}
