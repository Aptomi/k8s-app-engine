package plugin

import (
	"github.com/Aptomi/aptomi/pkg/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/event"
	"github.com/Aptomi/aptomi/pkg/external"
	"github.com/Aptomi/aptomi/pkg/lang"
)

// PostProcessPlugin is a definition of post-processing plugin which gets called once by an action from the engine
// applier, after engine is done processing all component instances
type PostProcessPlugin interface {
	Plugin

	Process(desiredPolicy *lang.Policy, desiredState *resolve.PolicyResolution, externalData *external.Data, eventLog *event.Log) error
}
