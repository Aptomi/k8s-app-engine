package diff

import (
	"github.com/Aptomi/aptomi/pkg/slinga/engine/apply/actions"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/resolve"
	"time"
)

// PolicyResolutionDiff represents a difference between two policy resolution data structs
type PolicyResolutionDiff struct {
	// Previous policy resolution data
	Prev *resolve.PolicyResolution

	// Previous policy resolution data
	Next *resolve.PolicyResolution

	// Actions that need to be taken, in the right order
	Actions []actions.Action
}

// NewPolicyResolutionDiff calculates difference between two given policy resolution structs
func NewPolicyResolutionDiff(next *resolve.PolicyResolution, prev *resolve.PolicyResolution) *PolicyResolutionDiff {
	result := &PolicyResolutionDiff{
		Prev:    prev,
		Next:    next,
		Actions: []actions.Action{},
	}
	result.compareAndProduceActions()
	return result
}

// On a component level -- see which component instance keys appear and disappear
func (diff *PolicyResolutionDiff) compareAndProduceActions() {
	actionsByKey := make(map[string][]actions.Action)

	// merge all instance keys from prev and next
	allKeys := make(map[string]bool)
	for key := range diff.Prev.Resolved.ComponentInstanceMap {
		allKeys[key] = true
	}
	for key := range diff.Next.Resolved.ComponentInstanceMap {
		allKeys[key] = true
	}

	// go over all the keys and see which one appear and which one disappear
	for key := range allKeys {
		uPrev := diff.Prev.Resolved.ComponentInstanceMap[key]
		uNext := diff.Next.Resolved.ComponentInstanceMap[key]

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
			actionsByKey[key] = append(actionsByKey[key], actions.NewComponentCreateAction(key, diff.Next, diff.Prev))
			diff.updateTimes(uNext.Key, time.Now(), time.Now())
		}

		// see if a component needs to be destructed
		if len(depIdsPrev) > 0 && len(depIdsNext) <= 0 {
			actionsByKey[key] = append(actionsByKey[key], actions.NewComponentDeleteAction(key, diff.Next, diff.Prev))
			diff.updateTimes(uPrev.Key, uPrev.CreatedOn, time.Now())
		}

		// see if a component needs to be updated
		if len(depIdsPrev) > 0 && len(depIdsNext) > 0 {
			sameParams := uPrev.CalculatedCodeParams.DeepEqual(uNext.CalculatedCodeParams)
			if !sameParams {
				actionsByKey[key] = append(actionsByKey[key], actions.NewComponentUpdateAction(key, diff.Next, diff.Prev))
				diff.updateTimes(uNext.Key, uPrev.CreatedOn, time.Now())
			} else {
				diff.updateTimes(uNext.Key, uPrev.CreatedOn, uPrev.UpdatedOn)
			}
		}

		// see if a user needs to be detached from a component
		for dependencyID := range depIdsPrev {
			if !depIdsNext[dependencyID] {
				actionsByKey[key] = append(actionsByKey[key], actions.NewComponentDetachDependencyAction(key, dependencyID, diff.Next, diff.Prev))
			}
		}

		// see if a user needs to be attached to a component
		for dependencyID := range depIdsNext {
			if !depIdsPrev[dependencyID] {
				actionsByKey[key] = append(actionsByKey[key], actions.NewComponentAttachDependencyAction(key, dependencyID, diff.Next, diff.Prev))
			}
		}
	}

	// Generation actions in the right order
	for _, key := range diff.Next.Resolved.ComponentProcessingOrder {
		actionList, found := actionsByKey[key]
		if found {
			diff.Actions = append(diff.Actions, actionList...)
			delete(actionsByKey, key)
		}
	}
	for _, key := range diff.Prev.Resolved.ComponentProcessingOrder {
		actionList, found := actionsByKey[key]
		if found {
			diff.Actions = append(diff.Actions, actionList...)
			delete(actionsByKey, key)
		}
	}
}

// updated timestamps for component (and root service, if/as needed)
func (diff *PolicyResolutionDiff) updateTimes(cik *resolve.ComponentInstanceKey, createdOn time.Time, updatedOn time.Time) {
	// update for a given node
	instance := diff.Next.Resolved.ComponentInstanceMap[cik.GetKey()]
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
