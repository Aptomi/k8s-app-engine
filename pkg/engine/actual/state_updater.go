package actual

import (
	"github.com/Aptomi/aptomi/pkg/engine/resolve"
)

// StateUpdater is an interface to process changes in actual state
type StateUpdater interface {
	CreateComponentInstance(*resolve.ComponentInstance, *resolve.PolicyResolution) error
	UpdateComponentInstance(string, *resolve.PolicyResolution, func(instance *resolve.ComponentInstance)) error
	DeleteComponentInstance(string, *resolve.PolicyResolution) error
}
