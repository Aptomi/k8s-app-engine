package apply

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/apply/actions"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/plugin"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/progress"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/resolve"
	. "github.com/Aptomi/aptomi/pkg/slinga/eventlog"
	"github.com/Aptomi/aptomi/pkg/slinga/external"
	"github.com/Aptomi/aptomi/pkg/slinga/language"
)

type EngineApply struct {
	// References to desired/actual objects
	desiredPolicy *language.PolicyNamespace
	desiredState  *resolve.PolicyResolution
	actualPolicy  *language.PolicyNamespace
	actualState   *resolve.PolicyResolution
	externalData  *external.Data

	// Actions to be applied
	actions []actions.Action

	// Plugins to execute
	plugins []plugin.EnginePlugin

	// Buffered event log - gets populated while applying changes
	eventLog *EventLog

	// Progress indicator
	progress progress.ProgressIndicator
}

func NewEngineApply(desiredPolicy *language.PolicyNamespace, desiredState *resolve.PolicyResolution, actualPolicy *language.PolicyNamespace, actualState *resolve.PolicyResolution, externalData *external.Data, actions []actions.Action, plugins []plugin.EnginePlugin) *EngineApply {
	return &EngineApply{
		desiredPolicy: desiredPolicy,
		desiredState:  desiredState,
		actualPolicy:  actualPolicy,
		actualState:   actualState,
		externalData:  externalData,
		actions:       actions,
		plugins:       plugins,
		eventLog:      NewEventLog(),
		progress:      progress.NewProgressConsole(),
	}
}

// Returns difference length (used for progress indicator)
func (apply *EngineApply) getApplyProgressLength() int {
	result := len(apply.actions)
	for _, pluginInstance := range apply.plugins {
		result += pluginInstance.GetCustomApplyProgressLength()
	}
	return result
}

// Apply method applies all changes via plugins, updates actual state, returns the updated actual state and event log
func (apply *EngineApply) Apply() (*resolve.PolicyResolution, *EventLog, error) {
	// initialize all plugins
	for _, pluginInstance := range apply.plugins {
		pluginInstance.Init(
			apply.desiredPolicy,
			apply.desiredState,
			apply.actualPolicy,
			apply.actualState,
			apply.externalData,
			apply.eventLog,
		)
	}

	// initialize progress indicator
	apply.progress.SetTotal(apply.getApplyProgressLength())

	// error count while applying changes
	foundErrors := false

	// process all actions
	for _, action := range apply.actions {
		apply.progress.Advance("Action")
		err := action.Apply(apply.plugins, apply.eventLog)
		if err != nil {
			foundErrors = true
		}
	}

	// call plugins to perform their custom apply actions
	for _, pluginInstance := range apply.plugins {
		err := pluginInstance.OnApplyCustom(apply.progress)
		if err != nil {
			apply.eventLog.LogError(fmt.Errorf("Error while calling OnApplyCustom() on a plugin: " + err.Error()))
			foundErrors = true
		}
	}

	// Finalize progress indicator
	apply.progress.Done()

	// Return error if there's been at least one error
	if foundErrors {
		err := fmt.Errorf("One or more errors occured while applying policy")
		apply.eventLog.LogError(err)
		return apply.actualState, apply.eventLog, err
	}

	// No errors occurred
	return apply.actualState, apply.eventLog, nil
}

func (apply *EngineApply) SaveLog() {
	// Save log
	hook := &HookBoltDB{}
	apply.eventLog.Save(hook)
}
