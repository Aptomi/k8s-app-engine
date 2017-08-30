package apply

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/diff"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/progress"
	"github.com/Aptomi/aptomi/pkg/slinga/errors"
	. "github.com/Aptomi/aptomi/pkg/slinga/eventlog"
)

type EngineApply struct {
	// Diff to be applied
	diff *diff.ServiceUsageStateDiff

	// Buffered event log - gets populated while applying changes
	eventLog *EventLog

	// Progress indicator
	progress progress.ProgressIndicator
}

func NewEngineApply(diff *diff.ServiceUsageStateDiff) *EngineApply {
	return &EngineApply{
		diff:     diff,
		eventLog: NewEventLog(),
		progress: progress.NewProgressConsole(),
	}
}

func (apply *EngineApply) logError(err error) {
	errWithDetails, isErrorWithDetails := err.(*errors.ErrorWithDetails)
	if isErrorWithDetails {
		apply.eventLog.WithFields(Fields(errWithDetails.Details())).Errorf(err.Error())
	} else {
		apply.eventLog.WithFields(Fields{}).Errorf(err.Error())
	}
}

// Apply method applies all changes via executors, saves usage state in Aptomi DB
func (apply *EngineApply) Apply() error {
	// initialize progress indicator
	apply.progress.SetTotal(apply.diff.GetApplyProgressLength())

	// error count while applying changes
	foundErrors := false

	// call plugins to perform their actions
	for _, plugin := range apply.diff.Plugins {
		err := plugin.OnApplyStart(apply.eventLog)
		if err != nil {
			apply.logError(fmt.Errorf("Error while calling OnApplyStart() on a plugin: " + err.Error()))
			foundErrors = true
		}
	}

	// process all actions
	err := apply.processDestructions()
	if err != nil {
		apply.logError(err)
		foundErrors = true
	}
	err = apply.processUpdates()
	if err != nil {
		apply.logError(err)
		foundErrors = true
	}
	err = apply.processInstantiations()
	if err != nil {
		apply.logError(err)
		foundErrors = true
	}

	// call plugins to perform their actions
	for _, plugin := range apply.diff.Plugins {
		err := plugin.OnApplyCustom(apply.progress)
		if err != nil {
			apply.logError(fmt.Errorf("Error while calling OnApplyCustom() on a plugin: " + err.Error()))
			foundErrors = true
		}
	}

	// Finalize progress indicator
	apply.progress.Done()

	if foundErrors {
		err := fmt.Errorf("One or more errors occured while applying policy")
		apply.logError(err)
		return err
	}
	return nil
}

func (apply *EngineApply) processInstantiations() error {
	// Process instantiations in the right order
	foundErrors := false
	for _, key := range apply.diff.Next.State.ResolvedData.ComponentProcessingOrder {
		// Does it need to be instantiated?
		if _, ok := apply.diff.ComponentInstantiate[key]; ok {
			// Advance progress indicator
			apply.progress.Advance("Create")

			// call plugins to perform their actions
			for _, plugin := range apply.diff.Plugins {
				err := plugin.OnApplyComponentInstanceCreate(apply.diff.Next.State.ResolvedData.ComponentInstanceMap[key])
				if err != nil {
					apply.logError(err)
					foundErrors = true
				}
			}
		}
	}

	if foundErrors {
		return fmt.Errorf("One or more errors while applying changes (creating new components)")
	}
	return nil
}

func (apply *EngineApply) processUpdates() error {
	// Process updates in the right order
	foundErrors := false
	for _, key := range apply.diff.Next.State.ResolvedData.ComponentProcessingOrder {
		// Does it need to be updated?
		if _, ok := apply.diff.ComponentUpdate[key]; ok {
			// Advance progress indicator
			apply.progress.Advance("Update")

			// call plugins to perform their actions
			for _, plugin := range apply.diff.Plugins {
				err := plugin.OnApplyComponentInstanceUpdate(apply.diff.Next.State.ResolvedData.ComponentInstanceMap[key])
				if err != nil {
					apply.logError(err)
					foundErrors = true
				}
			}
		}
	}
	if foundErrors {
		return fmt.Errorf("One or more errors while applying changes (updating running components)")
	}
	return nil
}

func (apply *EngineApply) processDestructions() error {
	// Process destructions in the right order
	foundErrors := false
	for _, key := range apply.diff.Prev.State.ResolvedData.ComponentProcessingOrder {
		// Does it need to be destructed?
		if _, ok := apply.diff.ComponentDestruct[key]; ok {
			// Advance progress indicator
			apply.progress.Advance("Delete")

			// call plugins to perform their actions
			for _, plugin := range apply.diff.Plugins {
				err := plugin.OnApplyComponentInstanceDelete(apply.diff.Prev.State.ResolvedData.ComponentInstanceMap[key])
				if err != nil {
					apply.logError(err)
					foundErrors = true
				}
			}
		}
	}
	if foundErrors {
		return fmt.Errorf("One or more errors while applying changes (deleting running components)")
	}
	return nil
}

func (apply *EngineApply) SaveLog() {
	// Save log
	hook := &HookBoltDB{}
	apply.eventLog.Save(hook)
}
