package plugin

import (
	"github.com/Aptomi/aptomi/pkg/slinga/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/slinga/eventlog"
	lang "github.com/Aptomi/aptomi/pkg/slinga/language"
	"github.com/Aptomi/aptomi/pkg/slinga/external"
)

type ClustersPostProcessPlugin interface {
	Plugin

	Process(desiredPolicy *lang.PolicyNamespace, desiredState *resolve.PolicyResolution, externalData  *external.Data, eventLog *eventlog.EventLog) error
}
