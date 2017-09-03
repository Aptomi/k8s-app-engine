package diff

import (
	"github.com/Aptomi/aptomi/pkg/slinga/engine/apply/actions"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/resolve"
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
	for key := range diff.Prev.ComponentInstanceMap {
		allKeys[key] = true
	}
	for key := range diff.Next.ComponentInstanceMap {
		allKeys[key] = true
	}

	// go over all the keys and see which one appear and which one disappear
	for key := range allKeys {
		uPrev := diff.Prev.ComponentInstanceMap[key]
		uNext := diff.Next.ComponentInstanceMap[key]

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
			actionsByKey[key] = append(actionsByKey[key], actions.NewComponentCreateAction(key))
		}

		// see if a component needs to be destructed
		if len(depIdsPrev) > 0 && len(depIdsNext) <= 0 {
			actionsByKey[key] = append(actionsByKey[key], actions.NewComponentDeleteAction(key))
		}

		// see if a component needs to be updated
		if len(depIdsPrev) > 0 && len(depIdsNext) > 0 {
			sameParams := uPrev.CalculatedCodeParams.DeepEqual(uNext.CalculatedCodeParams)
			if !sameParams {
				actionsByKey[key] = append(actionsByKey[key], actions.NewComponentUpdateAction(key))

				// if it has a parent service, indicate that it basically gets updated as well
				// this is required for adjusting update/creation times of a service with changed component
				// this may produce duplicate "update" actions for the parent service
				if uNext.Key.IsComponent() {
					serviceKey := uNext.Key.GetParentServiceKey().GetKey()
					actionsByKey[serviceKey] = append(actionsByKey[serviceKey], actions.NewComponentUpdateAction(serviceKey))
				}
			}
		}

		// see if a user needs to be detached from a component
		for dependencyID := range depIdsPrev {
			if !depIdsNext[dependencyID] {
				actionsByKey[key] = append(actionsByKey[key], actions.NewComponentDetachDependencyAction(key, dependencyID))
			}
		}

		// see if a user needs to be attached to a component
		for dependencyID := range depIdsNext {
			if !depIdsPrev[dependencyID] {
				actionsByKey[key] = append(actionsByKey[key], actions.NewComponentAttachDependencyAction(key, dependencyID))
			}
		}
	}

	// Generation actions in the right order
	for _, key := range diff.Next.ComponentProcessingOrder {
		actionList, found := actionsByKey[key]
		if found {
			diff.Actions = append(diff.Actions, normalize(actionList)...)
			delete(actionsByKey, key)
		}
	}
	for _, key := range diff.Prev.ComponentProcessingOrder {
		actionList, found := actionsByKey[key]
		if found {
			diff.Actions = append(diff.Actions, normalize(actionList)...)
			delete(actionsByKey, key)
		}
	}

	// Generate action for clusters
	diff.Actions = append(diff.Actions, actions.NewClustersPostProcessAction())
}

// TODO: refactor once we introduce Action kind
// Due to the nature of action list generation above, certain actions can be added more than once
// This will ensure that the list is normalized and there will be only one update action for each service instance
func normalize(list []actions.Action) []actions.Action {
	result := []actions.Action{}
	updateCnt := 0
	for _, action := range list {
		_, isUpdate := action.(*actions.ComponentUpdate)
		if isUpdate {
			if updateCnt == 0 {
				result = append(result, action)
			}
			updateCnt++
		} else {
			result = append(result, action)
		}
	}
	return result
}
