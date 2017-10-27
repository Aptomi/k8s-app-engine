package plugin

import (
	"github.com/Aptomi/aptomi/pkg/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/event"
	"github.com/Aptomi/aptomi/pkg/external"
	"github.com/Aptomi/aptomi/pkg/lang"
)

// ClustersPostProcessPlugin is a definition of post-processing plugin which gets called on each cluster after
// engine applier is done processing all component instances
type ClustersPostProcessPlugin interface {
	Plugin

	Process(desiredPolicy *lang.Policy, desiredState *resolve.PolicyResolution, externalData *external.Data, eventLog *event.Log) error
}
