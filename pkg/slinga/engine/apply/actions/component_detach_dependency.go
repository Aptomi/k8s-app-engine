package actions

import (
	"github.com/Aptomi/aptomi/pkg/slinga/engine/plugin"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/slinga/eventlog"
)

type ComponentDetachDependency struct {
	*ComponentBaseAction
	DependencyId string
}

func NewComponentDetachDependencyAction(key string, dependencyId string, desiredState *resolve.PolicyResolution, actualState *resolve.PolicyResolution) *ComponentDetachDependency {
	return &ComponentDetachDependency{ComponentBaseAction: NewComponentBaseAction(key, desiredState, actualState), DependencyId: dependencyId}
}

func (detachDependency *ComponentDetachDependency) Apply(plugins []plugin.EnginePlugin, eventLog *eventlog.EventLog) error {
	return nil
}
