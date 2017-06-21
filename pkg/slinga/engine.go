package slinga

import (
	"fmt"
)

/*
	Core engine for Slinga processing and evaluation
*/

// ResolveAllDependencies evaluates and resolves all recorded dependencies ("<user> needs <service> with <labels>"), calculating component allocations
func (state *ServiceUsageState) ResolveAllDependencies() error {

	// Run every declared dependency via policy and resolve it
	for _, dependencies := range state.Dependencies.DependenciesByService {
		for _, d := range dependencies {
			node := state.newResolutionNode(d)

			// resolve usage via applying policy
			err := state.resolveDependency(node)

			// see if there is an error
			if err != nil {
				return err
			}

			// record and put usages in the right place
			d.Resolved = node.resolved
			d.ServiceKey = node.serviceKey
			if node.resolved {
				state.ResolvedData.appendData(node.data)
			} else {
				state.UnresolvedData.appendData(node.data)
			}
		}
	}
	return nil
}

// Evaluate evaluates and resolves a single dependency ("<user> needs <service> with <labels>") and calculates component allocations
func (state *ServiceUsageState) resolveDependency(node *resolutionNode) error {
	// Error variable that we will be reusing
	var err error

	// Indicate that we are starting to resolve dependency
	node.debugResolvingDependencyStart()

	// Locate the service
	node.service, err = node.getMatchedService(state.Policy)
	if err != nil {
		return err
	}

	// Process service and transform labels
	node.labels = node.transformLabels(node.labels, node.service.Labels)

	// Match the context
	node.context, err = node.getMatchedContext(state.Policy)
	if err != nil {
		return err
	}
	// If no matching context is found, let's just exit
	if node.context == nil {
		return node.cannotResolve()
	}

	// Process context and transform labels
	node.labels = node.transformLabels(node.labels, node.context.Labels)

	// Match the allocation
	node.allocation, err = node.getMatchedAllocation(state.Policy)
	if err != nil {
		return err
	}
	// If no matching allocation is found, let's just exit
	if node.allocation == nil {
		return node.cannotResolve()
	}

	// Process allocation, transform
	node.labels = node.transformLabels(node.labels, node.allocation.Labels)

	// Create service key
	node.serviceKey = createServiceUsageKey(node.service, node.context, node.allocation, nil)

	// Once instance is figured out, make sure to attach rule logs to that instance
	node.ruleLogWriter.attachToInstance(node.serviceKey)

	// Store labels for service
	node.data.recordLabels(node.serviceKey, node.labels)

	// Store edge (last component instance -> service instance)
	node.data.storeEdge(node.arrivalKey, node.serviceKey)

	// Now, sort all components in topological order
	componentsOrdered, err := node.service.getComponentsSortedTopologically()
	if err != nil {
		return err
	}

	// Iterate over all service components and resolve them recursively
	// Note that discovery variables can refer to other variables announced by dependents in the discovery tree
	for _, node.component = range componentsOrdered {
		// Create key
		node.componentKey = createServiceUsageKey(node.service, node.context, node.allocation, node.component)

		// Store edge (service instance -> component instance)
		node.data.storeEdge(node.serviceKey, node.componentKey)

		// Calculate and store labels for component
		node.componentLabels = node.transformLabels(node.labels, node.component.Labels)
		node.data.recordLabels(node.componentKey, node.componentLabels)

		// Create new map with resolution keys for component
		node.discoveryTreeNode[node.component.Name] = NestedParameterMap{}

		// Calculate and store discovery params
		err := node.calculateAndStoreDiscoveryParams()
		if err != nil {
			return err
		}

		// Print information that we are starting to resolve dependency (on code, or on service)
		node.debugResolvingDependencyOnComponent()

		if node.component.Code != nil {
			// Evaluate code params
			err := node.calculateAndStoreCodeParams()
			if err != nil {
				return err
			}
		} else if node.component.Service != "" {
			// Create a child node for dependency resolution
			nodeNext := node.createChildNode()

			// Resolve dependency recursively
			err := state.resolveDependency(nodeNext)
			if err != nil {
				return err
			}

			// if a dependency has not been fulfilled, then exit
			if !nodeNext.resolved {
				return node.cannotResolve()
			}
		}

		// Record usage of a given component
		node.data.recordResolvedAndDependency(node.componentKey, node.dependency)
		node.data.recordProcessingOrder(node.componentKey)
	}

	// Mark object as resolved and record usage of a given service
	node.resolved = true
	node.data.recordResolvedAndDependency(node.serviceKey, node.dependency)
	node.data.recordProcessingOrder(node.serviceKey)

	// Indicate that we have resolved dependency
	node.debugResolvingDependencyEnd()

	return nil
}

// Topologically sort components and return true if there is a cycle detected
func (service *Service) dfsComponentSort(u *ServiceComponent, colors map[string]int) error {
	colors[u.Name] = 1

	for _, vName := range u.Dependencies {
		v, exists := service.getComponentsMap()[vName]
		if !exists {
			return fmt.Errorf("Service %s has a dependency to non-existing component %s", service.Name, vName)
		}
		if vColor, ok := colors[v.Name]; !ok {
			// not visited yet -> visit and exit if a cycle was found or another error occured
			if err := service.dfsComponentSort(v, colors); err != nil {
				return err
			}
		} else if vColor == 1 {
			return fmt.Errorf("Component cycle detected while processing service %s", service.Name)
		}
	}

	service.componentsOrdered = append(service.componentsOrdered, u)
	colors[u.Name] = 2
	return nil
}

// Sorts all components in a topological way
func (service *Service) getComponentsSortedTopologically() ([]*ServiceComponent, error) {
	if service.componentsOrdered == nil {
		// Initiate colors
		colors := make(map[string]int)

		// Dfs
		for _, c := range service.Components {
			if _, ok := colors[c.Name]; !ok {
				if err := service.dfsComponentSort(c, colors); err != nil {
					return nil, err
				}
			}
		}
	}

	return service.componentsOrdered, nil
}
