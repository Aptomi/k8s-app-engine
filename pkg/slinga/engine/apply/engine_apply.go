package apply

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/diff"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/progress"
	"github.com/Aptomi/aptomi/pkg/slinga/eventlog"
)

type EngineApply struct {
	// Diff to be applied
	diff *diff.ServiceUsageStateDiff

	// Buffered event log - gets populated while applying changes
	eventLog *eventlog.EventLog

	// Progress indicator
	progress progress.ProgressIndicator
}

func NewEngineApply(diff *diff.ServiceUsageStateDiff) *EngineApply {
	return &EngineApply{
		diff:     diff,
		eventLog: eventlog.NewEventLog(),
		progress: progress.NewProgressConsole(),
	}
}

// Apply method applies all changes via executors, saves usage state in Aptomi DB
func (apply *EngineApply) Apply() error {
	// initialize progress indicator
	apply.progress.SetTotal(apply.diff.GetApplyProgressLength())

	// call plugins to perform their actions
	for _, plugin := range apply.diff.Plugins {
		err := plugin.OnApplyStart(apply.eventLog)
		if err != nil {
			return fmt.Errorf("Error while calling OnApplyStart() on a plugin: " + err.Error())
		}
	}

	// process all actions
	err := apply.processDestructions()
	if err != nil {
		return fmt.Errorf("Error while destructing components: " + err.Error())
	}
	err = apply.processUpdates()
	if err != nil {
		return fmt.Errorf("Error while updating components: " + err.Error())
	}
	err = apply.processInstantiations()
	if err != nil {
		return fmt.Errorf("Error while instantiating components: " + err.Error())
	}

	// call plugins to perform their actions
	for _, plugin := range apply.diff.Plugins {
		err := plugin.OnApplyCustom(apply.progress)
		if err != nil {
			return fmt.Errorf("Error while calling OnApplyCustom() on a plugin: " + err.Error())
		}
	}

	// Finalize progress indicator
	apply.progress.Done()
	return nil
}

func (apply *EngineApply) processInstantiations() error {
	// Process instantiations in the right order
	for _, key := range apply.diff.Next.State.ResolvedData.ComponentProcessingOrder {
		// Does it need to be instantiated?
		if _, ok := apply.diff.ComponentInstantiate[key]; ok {
			// Advance progress indicator
			apply.progress.Advance("Create")

			// call plugins to perform their actions
			for _, plugin := range apply.diff.Plugins {
				err := plugin.OnApplyComponentInstanceCreate(apply.diff.Next.State.ResolvedData.ComponentInstanceMap[key])
				if err != nil {
					return fmt.Errorf("Error while calling OnApplyComponentInstanceCreate() on a plugin: " + err.Error())
				}
			}
		}
	}
	return nil
}

func (apply *EngineApply) processUpdates() error {
	// Process updates in the right order
	for _, key := range apply.diff.Next.State.ResolvedData.ComponentProcessingOrder {
		// Does it need to be updated?
		if _, ok := apply.diff.ComponentUpdate[key]; ok {
			// Advance progress indicator
			apply.progress.Advance("Update")

			// call plugins to perform their actions
			for _, plugin := range apply.diff.Plugins {
				err := plugin.OnApplyComponentInstanceUpdate(apply.diff.Next.State.ResolvedData.ComponentInstanceMap[key])
				if err != nil {
					return fmt.Errorf("Error while calling OnApplyComponentInstanceUpdate() on a plugin: " + err.Error())
				}
			}
		}
	}
	return nil
}

func (apply *EngineApply) processDestructions() error {
	// Process destructions in the right order
	for _, key := range apply.diff.Prev.State.ResolvedData.ComponentProcessingOrder {
		// Does it need to be destructed?
		if _, ok := apply.diff.ComponentDestruct[key]; ok {
			// Advance progress indicator
			apply.progress.Advance("Delete")

			// call plugins to perform their actions
			for _, plugin := range apply.diff.Plugins {
				err := plugin.OnApplyComponentInstanceDelete(apply.diff.Prev.State.ResolvedData.ComponentInstanceMap[key])
				if err != nil {
					return fmt.Errorf("Error while calling OnApplyComponentInstanceDelete() on a plugin: " + err.Error())
				}
			}
		}
	}
	return nil
}
