package diff

import (
	"github.com/Aptomi/aptomi/pkg/engine/apply/action"
	"github.com/Aptomi/aptomi/pkg/engine/apply/action/component"
	"github.com/Aptomi/aptomi/pkg/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/util"
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
	// Produce a map of all component instances
	allCompInstances := make(map[string]bool)
	for keyPrev := range diff.Prev.ComponentInstanceMap {
		allCompInstances[keyPrev] = true
	}
	for keyNext := range diff.Next.ComponentInstanceMap {
		allCompInstances[keyNext] = true
	}

	// Build a flat list of actions for every component instance
	for key := range allCompInstances {
		diff.buildActions(key)
	}

	// Generate dependencies between actions
	for key := range allCompInstances {
		outgoing := make(map[string]bool)
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
func (diff *PolicyResolutionDiff) buildActions(key string) { // nolint: gocyclo
	// Get action graph node for a given component key
	node := diff.ActionPlan.GetActionGraphNode(key)

	// Get previous claim keys
	var claimKeysPrev map[string]int
	prevInstance := diff.Prev.ComponentInstanceMap[key]
	if prevInstance != nil {
		claimKeysPrev = prevInstance.ClaimKeys
	}

	// Get next claim keys
	var claimKeysNext map[string]int
	nextInstance := diff.Next.ComponentInstanceMap[key]
	if nextInstance != nil {
		claimKeysNext = nextInstance.ClaimKeys
	}

	/*
		First of all, let's see if a component needs to be destructed. If so, destruct it and don't proceed to any further actions.
	*/

	// See if a claim needs to be detached from a component
	for claimKey := range claimKeysPrev {
		if _, found := claimKeysNext[claimKey]; !found {
			node.AddAction(component.NewDetachClaimAction(key, claimKey), diff.Prev, true)
		}
	}

	// See if a component needs to be destructed
	if len(claimKeysPrev) > 0 && len(claimKeysNext) <= 0 {
		node.AddAction(component.NewDeleteAction(key, prevInstance.CalculatedCodeParams), diff.Prev, true)
		return // exit right away
	}

	/*
		Now, let's see if a component needs to be created or updated.
	*/

	// See if it's a bundle or component
	isCodeComponent := (prevInstance != nil && prevInstance.IsCode) || (nextInstance != nil && nextInstance.IsCode)

	// See if a component needs to be instantiated
	if len(claimKeysPrev) <= 0 && len(claimKeysNext) > 0 {
		node.AddAction(component.NewCreateAction(key, nextInstance.CalculatedCodeParams), diff.Prev, true)
	}

	// See if a component needs to be updated
	if isCodeComponent && len(claimKeysPrev) > 0 && len(claimKeysNext) > 0 {
		sameParams := prevInstance.CalculatedCodeParams.DeepEqual(nextInstance.CalculatedCodeParams)
		if !sameParams {
			node.AddAction(component.NewUpdateAction(key, prevInstance.CalculatedCodeParams, nextInstance.CalculatedCodeParams), diff.Prev, true)

			// indicate that a parent bundle component instance gets updated as well
			// this is required for adjusting update/creation times of a bundle with changed component
			// this may produce duplicate "update" actions for the parent bundle
			bundleKey := nextInstance.Metadata.Key.GetParentBundleKey().GetKey()
			bundleNode := diff.ActionPlan.GetActionGraphNode(bundleKey)
			bundleNode.AddAction(component.NewUpdateAction(bundleKey, util.NestedParameterMap{}, util.NestedParameterMap{}), diff.Prev, true)
		}
	}

	// See if a claim needs to be attached to a component
	for claimKey, depth := range claimKeysNext {
		if _, found := claimKeysPrev[claimKey]; !found {
			node.AddAction(component.NewAttachClaimAction(key, claimKey, depth), diff.Prev, true)
		}
	}
}
