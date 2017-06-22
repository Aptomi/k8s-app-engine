package visibility

import (
	"github.com/Frostman/aptomi/pkg/slinga"
)

// ConsumerView represents a view from a particular consumer (service consumer point of view)
// TODO: userId and dependencyId must be userID and dependencyID (but it kinda breaks UI...)
type ConsumerView struct {
	userId       string
	dependencyId string
	state        slinga.ServiceUsageState
	g            *graph
}

// TODO: why the fuck NewConsumerView returns ConsumerView, and this returns graph...?!?!?!
func NewGlobalConsumerView(filterUserID string, users map[string]*slinga.User, state slinga.ServiceUsageState) graph {
	g := newGraph()
	for userID := range users {
		if filterUserID != "" && userID != filterUserID {
			continue
		}
		view := ConsumerView{
			userId:       userID,
			dependencyId: "",
			state:        state,
			g:            g,
		}
		view.GetData()
	}

	return *g
}

// NewConsumerView creates a new ConsumerView
func NewConsumerView(userID string, dependencyID string, state slinga.ServiceUsageState) ConsumerView {
	return ConsumerView{
		userId:       userID,
		dependencyId: dependencyID,
		state:        state,
		g:            newGraph(),
	}
}

// GetData returns graph for a given view
func (view ConsumerView) GetData() interface{} {
	// go over all dependencies of a given user
	for _, dependency := range view.state.Dependencies.DependenciesByID {
		if dependency.UserID == view.userId && (len(view.dependencyId) <= 0 || dependency.ID == view.dependencyId) {
			// Step 1 - add a node for every matching dependency found
			dependencyNode := newDependencyNode(dependency, false)
			view.g.addNode(dependencyNode, 0)

			// Step 2 - process subgraph (doesn't matter whether it's resolved successfully or not)
			view.addResolvedDependencies(dependency.ServiceKey, dependencyNode, 1)
		}
	}

	return view.g.GetData()
}

// Adds to the graph nodes/edges which are triggered by usage of a given dependency
func (view ConsumerView) addResolvedDependencies(key string, nodePrev graphNode, nextLevel int) {
	// retrieve instance
	service, context, allocation, component := slinga.ParseServiceUsageKey(key)

	// try to get this component instance from resolved data
	v := view.state.GetResolvedData().ComponentInstanceMap[key]

	// okay, this component likely failed to resolved, so let's look it up from unresolved pool
	if v == nil {
		v = view.state.UnresolvedData.ComponentInstanceMap[key]
	}

	// if it's a service, add node and connext with previous
	if component == slinga.ComponentRootName {
		// add service instance node
		svcInstanceNode := newServiceInstanceNode(key, view.state.Policy.Services[service], context, allocation, v, nextLevel <= 1)
		view.g.addNode(svcInstanceNode, nextLevel)

		// connect service instance nodes
		view.g.addEdge(nodePrev, svcInstanceNode)

		// update prev
		nodePrev = svcInstanceNode
	}

	// go over all outgoing edges
	for k := range v.EdgesOut {
		// proceed further with updated service instance node
		view.addResolvedDependencies(k, nodePrev, nextLevel+1)
	}
}
