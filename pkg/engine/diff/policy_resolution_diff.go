package diff

import (
	"github.com/Aptomi/aptomi/pkg/engine/apply/action"
	"github.com/Aptomi/aptomi/pkg/engine/apply/action/component"
	"github.com/Aptomi/aptomi/pkg/engine/apply/action/global"
	"github.com/Aptomi/aptomi/pkg/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/object"
)

// PolicyResolutionDiff represents a difference between two policy resolution data structs (actual and desired states)
type PolicyResolutionDiff struct {
	// Prev is actual policy resolution data
	Prev *resolve.PolicyResolution

	// Next is desired policy resolution data
	Next *resolve.PolicyResolution

	// Actions is a generated, ordered list of actions that need to be executed in order to get from an actual state to the desired state
	Actions []action.Base

	Revision object.Generation
}

// NewPolicyResolutionDiff calculates difference between two given policy resolution structs (actual and desired states).
// It iterates over all component instances and figures out which component instances have to be instantiated (new
// consumers appeared and they didn't exist before), which component instances have to be updated (e.g. parameters changed), which component
// instances have to be destroyed (that have no consumers left), and so on.
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

func appendUpdateAction(actions []action.Base, updateActions map[string]bool, updateAction *component.UpdateAction) []action.Base {
	if !updateActions[updateAction.GetName()] {
		updateActions[updateAction.GetName()] = true
		actions = append(actions, updateAction)
	}

	return actions
}

// On a component level -- see which component instance keys appear and disappear
// TODO: reduce cyclomatic complexity
func (diff *PolicyResolutionDiff) compareAndProduceActions() { // nolint: gocyclo
	actions := make(map[string][]action.Base)

	updateActions := make(map[string]bool)
	endpointsActions := make([]action.Base, 0)

	// merge all instance keys from prev and next
	allKeys := make(map[string]bool)
	for key := range diff.Prev.ComponentInstanceMap {
		allKeys[key] = true
	}
	for key := range diff.Next.ComponentInstanceMap {
		allKeys[key] = true
	}

	// go over all the keys and see which one appear and which one disappear
	for instanceKey := range allKeys {
		prevInstance := diff.Prev.ComponentInstanceMap[instanceKey]
		nextInstance := diff.Next.ComponentInstanceMap[instanceKey]

		var depKeysPrev map[string]bool
		if prevInstance != nil {
			depKeysPrev = prevInstance.DependencyKeys
		}

		var depKeysNext map[string]bool
		if nextInstance != nil {
			depKeysNext = nextInstance.DependencyKeys
		}

		createOrUpdate := false

		// see if a component needs to be instantiated
		if len(depKeysPrev) <= 0 && len(depKeysNext) > 0 {
			createOrUpdate = true
			actions[instanceKey] = append(actions[instanceKey], component.NewCreateAction(diff.Revision, instanceKey))
		}

		// see if a component needs to be destructed
		if len(depKeysPrev) > 0 && len(depKeysNext) <= 0 {
			actions[instanceKey] = append(actions[instanceKey], component.NewDeleteAction(diff.Revision, instanceKey))
		}

		// see if a component needs to be updated
		if len(depKeysPrev) > 0 && len(depKeysNext) > 0 {
			sameParams := prevInstance.CalculatedCodeParams.DeepEqual(nextInstance.CalculatedCodeParams)
			if !sameParams {
				createOrUpdate = true

				actions[instanceKey] = appendUpdateAction(actions[instanceKey], updateActions, component.NewUpdateAction(diff.Revision, instanceKey))

				// if it has a parent service, indicate that it basically gets updated as well
				// this is required for adjusting update/creation times of a service with changed component
				// this may produce duplicate "update" actions for the parent service
				if nextInstance.Metadata.Key.IsComponent() {
					serviceKey := nextInstance.Metadata.Key.GetParentServiceKey().GetKey()
					actions[serviceKey] = appendUpdateAction(actions[serviceKey], updateActions, component.NewUpdateAction(diff.Revision, serviceKey))
				}
			}
		}

		// see if a user needs to be detached from a component
		for dependencyID := range depKeysPrev {
			if !depKeysNext[dependencyID] {
				actions[instanceKey] = append(actions[instanceKey], component.NewDetachDependencyAction(diff.Revision, instanceKey, dependencyID))
			}
		}

		// see if a user needs to be attached to a component
		for dependencyID := range depKeysNext {
			if !depKeysPrev[dependencyID] {
				actions[instanceKey] = append(actions[instanceKey], component.NewAttachDependencyAction(diff.Revision, instanceKey, dependencyID))
			}
		}

		if createOrUpdate {
			endpointsActions = append(endpointsActions, component.NewEndpointsAction(diff.Revision, instanceKey))
		}
	}

	// Generate actions in the right order
	for _, key := range diff.Next.ComponentProcessingOrder {
		actionList, found := actions[key]
		if found {
			diff.Actions = append(diff.Actions, actionList...)
			delete(actions, key)
		}
	}
	for _, key := range diff.Prev.ComponentProcessingOrder {
		actionList, found := actions[key]
		if found {
			diff.Actions = append(diff.Actions, actionList...)
			delete(actions, key)
		}
	}

	diff.Actions = append(diff.Actions, endpointsActions...)

	// call a global post-processing action
	if len(diff.Actions) > 0 {
		diff.Actions = append(diff.Actions, global.NewPostProcessAction(diff.Revision))
	}
}
