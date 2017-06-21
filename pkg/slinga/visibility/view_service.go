package visibility

import (
	"github.com/Frostman/aptomi/pkg/slinga"
)

// ServiceView represents a view from a particular service (service owner point of view)
type ServiceView struct {
	serviceName string
	state       slinga.ServiceUsageState
	g           *graph
}

// NewServiceView creates a new ServiceView
func NewServiceView(serviceName string, state slinga.ServiceUsageState) ServiceView {
	return ServiceView{
		serviceName: serviceName,
		state:       state,
		g:           newGraph(),
	}
}

// GetData returns graph for a given view
func (view ServiceView) GetData() interface{} {
	// Step 1 - add a node with a given service
	svcNode := newServiceNode(view.serviceName)
	view.g.addNode(svcNode, 0)

	// Step 2 - find all instances of a given service. add them as "instance nodes"
	for k, v := range view.state.ResolvedData.ComponentInstanceMap {
		service, context, allocation, component := slinga.ParseServiceUsageKey(k)
		if service == view.serviceName && component == slinga.ComponentRootName {
			// add a node with an instance of our service
			svcInstanceNode := newServiceInstanceNode(k, view.state.Policy.Services[service], context, allocation, v, true)
			view.g.addNode(svcInstanceNode, 1)

			// connect service node and instance node
			view.g.addEdge(svcNode, svcInstanceNode)

			// Step 3 - from a given instance of a service, go and add everyone who is using this service
			view.addEveryoneWhoUses(k, svcInstanceNode, 2)
		}
	}

	return view.g.GetData()
}

// Adds to the graph nodes/edges which trigger usage of a given service instance
func (view ServiceView) addEveryoneWhoUses(serviceKey string, svcInstanceNodePrev graphNode, nextLevel int) {
	// retrieve service instance
	instance := view.state.GetResolvedData().ComponentInstanceMap[serviceKey]

	// if there are no incoming edges, it means we came to the very beginning of the chain (i.e. dependency)
	if len(instance.EdgesIn) <= 0 {
		// add nodes for all dependencies
		for dependencyID := range instance.DependencyIds {
			// add a node for dependency
			dependencyNode := newDependencyNode(view.state.Dependencies.DependenciesByID[dependencyID], true)
			view.g.addNode(dependencyNode, nextLevel)

			// connect prev service instance node and dependency node
			view.g.addEdge(svcInstanceNodePrev, dependencyNode)
		}
	} else {
		// go over all incoming edges
		for k := range instance.EdgesIn {
			service, context, allocation, component := slinga.ParseServiceUsageKey(k)
			v := view.state.GetResolvedData().ComponentInstanceMap[k]
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
