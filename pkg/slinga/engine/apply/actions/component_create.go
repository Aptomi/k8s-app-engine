package actions

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/plugin"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/slinga/eventlog"
	"time"
)

type ComponentCreate struct {
	*ComponentBaseAction
}

func NewComponentCreateAction(key string, desiredState *resolve.PolicyResolution, actualState *resolve.PolicyResolution) *ComponentCreate {
	return &ComponentCreate{ComponentBaseAction: NewComponentBaseAction(key, desiredState, actualState)}
}

func (componentCreate *ComponentCreate) Apply(plugins []plugin.EnginePlugin, eventLog *eventlog.EventLog) error {
	// Process instantiations in the right order
	foundErrors := false

	// call plugins to perform their actions
	for _, pluginInstance := range plugins {
		err := pluginInstance.OnApplyComponentInstanceCreate(componentCreate.key)
		if err != nil {
			eventLog.LogError(err)
			foundErrors = true
		}
	}
	if foundErrors {
		return fmt.Errorf("One or more errors while applying changes (creating component '%s')", componentCreate.key)
	}

	componentCreate.updateActualState()

	return nil
}

func (componentCreate *ComponentCreate) updateActualState() {
	// get instance from desired state
	instance := componentCreate.desiredState.Resolved.ComponentInstanceMap[componentCreate.key]

	// copy it over to the actual state
	componentCreate.actualState.Resolved.ComponentInstanceMap[componentCreate.key] = instance

	// update creation and update times
	instance.UpdateTimes(time.Now(), time.Now())
}
