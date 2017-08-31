package base

import (
	"github.com/Aptomi/aptomi/pkg/slinga/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/slinga/eventlog"
)

type BasePlugin struct {
	Next     *resolve.Revision
	Prev     *resolve.Revision
	EventLog *eventlog.EventLog
}

func (basePlugin *BasePlugin) Init(next *resolve.Revision, prev *resolve.Revision) {
	basePlugin.Next = next
	basePlugin.Prev = prev
}

func (basePlugin *BasePlugin) OnApplyStart(eventLog *eventlog.EventLog) error {
	basePlugin.EventLog = eventLog
	return nil
}
