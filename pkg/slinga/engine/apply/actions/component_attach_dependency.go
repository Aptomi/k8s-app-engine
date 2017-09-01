package actions

import (
	"github.com/Aptomi/aptomi/pkg/slinga/engine/plugin"
	"github.com/Aptomi/aptomi/pkg/slinga/eventlog"
)

type ComponentAttachDependency struct {
	*ComponentBaseAction
	DependencyId string
}

func NewComponentAttachDependencyAction(key string, dependencyId string) *ComponentAttachDependency {
	return &ComponentAttachDependency{ComponentBaseAction: &ComponentBaseAction{Key: key}, DependencyId: dependencyId}
}

func (attachDependency *ComponentAttachDependency) Apply(plugins []plugin.EnginePlugin, eventLog *eventlog.EventLog) error {
	return nil
}
