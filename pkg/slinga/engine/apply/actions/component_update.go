package actions

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/plugin"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/slinga/eventlog"
)

type ComponentUpdate struct {
	*ComponentBaseAction
}

func NewComponentUpdateAction(key string, desiredState *resolve.PolicyResolution, actualState *resolve.PolicyResolution) *ComponentUpdate {
	return &ComponentUpdate{ComponentBaseAction: NewComponentBaseAction(key, desiredState, actualState)}
}

func (componentUpdate *ComponentUpdate) Apply(plugins []plugin.EnginePlugin, eventLog *eventlog.EventLog) error {
	// Process updates in the right order
	foundErrors := false

	// call plugins to perform their actions
	for _, pluginInstance := range plugins {
		err := pluginInstance.OnApplyComponentInstanceUpdate(componentUpdate.key)
		if err != nil {
			eventLog.LogError(err)
			foundErrors = true
		}
	}
	if foundErrors {
		return fmt.Errorf("One or more errors while applying changes (updating component '%s')", componentUpdate.key)
	}

	// update actual state
	componentUpdate.actualState.Resolved.ComponentInstanceMap[componentUpdate.key] = componentUpdate.desiredState.Resolved.ComponentInstanceMap[componentUpdate.key]

	return nil
}
