package actual

import "github.com/Aptomi/aptomi/pkg/object"

// NewNoOpActionStateUpdater creates a mock state updater for unit tests, which does nothing
func NewNoOpActionStateUpdater() StateUpdater {
	return &noOpActualStateUpdater{}
}

type noOpActualStateUpdater struct {
}

func (*noOpActualStateUpdater) Create(obj object.Base) error {
	return nil
}

func (*noOpActualStateUpdater) Update(obj object.Base) error {
	return nil
}

func (*noOpActualStateUpdater) Delete(string) error {
	return nil
}
