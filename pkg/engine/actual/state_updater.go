package actual

import "github.com/Aptomi/aptomi/pkg/object"

// StateUpdater is an interface to process changes in actual state, which get triggered from actions in state applier.
// When a new object gets created, changed or updated, state updater will persist those changes in the underlying store
type StateUpdater interface {
	// Create will get called when a new object (ComponentInstance) is created in the actual state
	Create(obj object.Base) error

	// Update will get called when an existing object (ComponentInstance) is changed in the actual state
	Update(obj object.Base) error

	// Delete will get called when an existing object (ComponentInstance) is deleted from  the actual state
	Delete(string) error
}
