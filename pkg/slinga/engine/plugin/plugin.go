package plugin

import (
	"github.com/Aptomi/aptomi/pkg/slinga/engine/progress"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/slinga/eventlog"
	"github.com/Aptomi/aptomi/pkg/slinga/language"
)

// EngineDiffPlugin contains methods which will be called by the engine during diff calculation
type EngineDiffPlugin interface {
}

// EngineApplyPlugin contains methods which will be called by the engine during apply phase
type EngineApplyPlugin interface {
	// Init will be called by the engine when apply starts
	Init(desiredPolicy *language.PolicyNamespace, desiredState *resolve.PolicyResolution, actualPolicy *language.PolicyNamespace, actualState *resolve.PolicyResolution, userLoader language.UserLoader, eventLog *eventlog.EventLog)

	// OnApplyComponentInstanceCreate will be called by the engine when a new component instance has to be instantiated
	OnApplyComponentInstanceCreate(string) error

	// OnApplyComponentInstanceUpdate will be called by the engine when an existing component instance has to be updated
	OnApplyComponentInstanceUpdate(string) error

	// OnApplyComponentInstanceDelete will be called by the engine when an existing component instance has to be deleted
	OnApplyComponentInstanceDelete(string) error

	// GetCustomApplyProgressLength should return the number of times a plugin
	// will increment progress indicator during OnApplyCustom() phase
	GetCustomApplyProgressLength() int

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
