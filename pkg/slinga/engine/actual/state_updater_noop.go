package actual

import "github.com/Aptomi/aptomi/pkg/slinga/object"

// NewNoOpActionStateUpdater creates a mock state updater which does nothing (useful in unit tests)
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
