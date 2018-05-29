package apply

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/engine/actual"
	"github.com/Aptomi/aptomi/pkg/engine/apply/action"
	"github.com/Aptomi/aptomi/pkg/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/event"
	"github.com/Aptomi/aptomi/pkg/external"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/plugin"
	"runtime/debug"
)

// EngineApply executes actions to get from an actual state to desired state
type EngineApply struct {
	// References to desired/actual objects
	desiredPolicy      *lang.Policy
	desiredState       *resolve.PolicyResolution
	actualState        *resolve.PolicyResolution
	actualStateUpdater actual.StateUpdater
	externalData       *external.Data
	plugins            plugin.Registry

	// Action plan to be applied
	actionPlan *action.Plan

	// Buffered event log - gets populated while applying actions
	eventLog *event.Log

	// Result/progress updater
	updater action.ApplyResultUpdater
}

// NewEngineApply creates an instance of EngineApply
// todo(slukjanov): make sure that plugins are created once per revision, b/c we need to cache only for single policy, when it changed some credentials could change as well
// todo(slukjanov): run cleanup on all plugins after apply done for the revision
func NewEngineApply(desiredPolicy *lang.Policy, desiredState *resolve.PolicyResolution, actualState *resolve.PolicyResolution, actualStateUpdater actual.StateUpdater, externalData *external.Data, plugins plugin.Registry, actionPlan *action.Plan, eventLog *event.Log, updater action.ApplyResultUpdater) *EngineApply {
	return &EngineApply{
		desiredPolicy:      desiredPolicy,
		desiredState:       desiredState,
		actualState:        actualState,
		actualStateUpdater: actualStateUpdater,
		externalData:       externalData,
		plugins:            plugins,
		actionPlan:         actionPlan,
		eventLog:           eventLog,
		updater:            updater,
	}
}

// Apply method executes all actions, actions call plugins to apply changes and roll them out to the cloud.
// It returns the updated actual state inside PolicyResolution and event log, as well as result/stats about how many actions
// have been applied successfully vs. failed vs. skipped.
//
// As actions get executed, they will instantiate/update/delete components according to the resolved
// policy, as well as configure the underlying cloud components appropriately. In case of errors (e.g. cloud is not
// available), actual state may not be equal to desired state after performing all the actions.
func (apply *EngineApply) Apply() (*resolve.PolicyResolution, *action.ApplyResult) {
	// process all actions
	context := action.NewContext(
		apply.desiredPolicy,
		apply.desiredState,
		apply.actualState,
		apply.actualStateUpdater,
		apply.externalData,
		apply.plugins,
		apply.eventLog,
	)

	// TODO: apply in parallel, https://github.com/Aptomi/aptomi/issues/310
	// Note that this function will be called in different go routines by apply
	// So we will need to
	// (1) Remove WrapSequential to make action execution parallel
	// (2) Restrict the number of parallel actions per cluster (make sure plugin can take care of that)
	// (3) Ensure that plugins are "thread-safe"
	// (4) Ensure that when we are updating states (e.g. Desired -> Actual), this is also "thread-safe"
	result := apply.actionPlan.Apply(action.WrapSequential(func(act action.Base) error {
		err := apply.executeAction(act, context)
		if err != nil {
			apply.eventLog.NewEntry().Errorf("error while applying action '%s': %s", act, err)
		}
		return err
	}), apply.updater)

	// No errors occurred
	return apply.actualState, result
}

func (apply *EngineApply) executeAction(action action.Base, context *action.Context) (errResult error) {
	// make sure we are converting panics into errors
	defer func() {
		if err := recover(); err != nil {
			errResult = fmt.Errorf("panic: %s\n%s", err, string(debug.Stack()))
		}
	}()

	return action.Apply(context)
}
