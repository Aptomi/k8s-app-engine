package actions

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/plugin"
	"github.com/Aptomi/aptomi/pkg/slinga/eventlog"
)

type ComponentDelete struct {
	*ComponentBaseAction
}

func NewComponentDeleteAction(key string) *ComponentDelete {
	return &ComponentDelete{ComponentBaseAction: &ComponentBaseAction{Key: key}}
}

func (delete *ComponentDelete) Apply(plugins []plugin.EnginePlugin, eventLog *eventlog.EventLog) error {
	// Process destructions in the right order
	foundErrors := false

	// call plugins to perform their actions
	for _, pluginInstance := range plugins {
		err := pluginInstance.OnApplyComponentInstanceDelete(delete.Key)
		if err != nil {
			eventLog.LogError(err)
			foundErrors = true
		}
	}
	if foundErrors {
		return fmt.Errorf("One or more errors while applying changes (deleting running components)")
	}
	return nil
}
