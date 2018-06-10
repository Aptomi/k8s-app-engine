package actual

import (
	"github.com/Aptomi/aptomi/pkg/engine/resolve"
)

// StateUpdater is an interface to process changes in actual state
type StateUpdater interface {
	GetComponentInstance(string) *resolve.ComponentInstance
	CreateComponentInstance(*resolve.ComponentInstance) error
	UpdateComponentInstance(string, func(instance *resolve.ComponentInstance)) error
	DeleteComponentInstance(string) error
	GetUpdatedActualState() *resolve.PolicyResolution
}
