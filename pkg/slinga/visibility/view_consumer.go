package visibility

import (
	"github.com/Frostman/aptomi/pkg/slinga"
)

// ConsumerView represents a view from a particular consumer (service consumer point of view)
type ConsumerView struct {
	userId string
	state  slinga.ServiceUsageState
	g      *graph
}

// NewConsumerView creates a new ConsumerView
func NewConsumerView(userId string, state slinga.ServiceUsageState) ConsumerView {
	return ConsumerView{
		userId: userId,
		state:  state,
		g:      NewGraph(),
	}
}

// GetData returns graph for a given view
func (view ConsumerView) GetData() graphEntry {
	// go over all dependencies of a given user
	for _, dependency := range view.state.Dependencies.DependenciesByID {
		if dependency.UserID == view.userId {
			// Step 1 - add a node for every dependency found
			dependencyNode := newDependencyNode(dependency, false)
			view.g.addNode(dependencyNode, 0)
		}
	}

	return view.g.GetData()
}

/*
// Adds to the graph nodes/edges which trigger usage of a given service instance
func (view ConsumerView) addEveryoneWhoUses(serviceKey string, svcInstanceNodePrev graphNode, nextLevel int) {
	// retrieve service instance
	instance := view.state.GetResolvedUsage().ComponentInstanceMap[serviceKey]

	// if there are no incoming edges, it means we came to the very beginning of the chain (i.e. dependency)
	if len(instance.EdgesIn) <= 0 {
		// add nodes for all dependencies
		for _, dependencyID := range instance.DependencyIds {
			// add a node for dependency
			dependencyNode := newDependencyNode(view.state.Dependencies.DependenciesByID[dependencyID])
			view.g.addNode(dependencyNode, nextLevel)

			// connect prev service instance node and dependency node
			view.g.addEdge(svcInstanceNodePrev, dependencyNode)
		}
	} else {
		// go over all incoming edges
		for k := range instance.EdgesIn {
			service, context, allocation, component := slinga.ParseServiceUsageKey(k)
			v := view.state.GetResolvedUsage().ComponentInstanceMap[k]
			if component == slinga.ComponentRootName {
				// if it's a service instance, add a node
				svcInstanceNode := newServiceInstanceNode(k, view.state.Policy.Services[service], context, allocation, v, false)
				view.g.addNode(svcInstanceNode, nextLevel)

				// connect service instance nodes
				view.g.addEdge(svcInstanceNodePrev, svcInstanceNode)

				// proceed further with updated service instance node
				view.addEveryoneWhoUses(k, svcInstanceNode, nextLevel+1)
			} else {
				// proceed further, carry prev service instance node
				view.addEveryoneWhoUses(k, svcInstanceNodePrev, nextLevel)
			}
		}
	}
}
*/
