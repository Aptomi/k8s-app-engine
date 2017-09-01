package actions

import (
	"github.com/Aptomi/aptomi/pkg/slinga/engine/plugin"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/slinga/eventlog"
)

type ComponentAttachDependency struct {
	*ComponentBaseAction
	DependencyId string
}

func NewComponentAttachDependencyAction(key string, dependencyId string, desiredState *resolve.PolicyResolution, actualState *resolve.PolicyResolution) *ComponentAttachDependency {
	return &ComponentAttachDependency{ComponentBaseAction: NewComponentBaseAction(key, desiredState, actualState), DependencyId: dependencyId}
}

func (attachDependency *ComponentAttachDependency) Apply(plugins []plugin.EnginePlugin, eventLog *eventlog.EventLog) error {
	return nil
}
