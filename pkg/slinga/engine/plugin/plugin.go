package plugin

import (
	"github.com/Aptomi/aptomi/pkg/slinga/engine/progress"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/slinga/eventlog"
)

// EngineDiffPlugin contains methods which will be called by the engine during diff calculation
type EngineDiffPlugin interface {
	// Init will be called by the engine after diff is calculated and populated with data
	Init(next *resolve.ResolvedState, prev *resolve.ResolvedState)

	// GetCustomApplyProgressLength should return the number of times a plugin
	// will increment progress indicator during OnApplyCustom() phase
	GetCustomApplyProgressLength() int
}

// EngineApplyPlugin contains methods which will be called by the engine during diff calculation
type EngineApplyPlugin interface {
	// OnApplyStart will be called by the engine when apply process starts, so a plugin can save a pointer to event log and write to it later
	OnApplyStart(*eventlog.EventLog) error

	// OnApplyComponentInstanceCreate will be called by the engine when a new component instance has to be instantiated
	OnApplyComponentInstanceCreate(*resolve.ComponentInstance) error

	// OnApplyComponentInstanceUpdate will be called by the engine when an existing component instance has to be updated
	OnApplyComponentInstanceUpdate(*resolve.ComponentInstance) error

	// OnApplyComponentInstanceDelete will be called by the engine when an existing component instance has to be deleted
	OnApplyComponentInstanceDelete(*resolve.ComponentInstance) error

	// OnApplyCustom method will be called after engine is done with processing all component instances
	// It will pass progress indicator as a parameter. Plugin should advance it by GetCustomApplyProgressLength() steps throughout
	// execution of OnApplyCustom method
	OnApplyCustom(progress.ProgressIndicator) error
}

// EnginePlugin contains all plugin methods combined
type EnginePlugin interface {
	EngineDiffPlugin
	EngineApplyPlugin
}
