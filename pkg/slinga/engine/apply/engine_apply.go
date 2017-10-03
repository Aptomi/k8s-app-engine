package apply

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/actual"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/apply/action"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/progress"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/slinga/eventlog"
	"github.com/Aptomi/aptomi/pkg/slinga/external"
	"github.com/Aptomi/aptomi/pkg/slinga/lang"
	"github.com/Aptomi/aptomi/pkg/slinga/plugin"
)

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
	eventLog *eventlog.EventLog

	// Progress indicator
	progress progress.ProgressIndicator
}

// todo(slukjanov): make sure that plugins are created once per revision, b/c we need to cache only for single policy, when it changed some credentials could change as well
// todo(slukjanov): run cleanup on all plugins after apply done for the revision
func NewEngineApply(desiredPolicy *lang.Policy, desiredState *resolve.PolicyResolution, actualPolicy *lang.Policy, actualState *resolve.PolicyResolution, actualStateUpdater actual.StateUpdater, externalData *external.Data, plugins plugin.Registry, actions []action.Base) *EngineApply {
	return &EngineApply{
		desiredPolicy:      desiredPolicy,
		desiredState:       desiredState,
		actualPolicy:       actualPolicy,
		actualState:        actualState,
		actualStateUpdater: actualStateUpdater,
		externalData:       externalData,
		plugins:            plugins,
		actions:            actions,
		eventLog:           eventlog.NewEventLog(),
		progress:           progress.NewProgressConsole(),
	}
}

// Apply method applies all changes via plugins, updates actual state, returns the updated actual state and event log
func (apply *EngineApply) Apply() (*resolve.PolicyResolution, *eventlog.EventLog, error) {
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
			foundErrors = true
		}
	}

	// Finalize progress indicator
	apply.progress.Done()

	// Return error if there's been at least one error
	if foundErrors {
		err := fmt.Errorf("One or more errors occured while running actions")
		apply.eventLog.LogError(err)
		return apply.actualState, apply.eventLog, err
	}

	// No errors occurred
	return apply.actualState, apply.eventLog, nil
}
