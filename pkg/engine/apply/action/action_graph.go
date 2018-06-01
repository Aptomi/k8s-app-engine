package action

import "github.com/Aptomi/aptomi/pkg/engine/resolve"

// GraphNode represents a node in the graph of apply actions
type GraphNode struct {
	// Key is unique identifier of the node
	Key string

	// Before is a map of actions which have to be executed before main actions of this node. If one of them fails,
	// main actions will not be executed
	Before []*GraphNode

	// BeforeRev is the opposite to Before, indicating which actions have to be executed after this node finishes
	// execution
	BeforeRev []*GraphNode

	// Main actions which have to be executed sequentially. If one fails, the rest will not be executed
	Actions []Base
}

// NewGraphNode creates a new GraphNode of apply actions
func NewGraphNode(key string) *GraphNode {
	return &GraphNode{
		Key: key,
	}
}

// AddBefore adds an action to be executed before the list of main actions
func (node *GraphNode) AddBefore(that *GraphNode) {
	node.Before = append(node.Before, that)
	that.BeforeRev = append(that.BeforeRev, node)
}

// AddAction adds an action to the list of main actions. If avoidDuplicates is true, then duplicate actions will not be
// added (e.g. update action)
func (node *GraphNode) AddAction(action Base, actualState *resolve.PolicyResolution, avoidDuplicates bool) {
	add := true
	if avoidDuplicates {
		// go over existing actions and make sure we don't add duplicates
		for _, existing := range node.Actions {
			if existing.GetName() == action.GetName() {
				add = false
				break
			}
		}
	}

	if add {
		// call AfterCreated on the action
		action.AfterCreated(actualState)

		// schedule the action
		node.Actions = append(node.Actions, action)
	}
}
