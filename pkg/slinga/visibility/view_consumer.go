package visibility

import (
	"github.com/Frostman/aptomi/pkg/slinga"
)

// ConsumerView represents a view from a particular consumer (service consumer point of view)
type ConsumerView struct {
	userID       string
	dependencyID string
	state        slinga.ServiceUsageState
	g            *graph
}

func NewGlobalView(filterUserId string, users map[string]*slinga.User, state slinga.ServiceUsageState) graph {
	g := newGraph()
	for userId := range users {
		if userId != "" && userId != filterUserId {
			continue
		}
		view := ConsumerView{
			userID:       userId,
			dependencyID: "",
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
		userID:       userID,
		dependencyID: dependencyID,
		state:        state,
		g:            newGraph(),
	}
}

// GetData returns graph for a given view
func (view ConsumerView) GetData() interface{} {
	// go over all dependencies of a given user
	for _, dependency := range view.state.Dependencies.DependenciesByID {
		if dependency.UserID == view.userID && (len(view.dependencyID) <= 0 || dependency.ID == view.dependencyID) {
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
