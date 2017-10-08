package diff

import (
	"github.com/Aptomi/aptomi/pkg/slinga/engine/apply/action"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/apply/action/cluster"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/apply/action/component"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/slinga/object"
)

// PolicyResolutionDiff represents a difference between two policy resolution data structs
type PolicyResolutionDiff struct {
	// Prev is actual policy resolution data
	Prev *resolve.PolicyResolution

	// Next is desired policy resolution data
	Next *resolve.PolicyResolution

	// Actions is a generated, ordered list of actions that need to be executed in order to get from an actual state to the desired state
	Actions []action.Base

	Revision object.Generation
}

// NewPolicyResolutionDiff calculates difference between two given policy resolution structs
func NewPolicyResolutionDiff(next *resolve.PolicyResolution, prev *resolve.PolicyResolution, revision object.Generation) *PolicyResolutionDiff {
	result := &PolicyResolutionDiff{
		Prev:     prev,
		Next:     next,
		Actions:  []action.Base{},
		Revision: revision,
	}
	result.compareAndProduceActions()
	return result
}

// On a component level -- see which component instance keys appear and disappear
func (diff *PolicyResolutionDiff) compareAndProduceActions() {
	actionsByKey := make(map[string][]action.Base)

	// merge all instance keys from prev and next
	allKeys := make(map[string]bool)
	for key := range diff.Prev.ComponentInstanceMap {
		allKeys[key] = true
	}
	for key := range diff.Next.ComponentInstanceMap {
		allKeys[key] = true
	}

	// go over all the keys and see which one appear and which one disappear
	for componentKey := range allKeys {
		uPrev := diff.Prev.ComponentInstanceMap[componentKey]
		uNext := diff.Next.ComponentInstanceMap[componentKey]

		var depKeysPrev map[string]bool
		if uPrev != nil {
			depKeysPrev = uPrev.DependencyKeys
		}

		var depKeysNext map[string]bool
		if uNext != nil {
			depKeysNext = uNext.DependencyKeys
		}

		// see if a component needs to be instantiated
		if len(depKeysPrev) <= 0 && len(depKeysNext) > 0 {
			actionsByKey[componentKey] = append(actionsByKey[componentKey], component.NewCreateAction(diff.Revision, componentKey))
		}

		// see if a component needs to be destructed
		if len(depKeysPrev) > 0 && len(depKeysNext) <= 0 {
			actionsByKey[componentKey] = append(actionsByKey[componentKey], component.NewDeleteAction(diff.Revision, componentKey))
		}

		// see if a component needs to be updated
		if len(depKeysPrev) > 0 && len(depKeysNext) > 0 {
			sameParams := uPrev.CalculatedCodeParams.DeepEqual(uNext.CalculatedCodeParams)
			if !sameParams {
				actionsByKey[componentKey] = append(actionsByKey[componentKey], component.NewUpdateAction(diff.Revision, componentKey))

				// if it has a parent service, indicate that it basically gets updated as well
				// this is required for adjusting update/creation times of a service with changed component
				// this may produce duplicate "update" actions for the parent service
				if uNext.Metadata.Key.IsComponent() {
					serviceKey := uNext.Metadata.Key.GetParentServiceKey().GetKey()
					actionsByKey[serviceKey] = append(actionsByKey[serviceKey], component.NewUpdateAction(diff.Revision, serviceKey))
				}
			}
		}

		// see if a user needs to be detached from a component
		for dependencyID := range depKeysPrev {
			if !depKeysNext[dependencyID] {
				actionsByKey[componentKey] = append(actionsByKey[componentKey], component.NewDetachDependencyAction(diff.Revision, componentKey, dependencyID))
			}
		}

		// see if a user needs to be attached to a component
		for dependencyID := range depKeysNext {
			if !depKeysPrev[dependencyID] {
				actionsByKey[componentKey] = append(actionsByKey[componentKey], component.NewAttachDependencyAction(diff.Revision, componentKey, dependencyID))
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
	diff.Actions = append(diff.Actions, cluster.NewClustersPostProcessAction(diff.Revision))
}

// TODO: refactor once we introduce Action kind
// Due to the nature of action list generation above, certain actions can be added more than once
// This will ensure that the list is normalized and there will be only one update action for each service instance
func normalize(actions []action.Base) []action.Base {
	result := []action.Base{}
	updateCnt := 0
	for _, act := range actions {
		_, isUpdate := act.(*component.UpdateAction)
		if isUpdate {
			if updateCnt == 0 {
				result = append(result, act)
			}
			updateCnt++
		} else {
			result = append(result, act)
		}
	}
	return result
}

func (diff *PolicyResolutionDiff) IsChanged() bool {
	return len(diff.Actions) > 0
}
