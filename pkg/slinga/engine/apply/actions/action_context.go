package actions

import (
	"github.com/Aptomi/aptomi/pkg/slinga/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/slinga/eventlog"
	"github.com/Aptomi/aptomi/pkg/slinga/external"
	"github.com/Aptomi/aptomi/pkg/slinga/language"
	"github.com/Aptomi/aptomi/pkg/slinga/plugin"
)

type ActionContext struct {
	DesiredPolicy *language.Policy
	DesiredState  *resolve.PolicyResolution
	ActualPolicy  *language.Policy
	ActualState   *resolve.PolicyResolution
	ExternalData  *external.Data
	Plugins       plugin.Registry
	EventLog      *eventlog.EventLog
}

func NewActionContext(desiredPolicy *language.Policy, desiredState *resolve.PolicyResolution, actualPolicy *language.Policy, actualState *resolve.PolicyResolution, externalData *external.Data, plugins plugin.Registry, eventLog *eventlog.EventLog) *ActionContext {
	return &ActionContext{
		DesiredPolicy: desiredPolicy,
		DesiredState:  desiredState,
		ActualPolicy:  actualPolicy,
		ActualState:   actualState,
		ExternalData:  externalData,
		Plugins:       plugins,
		EventLog:      eventLog,
	}
}
