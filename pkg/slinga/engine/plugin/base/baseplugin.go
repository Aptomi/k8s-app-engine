package base

import (
	"github.com/Aptomi/aptomi/pkg/slinga/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/slinga/eventlog"
	"github.com/Aptomi/aptomi/pkg/slinga/language"
)

type BasePlugin struct {
	Next *struct {
		Policy     *language.PolicyNamespace
		Resolution *resolve.PolicyResolution
	}
	Prev *struct {
		Policy     *language.PolicyNamespace
		Resolution *resolve.PolicyResolution
	}
	UserLoader language.UserLoader

	EventLog *eventlog.EventLog
}

func (basePlugin *BasePlugin) Init(nextPolicy *language.PolicyNamespace, nextResolution *resolve.PolicyResolution, prevPolicy *language.PolicyNamespace, prevResolution *resolve.PolicyResolution, userLoader language.UserLoader, log *eventlog.EventLog) {
	basePlugin.Next = &struct {
		Policy     *language.PolicyNamespace
		Resolution *resolve.PolicyResolution
	}{
		Policy:     nextPolicy,
		Resolution: nextResolution,
	}

	basePlugin.Prev = &struct {
		Policy     *language.PolicyNamespace
		Resolution *resolve.PolicyResolution
	}{
		Policy:     prevPolicy,
		Resolution: prevResolution,
	}

	basePlugin.EventLog = log
}
