package visibility

import (
	"github.com/Aptomi/aptomi/pkg/slinga/engine"
)

// ConsumerView represents a view from a particular consumer(s) (service consumer point of view)
// TODO: userId and dependencyId must be userID and dependencyID (but it kinda breaks UI...)
type ConsumerView struct {
	userId       string
	dependencyId string
	state        engine.ServiceUsageState
	g            *graph
}

// NewConsumerView creates a new ConsumerView
func NewConsumerView(userID string, dependencyID string, state engine.ServiceUsageState) ConsumerView {
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
	for _, dependency := range view.state.Policy.Dependencies.DependenciesByID {
		if filterMatches(dependency.UserID, view.userId) && filterMatches(dependency.ID, view.dependencyId) {
			// Step 1 - add a node for every matching dependency found
			dependencyNode := newDependencyNode(dependency, false, view.state.GetUserLoader())
			view.g.addNode(dependencyNode, 0)

			// Step 2 - process subgraph (doesn't matter whether it's resolved successfully or not)
			view.addResolvedDependencies(dependency.ServiceKey, dependencyNode, 1)
		}
	}

	return view.g.GetData()
}

// Returns if value is good with respect to filterValue
func filterMatches(value string, filterValue string) bool {
	return len(filterValue) <= 0 || filterValue == value
}

// Adds to the graph nodes/edges which are triggered by usage of a given dependency
func (view ConsumerView) addResolvedDependencies(key string, nodePrev graphNode, nextLevel int) {
	// retrieve instance
	service, context, allocation, component := engine.ParseServiceUsageKey(key)

	// try to get this component instance from resolved data
	v := view.state.GetResolvedData().ComponentInstanceMap[key]

	// okay, this component likely failed to resolved, so let's look it up from unresolved pool
	if v == nil {
		v = view.state.UnresolvedData.ComponentInstanceMap[key]
	}

	// if it's a service, add node and connext with previous
	if component == engine.ComponentRootName {
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
