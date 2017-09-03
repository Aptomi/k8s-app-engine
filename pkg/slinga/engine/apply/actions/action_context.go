package actions

import (
	"github.com/Aptomi/aptomi/pkg/slinga/engine/plugin"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/slinga/eventlog"
	"github.com/Aptomi/aptomi/pkg/slinga/external"
	"github.com/Aptomi/aptomi/pkg/slinga/language"
)

type ActionContext struct {
	DesiredPolicy *language.PolicyNamespace
	DesiredState  *resolve.PolicyResolution
	ActualPolicy  *language.PolicyNamespace
	ActualState   *resolve.PolicyResolution
	ExternalData  *external.Data
	Plugins       *plugin.PluginRegistry
	EventLog      *eventlog.EventLog
}

func NewActionContext(desiredPolicy *language.PolicyNamespace, desiredState *resolve.PolicyResolution, actualPolicy *language.PolicyNamespace, actualState *resolve.PolicyResolution, externalData *external.Data, eventLog *eventlog.EventLog) *ActionContext {
	return &ActionContext{
		DesiredPolicy: desiredPolicy,
		DesiredState:  desiredState,
		ActualPolicy:  actualPolicy,
		ActualState:   actualState,
		ExternalData:  externalData,
		Plugins:       &plugin.PluginRegistry{},
		EventLog:      eventLog,
	}
}
