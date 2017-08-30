package diff

import (
	"fmt"
	. "github.com/Aptomi/aptomi/pkg/slinga/db"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/plugin"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/slinga/eventlog"
	log "github.com/Sirupsen/logrus"
	"time"
)

// ServiceUsageDependencyAction is a <ComponentKey, DependencyID> object. It holds data for attach/detach operations for component instance <-> dependency
type ServiceUsageDependencyAction struct {
	ComponentKey string
	DependencyID string
}

// ServiceUsageStateDiff represents a difference between two calculated usage states
type ServiceUsageStateDiff struct {
	// Previous resolved state (policy, resolved policy, external objects)
	Prev *resolve.ResolvedState

	// Next resolved state (policy, resolved policy, external objects)
	Next *resolve.ResolvedState

	// Actions that need to be taken
	ComponentInstantiate      map[string]bool
	ComponentDestruct         map[string]bool
	ComponentUpdate           map[string]bool
	ComponentAttachDependency []ServiceUsageDependencyAction
	ComponentDetachDependency []ServiceUsageDependencyAction

	// Plugins (will be called during diff processing)
	Plugins []plugin.EnginePlugin

	// Diff stored as text
	DiffAsText string
}

// CalculateDifference calculates difference between two given usage states
func NewServiceUsageStateDiff(next *resolve.ResolvedState, prev *resolve.ResolvedState) *ServiceUsageStateDiff {
	// resulting difference
	result := &ServiceUsageStateDiff{
		Prev:                 prev,
		Next:                 next,
		ComponentInstantiate: make(map[string]bool),
		ComponentDestruct:    make(map[string]bool),
		ComponentUpdate:      make(map[string]bool),
		Plugins:              plugin.AllPlugins(),
	}

	result.calculateDifference()
	return result
}

// ProcessSuccessfulExecution increments revision and saves results of the current run when policy processing executed successfully
func (diff *ServiceUsageStateDiff) ProcessSuccessfulExecution(revision AptomiRevision, newrevision bool, noop bool) {
	fmt.Println("[Revision]")
	if newrevision || (!noop && diff.ShouldGenerateNewRevision()) {
		// Increment a revision
		newRevision := revision.Increment()

		// Save results of the current run
		newRevision.SaveCurrentRun()

		// Save updated revision number
		newRevision.SaveAsLastRevision()

		// Print revision numbers
		fmt.Printf("  Previous: %s\n", revision.String())
		fmt.Printf("  Current: %s\n", newRevision.String())
	} else {
		fmt.Printf("  Current: %s (no changes made)\n", revision.String())
	}
}

// On a component level -- see which component instance keys appear and disappear
func (diff *ServiceUsageStateDiff) calculateDifference() {
	// map of all instances
	allKeys := make(map[string]bool)

	// merge all the keys
	for k := range diff.Prev.State.ResolvedData.ComponentInstanceMap {
		allKeys[k] = true
	}
	for k := range diff.Next.State.ResolvedData.ComponentInstanceMap {
		allKeys[k] = true
	}
	// Go over all the keys and see which one appear and which one disappear
	for k := range allKeys {
		uPrev := diff.Prev.State.ResolvedData.ComponentInstanceMap[k]
		uNext := diff.Next.State.ResolvedData.ComponentInstanceMap[k]

		var depIdsPrev map[string]bool
		if uPrev != nil {
			depIdsPrev = uPrev.DependencyIds
		}

		var depIdsNext map[string]bool
		if uNext != nil {
			depIdsNext = uNext.DependencyIds
		}

		// see if a component needs to be instantiated
		if len(depIdsPrev) <= 0 && len(depIdsNext) > 0 {
			diff.ComponentInstantiate[k] = true
			diff.updateTimes(uNext.Key, time.Now(), time.Now())
		}

		// see if a component needs to be destructed
		if len(depIdsPrev) > 0 && len(depIdsNext) <= 0 {
			diff.ComponentDestruct[k] = true
			diff.updateTimes(uPrev.Key, uPrev.CreatedOn, time.Now())
		}

		// see if a component needs to be updated
		if len(depIdsPrev) > 0 && len(depIdsNext) > 0 {
			sameParams := uPrev.CalculatedCodeParams.DeepEqual(uNext.CalculatedCodeParams)
			if !sameParams {
				diff.ComponentUpdate[k] = true
				diff.updateTimes(uNext.Key, uPrev.CreatedOn, time.Now())
			} else {
				diff.updateTimes(uNext.Key, uPrev.CreatedOn, uPrev.UpdatedOn)
			}
		}

		// see if a user needs to be detached from a component
		for dependencyID := range depIdsPrev {
			if !depIdsNext[dependencyID] {
				diff.ComponentDetachDependency = append(diff.ComponentDetachDependency, ServiceUsageDependencyAction{ComponentKey: k, DependencyID: dependencyID})
			}
		}

		// see if a user needs to be attached to a component
		for dependencyID := range depIdsNext {
			if !depIdsPrev[dependencyID] {
				diff.ComponentAttachDependency = append(diff.ComponentAttachDependency, ServiceUsageDependencyAction{ComponentKey: k, DependencyID: dependencyID})
			}
		}
	}

	// initialize all plugins
	for _, pluginInstance := range diff.Plugins {
		pluginInstance.Init(diff.Next, diff.Prev)
	}
}

