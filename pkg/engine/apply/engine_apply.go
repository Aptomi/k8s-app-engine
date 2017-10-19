package apply

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/engine/actual"
	"github.com/Aptomi/aptomi/pkg/engine/apply/action"
	"github.com/Aptomi/aptomi/pkg/engine/progress"
	"github.com/Aptomi/aptomi/pkg/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/event"
	"github.com/Aptomi/aptomi/pkg/external"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/plugin"
)

// EngineApply executes actions to convert desired state to actual state
type EngineApply struct {
	// References to desired/actual objects
	desiredPolicy      *lang.Policy
	desiredState       *resolve.PolicyResolution
	actualPolicy       *lang.Policy
	actualState        *resolve.PolicyResolution
	actualStateUpdater actual.StateUpdater
	externalData       *external.Data
	plugins            plugin.Registry

	// Actions to be applied
	actions []action.Base

	// Buffered event log - gets populated while applying changes
	eventLog *event.Log

	// Progress indicator
	progress progress.Indicator
}

// NewEngineApply creates an instance of EngineApply
// todo(slukjanov): make sure that plugins are created once per revision, b/c we need to cache only for single policy, when it changed some credentials could change as well
// todo(slukjanov): run cleanup on all plugins after apply done for the revision
func NewEngineApply(desiredPolicy *lang.Policy, desiredState *resolve.PolicyResolution, actualPolicy *lang.Policy, actualState *resolve.PolicyResolution, actualStateUpdater actual.StateUpdater, externalData *external.Data, plugins plugin.Registry, actions []action.Base, progress progress.Indicator) *EngineApply {
	return &EngineApply{
		desiredPolicy:      desiredPolicy,
		desiredState:       desiredState,
		actualPolicy:       actualPolicy,
		actualState:        actualState,
		actualStateUpdater: actualStateUpdater,
		externalData:       externalData,
		plugins:            plugins,
		actions:            actions,
		eventLog:           event.NewLog(),
		progress:           progress,
	}
}

// Apply method applies all changes via plugins, updates actual state, returns the updated actual state and event log
func (apply *EngineApply) Apply() (*resolve.PolicyResolution, *event.Log, error) {
	// initialize progress indicator
	apply.progress.SetTotal(len(apply.actions))

	// error count while applying changes
	foundErrors := false

	// process all actions
	context := action.NewContext(
		apply.desiredPolicy,
		apply.desiredState,
		apply.actualPolicy,
		apply.actualState,
		apply.actualStateUpdater,
		apply.externalData,
		apply.plugins,
		apply.eventLog,
	)
	for _, act := range apply.actions {
		apply.progress.Advance("Action")
		err := act.Apply(context)
		if err != nil {
			err = fmt.Errorf("error while applying action '%s': %s", act, err)
			apply.eventLog.LogError(err)
			foundErrors = true
		}
	}

	// Finalize progress indicator
	apply.progress.Done()

	// Return error if there's been at least one error
	if foundErrors {
		err := fmt.Errorf("one or more errors occurred while running actions")
		apply.eventLog.LogError(err)
		return apply.actualState, apply.eventLog, err
	}

	// No errors occurred
	return apply.actualState, apply.eventLog, nil
}
