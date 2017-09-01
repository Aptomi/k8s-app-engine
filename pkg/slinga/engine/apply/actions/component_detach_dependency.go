package actions

import (
	"github.com/Aptomi/aptomi/pkg/slinga/engine/plugin"
	"github.com/Aptomi/aptomi/pkg/slinga/eventlog"
)

type ComponentDetachDependency struct {
	*ComponentBaseAction
	DependencyId string
}

func NewComponentDetachDependencyAction(key string, dependencyId string) *ComponentDetachDependency {
	return &ComponentDetachDependency{ComponentBaseAction: &ComponentBaseAction{Key: key}, DependencyId: dependencyId}
}

func (detachDependency *ComponentDetachDependency) Apply(plugins []plugin.EnginePlugin, eventLog *eventlog.EventLog) error {
	return nil
}
