package slinga

import (
	"fmt"
)

/*
	Core engine for Slinga processing and evaluation
*/

// ResolveAllDependencies evaluates and resolves all recorded dependencies ("<user> needs <service> with <labels>"), calculating component allocations
func (usage *ServiceUsageState) ResolveAllDependencies(dir string) error {

	// Run every declared dependency via policy and resolve it
	for _, dependencies := range usage.Dependencies.Dependencies {
		for _, d := range dependencies {
			node := usage.newResolutionNode(d, dir)

			// see if it needs to be traced (addl debug output on console)
			tracing.setEnable(d.Trace)

			// resolve usage via applying policy
			// TODO: if a dependency cannot be fulfilled, we need to handle it correctly. i.e. usages should be recorded in different context and not applied
			err := usage.resolveDependency(node, usage.ResolvedUsage, dir)

			// disable tracing
			tracing.setEnable(false)

			// see if there is an error
			if err != nil {
				return err
			}

			// record high-level service resolution
			d.ResolvesTo = node.debugResolvedKey
		}
	}
	return nil
}

// Evaluate evaluates and resolves a single dependency ("<user> needs <service> with <labels>") and calculates component allocations
func (usage *ServiceUsageState) resolveDependency(node *resolutionNode, resolvedUsage *ResolvedServiceUsageData, dir string) error {
	// Error variable that we will be reusing
	var err error

	// Print information that we are starting to resolve dependency on service
	node.debugResolvingDependency()

	// Locate the service
	node.service, err = node.getMatchedService(usage.Policy)
	if err != nil {
		return err
	}

	// Process service and transform labels
	node.labels = node.transformLabels(node.labels, node.service.Labels)

	// Match the context
	node.context, err = node.getMatchedContext(usage.Policy)
	if err != nil {
		return err
	}
	// If no matching context is found, let's just exit
	if node.context == nil {
		return nil
	}

	// Print information that we are starting to resolve context
	node.debugResolvingContext()

	// Process context and transform labels
	node.labels = node.transformLabels(node.labels, node.context.Labels)

	// Match the allocation
	node.allocation, err = node.getMatchedAllocation(usage.Policy)
	if err != nil {
		return err
	}
	// If no matching allocation is found, let's just exit
	if node.allocation == nil {
		return nil
	}

	// Print information that we are starting to resolve allocation
	node.debugResolvingAllocation()

	// Process allocation and transform labels
	node.labels = node.transformLabels(node.labels, node.allocation.Labels)

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

		// Calculate and store labels
		node.componentLabels = node.transformLabels(node.labels, node.component.Labels)
		resolvedUsage.storeLabels(node.componentKey, node.componentLabels)

		// Create new map with resolution keys for component
		node.discoveryTreeNode[node.component.Name] = NestedParameterMap{}

		// Calculate and store discovery params
		err := node.calculateAndStoreDiscoveryParams(resolvedUsage)
		if err != nil {
			return err
		}

		// Print information that we are starting to resolve dependency (on code, or on service)
		node.debugResolvingDependencyOnComponent()

		if node.component.Code != nil {
			// Evaluate code params
			err := node.calculateAndStoreCodeParams(resolvedUsage)
			if err != nil {
				return err
			}
		} else if node.component.Service != "" {
			// Create a child node for dependency resolution
			nodeNext := node.createChildNode()

			// Resolve dependency recursively
			err := usage.resolveDependency(nodeNext, resolvedUsage, dir)
			if err != nil {
				return err
			}

			// if a dependency has not been fulfilled, then exit
			if !nodeNext.resolved {
				return node.cannotResolveDependency()
			}
		}

		// Record usage of a given component
		resolvedUsage.recordUsage(node.componentKey, node.user)
	}

	// Record usage of a given service
	node.serviceKey = createServiceUsageKey(node.service, node.context, node.allocation, nil)
	resolvedUsage.recordUsage(node.serviceKey, node.user)

	// Mark object as successfully resolved
	node.resolved = true
	node.debugResolvedKey = node.context.Name + "#" + node.allocation.NameResolved
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
