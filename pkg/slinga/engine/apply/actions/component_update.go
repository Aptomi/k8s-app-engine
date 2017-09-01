package actions

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/plugin"
	"github.com/Aptomi/aptomi/pkg/slinga/eventlog"
)

type ComponentUpdate struct {
	*ComponentBaseAction
}

func NewComponentUpdateAction(key string) *ComponentUpdate {
	return &ComponentUpdate{ComponentBaseAction: &ComponentBaseAction{Key: key}}
}

func (update *ComponentUpdate) Apply(plugins []plugin.EnginePlugin, eventLog *eventlog.EventLog) error {
	// Process updates in the right order
	foundErrors := false

	// call plugins to perform their actions
	for _, pluginInstance := range plugins {
		err := pluginInstance.OnApplyComponentInstanceUpdate(update.Key)
		if err != nil {
			eventLog.LogError(err)
			foundErrors = true
		}
	}
	if foundErrors {
		return fmt.Errorf("One or more errors while applying changes (updating running components)")
	}
	return nil
}
