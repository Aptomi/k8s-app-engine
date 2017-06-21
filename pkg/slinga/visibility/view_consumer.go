package visibility

import (
	"github.com/Frostman/aptomi/pkg/slinga"
)

// ConsumerView represents a view from a particular consumer (service consumer point of view)
type ConsumerView struct {
	userId       string
	dependencyId string
	state        slinga.ServiceUsageState
	g            *graph
}

// NewConsumerView creates a new ConsumerView
func NewConsumerView(userId string, dependencyId string, state slinga.ServiceUsageState) ConsumerView {
	return ConsumerView{
		userId:       userId,
		dependencyId: dependencyId,
		state:        state,
		g:            NewGraph(),
	}
}

// GetData returns graph for a given view
func (view ConsumerView) GetData() graphEntry {
	// go over all dependencies of a given user
	for _, dependency := range view.state.Dependencies.DependenciesByID {
		if dependency.UserID == view.userId && (len(view.dependencyId) <= 0 || dependency.ID == view.dependencyId) {
			// Step 1 - add a node for every matching dependency found
			dependencyNode := newDependencyNode(dependency, false)
			view.g.addNode(dependencyNode, 0)

			// Step 2 - process subgraph
			if len(dependency.ResolvesTo) > 0 {
				view.addResolvedDependencies(dependency.ResolvesTo, dependencyNode, 1)
			}
		}
	}

	return view.g.GetData()
}

// Adds to the graph nodes/edges which are triggered by usage of a given dependency
func (view ConsumerView) addResolvedDependencies(key string, nodePrev graphNode, nextLevel int) {
	// retrieve instance
	service, context, allocation, component := slinga.ParseServiceUsageKey(key)
	v := view.state.GetResolvedUsage().ComponentInstanceMap[key]

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