// updated timestamps for component (and root service, if/as needed)
func (diff *ServiceUsageStateDiff) updateTimes(cik *resolve.ComponentInstanceKey, createdOn time.Time, updatedOn time.Time) {
	// update for a given node
	instance := diff.Next.State.ResolvedData.ComponentInstanceMap[cik.GetKey()]
	if instance == nil {
		// likely this component has been deleted as a part of the diff
		return
	}
	instance.UpdateTimes(createdOn, updatedOn)

	// if it's a component instance, then update for its parent service instance as well
	if cik.IsComponent() {
		diff.updateTimes(cik.GetParentServiceKey(), createdOn, updatedOn)
	}
}

// Returns difference length (used for progress indicator)
func (diff *ServiceUsageStateDiff) GetApplyProgressLength() int {
	result := len(diff.ComponentInstantiate) +
		len(diff.ComponentDestruct) +
		len(diff.ComponentUpdate)

	for _, pluginInstance := range diff.Plugins {
		result += pluginInstance.GetCustomApplyProgressLength()
	}

	return result
}

// On a component level -- see which keys appear and disappear
func (diff *ServiceUsageStateDiff) writeDifferenceOnComponentLevel(verbose bool, log *log.Logger) {
	log.Println("[Components]")

	// Print
	changes := diff.GetApplyProgressLength() > 0

	if changes {
		log.Printf("  New instances:     %d", len(diff.ComponentInstantiate))
		log.Printf("  Deleted instances: %d", len(diff.ComponentDestruct))
		log.Printf("  Updated instances: %d", len(diff.ComponentUpdate))

		log.Println("[Component Instances]")
		if verbose {
			if len(diff.ComponentInstantiate) > 0 {
				log.Debug("  New:")
				for k := range diff.ComponentInstantiate {
					log.Printf("    [+] %s", k)
				}
			}

			if len(diff.ComponentDestruct) > 0 {
				log.Debug("  Deleted:")
				for k := range diff.ComponentDestruct {
					log.Printf("    [-] %s", k)
				}
			}

			if len(diff.ComponentUpdate) > 0 {
				log.Debug("  Updated:")
				for k := range diff.ComponentUpdate {
					log.Printf("    [*] %s", k)
				}
			}
		} else {
			log.Println("  Use --verbose to see the list")
		}
	}

	if !changes {
		log.Println("  [*] No changes")
	}

}

// StoreDiffAsText method prints changes onto the screen (i.e. delta - what got added/removed)
func (diff *ServiceUsageStateDiff) StoreDiffAsText(verbose bool) {
	memLog := eventlog.NewPlainMemoryLogger(verbose)
	diff.writeSummary(memLog.GetLogger())
	diff.writeDifferenceOnServicesLevel(memLog.GetLogger())
	diff.writeDifferenceOnComponentLevel(verbose, memLog.GetLogger())
	diff.DiffAsText = memLog.GetBuffer().String()
}

// AlterDifference basically marks all objects in diff for recreation/update if full update is requested. This is useful to re-create missing objects if they got deleted externally after deployment
func (diff *ServiceUsageStateDiff) AlterDifference(full bool) {
	// If we are requesting full policy processing, then we will need to re-create all objects
	if full {
		for key, instance := range diff.Next.State.ResolvedData.ComponentInstanceMap {
			diff.ComponentInstantiate[key] = true
			diff.updateTimes(instance.Key, instance.CreatedOn, time.Now())
		}
	}
}
