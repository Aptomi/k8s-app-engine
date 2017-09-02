package actions

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/plugin"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/slinga/eventlog"
)

type ComponentDelete struct {
	*ComponentBaseAction
}

func NewComponentDeleteAction(key string, desiredState *resolve.PolicyResolution, actualState *resolve.PolicyResolution) *ComponentDelete {
	return &ComponentDelete{ComponentBaseAction: NewComponentBaseAction(key, desiredState, actualState)}
}

func (componentDelete *ComponentDelete) Apply(plugins []plugin.EnginePlugin, eventLog *eventlog.EventLog) error {
	// Process destructions in the right order
	foundErrors := false

	// call plugins to perform their actions
	for _, pluginInstance := range plugins {
		err := pluginInstance.OnApplyComponentInstanceDelete(componentDelete.key)
		if err != nil {
			eventLog.LogError(err)
			foundErrors = true
		}
	}
	if foundErrors {
		return fmt.Errorf("One or more errors while applying changes (deleting component '%s')", componentDelete.key)
	}

	componentDelete.updateActualState()

	return nil
}

func (componentDelete *ComponentDelete) updateActualState() {
	// delete component from the actual state
	delete(componentDelete.actualState.Resolved.ComponentInstanceMap, componentDelete.key)
}
