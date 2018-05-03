package diff

import (
	"github.com/Aptomi/aptomi/pkg/engine/apply/action"
	"github.com/Aptomi/aptomi/pkg/engine/apply/action/component"
	"github.com/Aptomi/aptomi/pkg/engine/resolve"
)

// PolicyResolutionDiff represents a difference between two policy resolution data structs (actual and desired states)
type PolicyResolutionDiff struct {
	// Prev is actual policy resolution data
	Prev *resolve.PolicyResolution

	// Next is desired policy resolution data
	Next *resolve.PolicyResolution

	// Plan is a plan of actions to transform Prev to Next
	ActionPlan *action.Plan
}

// NewPolicyResolutionDiff calculates difference between prev and next policy resolution structs (actual and desired states).
// It figures out which component instances have to be instantiated (new consumers appeared and they didn't exist before),
// which component instances have to be updated (e.g. parameters changed), which component instances have to be destroyed
// (that have no consumers left), and so on.
//
// Based on that it produces a graph of actions which have to be executed to transform prev to next.
func NewPolicyResolutionDiff(next *resolve.PolicyResolution, prev *resolve.PolicyResolution) *PolicyResolutionDiff {
	result := &PolicyResolutionDiff{
		Prev:       prev,
		Next:       next,
		ActionPlan: action.NewPlan(),
	}
	result.compareAndProduceActions()
	return result
}

// Produce a list of actions
func (diff *PolicyResolutionDiff) compareAndProduceActions() {
	// Produce a map of all component instances & root component instances
	allCompInstances := make(map[string]bool)
	for keyPrev, compPrev := range diff.Prev.ComponentInstanceMap {
		allCompInstances[keyPrev] = true
		if len(compPrev.EdgesOut) <= 0 {
			diff.ActionPlan.AddLeafNode(keyPrev)
		}
	}
	for keyNext, compNext := range diff.Next.ComponentInstanceMap {
		allCompInstances[keyNext] = true
		if len(compNext.EdgesOut) <= 0 {
			diff.ActionPlan.AddLeafNode(keyNext)
		}
	}

	// Build a flat list of actions for every component key
	for key := range allCompInstances {
		diff.buildActions(key)
	}

	// Generate dependencies between actions
	for key := range allCompInstances {
		outgoing := make(map[string]bool)
		if diff.Prev.ComponentInstanceMap[key] != nil {
			for keyOutPrev := range diff.Prev.ComponentInstanceMap[key].EdgesOut {
				outgoing[keyOutPrev] = true
			}
		}
		if diff.Next.ComponentInstanceMap[key] != nil {
			for keyOutNext := range diff.Next.ComponentInstanceMap[key].EdgesOut {
				outgoing[keyOutNext] = true
			}
		}

		for keyOut := range outgoing {
			diff.ActionPlan.GetActionGraphNode(key).AddBefore(diff.ActionPlan.GetActionGraphNode(keyOut))
		}
	}
}

// Traverse a graph for a given component instance
func (diff *PolicyResolutionDiff) buildActions(key string) {
	// Get action graph node for a given component key
	node := diff.ActionPlan.GetActionGraphNode(key)

	// Get previous dependency keys
	var depKeysPrev map[string]bool
	prevInstance := diff.Prev.ComponentInstanceMap[key]
	if prevInstance != nil {
		depKeysPrev = prevInstance.DependencyKeys
	}

	// Get next dependency keys
	var depKeysNext map[string]bool
	nextInstance := diff.Next.ComponentInstanceMap[key]
	if nextInstance != nil {
		depKeysNext = nextInstance.DependencyKeys
	}

	// See if it's a service or component
	isCodeComponent := (prevInstance != nil && prevInstance.IsCode) || (nextInstance != nil && nextInstance.IsCode)

	// Bool that says that we should retrieve endpoints
	endpointsAction := false

	// See if a component needs to be instantiated
	if len(depKeysPrev) <= 0 && len(depKeysNext) > 0 {
		node.AddAction(component.NewCreateAction(key), true)
		endpointsAction = true
	}

	// See if a component needs to be updated
	if len(depKeysPrev) > 0 && len(depKeysNext) > 0 && isCodeComponent {
		sameParams := prevInstance.CalculatedCodeParams.DeepEqual(nextInstance.CalculatedCodeParams)
		if !sameParams {
			node.AddAction(component.NewUpdateAction(key), true)

			// indicate that a parent service component instance gets updated as well
			// this is required for adjusting update/creation times of a service with changed component
			// this may produce duplicate "update" actions for the parent service
			serviceKey := nextInstance.Metadata.Key.GetParentServiceKey().GetKey()
			serviceNode := diff.ActionPlan.GetActionGraphNode(serviceKey)
			serviceNode.AddAction(component.NewUpdateAction(serviceKey), true)

			endpointsAction = true
		}
	}

	// See if a user needs to be attached to a component
	for dependencyID := range depKeysNext {
		if !depKeysPrev[dependencyID] {
			node.AddAction(component.NewAttachDependencyAction(key, dependencyID), true)
			endpointsAction = true
		}
	}

	// See if a user needs to be detached from a component
	for dependencyID := range depKeysPrev {
		if !depKeysNext[dependencyID] {
			node.AddAction(component.NewDetachDependencyAction(key, dependencyID), true)
			endpointsAction = true
		}
	}

	// See if a component needs to be destructed
	if len(depKeysPrev) > 0 && len(depKeysNext) <= 0 {
		node.AddAction(component.NewDeleteAction(key), true)
		endpointsAction = false
	}

	// See if we need to retrieve component endpoints
	if endpointsAction && isCodeComponent {
		node.AddAction(component.NewEndpointsAction(key), true)
	}
}
