package plugin

import (
	"github.com/Aptomi/aptomi/pkg/slinga/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/slinga/eventlog"
	"github.com/Aptomi/aptomi/pkg/slinga/external"
	lang "github.com/Aptomi/aptomi/pkg/slinga/language"
)

type ClustersPostProcessPlugin interface {
	Plugin

	Process(desiredPolicy *lang.PolicyNamespace, desiredState *resolve.PolicyResolution, externalData *external.Data, eventLog *eventlog.EventLog) error
}
