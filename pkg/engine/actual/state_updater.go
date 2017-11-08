package actual

import (
	"github.com/Aptomi/aptomi/pkg/runtime"
)

// StateUpdater is an interface to process changes in actual state, which get triggered from actions in state applier.
// When a new object gets created, changed or updated, state updater will persist those changes in the underlying store.
type StateUpdater interface {
	// Save will get called when a new object need to be created or existing object is changed in the actual state
	Save(obj runtime.Storable) error

	// Delete will get called when an existing object (ComponentInstance) is deleted from  the actual state
	Delete(string) error
}
