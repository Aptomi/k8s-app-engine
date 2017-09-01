package actions

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/plugin"
	"github.com/Aptomi/aptomi/pkg/slinga/eventlog"
)

type ComponentCreate struct {
	*ComponentBaseAction
}

func NewComponentCreateAction(key string) *ComponentCreate {
	return &ComponentCreate{ComponentBaseAction: &ComponentBaseAction{Key: key}}
}

func (create *ComponentCreate) Apply(plugins []plugin.EnginePlugin, eventLog *eventlog.EventLog) error {
	// Process instantiations in the right order
	foundErrors := false

	// call plugins to perform their actions
	for _, pluginInstance := range plugins {
		err := pluginInstance.OnApplyComponentInstanceCreate(create.Key)
		if err != nil {
			eventLog.LogError(err)
			foundErrors = true
		}
	}
	if foundErrors {
		return fmt.Errorf("One or more errors while applying changes (creating new components)")
	}
	return nil
}
