package actual

import "github.com/Aptomi/aptomi/pkg/slinga/object"

// StateUpdater is an interface to process actual state updates
// When a new object gets created, changed  or updated, it will give a signal to the storage layer to persist those changes
type StateUpdater interface {
	Create(obj object.Base) error
	Update(obj object.Base) error
	Delete(string) error
}
