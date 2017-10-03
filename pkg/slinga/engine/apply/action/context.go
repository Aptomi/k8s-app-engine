package action

import (
	"github.com/Aptomi/aptomi/pkg/slinga/engine/actual"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/slinga/event"
	"github.com/Aptomi/aptomi/pkg/slinga/external"
	"github.com/Aptomi/aptomi/pkg/slinga/lang"
	"github.com/Aptomi/aptomi/pkg/slinga/plugin"
)

// Context is a data struct that will be passed into all state update actions
// As actions need access to desired and actual data (policy, state), list of plugins, event log, etc
type Context struct {
	DesiredPolicy      *lang.Policy
	DesiredState       *resolve.PolicyResolution
	ActualPolicy       *lang.Policy
	ActualState        *resolve.PolicyResolution
	ActualStateUpdater actual.StateUpdater
	ExternalData       *external.Data
	Plugins            plugin.Registry
	EventLog           *event.Log
}

// NewContext creates a new instance of Context
func NewContext(desiredPolicy *lang.Policy, desiredState *resolve.PolicyResolution, actualPolicy *lang.Policy,
	actualState *resolve.PolicyResolution, actualStateUpdater actual.StateUpdater, externalData *external.Data,
	plugins plugin.Registry, eventLog *event.Log) *Context {

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
