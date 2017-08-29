package plugin

import (
	"github.com/Aptomi/aptomi/pkg/slinga/engine/progress"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/resolve"
)

// EngineDiffPlugin contains methods which will be called by the engine during diff calculation
type EngineDiffPlugin interface {
	// Init will be called by the engine after diff is calculated and populated with data
	Init(next *resolve.ResolvedState, prev *resolve.ResolvedState)

	// GetApplyProgressLength should return the number of times a plugin
	// will increment progress indicator during Apply() phase
	GetApplyProgressLength() int
}

// EngineApplyPlugin contains methods which will be called by the engine during diff calculation
type EngineApplyPlugin interface {
	// Apply method will be called with progress indicator as a parameter
	// It should be advanced by GetApplyProgressLength() steps throughout execution of the plugin
	Apply(progress progress.ProgressIndicator)
}

// EnginePlugin contains all plugin methods combined
type EnginePlugin interface {
	EngineDiffPlugin
	EngineApplyPlugin
}
