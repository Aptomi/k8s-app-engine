package base

import (
	"github.com/Aptomi/aptomi/pkg/slinga/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/slinga/eventlog"
)

type BasePlugin struct {
	Next     *resolve.ResolvedState
	Prev     *resolve.ResolvedState
	EventLog *eventlog.EventLog
}

func (basePlugin *BasePlugin) Init(next *resolve.ResolvedState, prev *resolve.ResolvedState) {
	basePlugin.Next = next
	basePlugin.Prev = prev
}

func (basePlugin *BasePlugin) OnApplyStart(eventLog *eventlog.EventLog) error {
	basePlugin.EventLog = eventLog
	return nil
}
