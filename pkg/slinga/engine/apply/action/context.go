package action

import (
	"github.com/Aptomi/aptomi/pkg/slinga/engine/actual"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/slinga/eventlog"
	"github.com/Aptomi/aptomi/pkg/slinga/external"
	"github.com/Aptomi/aptomi/pkg/slinga/lang"
	"github.com/Aptomi/aptomi/pkg/slinga/plugin"
)

type Context struct {
	DesiredPolicy      *lang.Policy
	DesiredState       *resolve.PolicyResolution
	ActualPolicy       *lang.Policy
	ActualState        *resolve.PolicyResolution
	ActualStateUpdater actual.StateUpdater
	ExternalData       *external.Data
	Plugins            plugin.Registry
	EventLog           *eventlog.EventLog
}

func NewContext(desiredPolicy *lang.Policy, desiredState *resolve.PolicyResolution, actualPolicy *lang.Policy,
	actualState *resolve.PolicyResolution, actualStateUpdater actual.StateUpdater, externalData *external.Data,
	plugins plugin.Registry, eventLog *eventlog.EventLog) *Context {

	return &Context{
		DesiredPolicy:      desiredPolicy,
		DesiredState:       desiredState,
		ActualPolicy:       actualPolicy,
		ActualState:        actualState,
		ActualStateUpdater: actualStateUpdater,
		ExternalData:       externalData,
		Plugins:            plugins,
		EventLog:           eventLog,
	}
}
