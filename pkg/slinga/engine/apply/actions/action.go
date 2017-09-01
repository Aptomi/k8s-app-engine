package actions

import (
	"github.com/Aptomi/aptomi/pkg/slinga/engine/plugin"
	"github.com/Aptomi/aptomi/pkg/slinga/eventlog"
)

type Action interface {
	Apply(plugins []plugin.EnginePlugin, eventLog *eventlog.EventLog) error
}
