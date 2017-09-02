package base

import (
	"github.com/Aptomi/aptomi/pkg/slinga/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/slinga/eventlog"
	"github.com/Aptomi/aptomi/pkg/slinga/external"
	"github.com/Aptomi/aptomi/pkg/slinga/language"
)

type BasePlugin struct {
	Desired *struct {
		Policy     *language.PolicyNamespace
		Resolution *resolve.PolicyResolution
	}
	Actual *struct {
		Policy     *language.PolicyNamespace
		Resolution *resolve.PolicyResolution
	}
	ExternalData *external.Data

	EventLog *eventlog.EventLog
}

func (basePlugin *BasePlugin) Init(desiredPolicy *language.PolicyNamespace, desiredState *resolve.PolicyResolution, actualPolicy *language.PolicyNamespace, actualState *resolve.PolicyResolution, externalData *external.Data, log *eventlog.EventLog) {
	basePlugin.Desired = &struct {
		Policy     *language.PolicyNamespace
		Resolution *resolve.PolicyResolution
	}{
		Policy:     desiredPolicy,
		Resolution: desiredState,
	}

	basePlugin.Actual = &struct {
		Policy     *language.PolicyNamespace
		Resolution *resolve.PolicyResolution
	}{
		Policy:     actualPolicy,
		Resolution: actualState,
	}

	basePlugin.ExternalData = externalData

	basePlugin.EventLog = log
}
