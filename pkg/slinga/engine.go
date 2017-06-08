package slinga

import (
	"errors"
	log "github.com/Sirupsen/logrus"
)

/*
	Core engine for Slinga processing and evaluation
*/

// ResolveAllDependencies evaluates and resolves all recorded dependencies ("<user> needs <service> with <labels>"), calculating component allocations
func (usage *ServiceUsageState) ResolveAllDependencies() error {

	// Run every declared dependency via policy and resolve it
	for _, dependencies := range usage.Dependencies.Dependencies {
		for _, d := range dependencies {
			node := usage.newResolutionNode(d)

			// see if it needs to be traced (addl debug output on console)
			tracing.setEnable(d.Trace)

			// resolve usage via applying policy
			// TODO: if a dependency cannot be fulfilled, we need to handle it correctly. i.e. usages should be recorded in different context (not in usage.DiscoveryTree and not even in usage) and not applied
			err := usage.resolveDependency(node, usage.ResolvedUsage)

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
func (usage *ServiceUsageState) resolveDependency(node *resolutionNode, resolvedUsage ResolvedServiceUsageData) error {
	// Error variable that we will be reusing
	var err error

	// Print information that we are starting to resolve dependency on service
	node.debugResolvingDependency()

	// Locate the service
	node.service, err = usage.Policy.getService(node.serviceName)
	if err != nil {
		return node.errorDuringServiceLookup(err)
	}

	// Process service and transform labels
	node.labels = node.labels.applyTransform(node.service.Labels)
	node.debugNewLabels()

	// Match the context
	node.context, err = node.getMatchedContext(usage.Policy)
	if err != nil {
		return node.errorDuringMatchingContext(err)
	}
	// If no matching context is found, let's just exit
	if node.context == nil {
		return nil
	}

	// Print information that we are starting to resolve context
	node.debugResolvingContext()

	// Process context and transform labels
	node.labels = node.labels.applyTransform(node.context.Labels)
	node.debugNewLabels()

	// Match the allocation
	node.allocation, err = node.getMatchedAllocation(usage.Policy)
	if err != nil {
		return node.errorDuringMatchingAllocation(err)
	}
	// If no matching allocation is found, let's just exit
	if node.allocation == nil {
		return nil
	}

	// Print information that we are starting to resolve allocation
	node.debugResolvingAllocation()

	// Process allocation and transform labels
	node.labels = node.labels.applyTransform(node.allocation.Labels)
	node.debugNewLabels()

	// Now, sort all components in topological order
	componentsOrdered, err := node.service.getComponentsSortedTopologically()
	if err != nil {
		return node.errorDuringSortingComponentsTopologically(err)
	}

	// Iterate over all service components and resolve them recursively
	// Note that discovery variables can refer to other variables announced by dependents in the discovery tree
	for _, node.component = range componentsOrdered {
		// Create key
		node.componentKey = createServiceUsageKey(node.service, node.context, node.allocation, node.component)

		// Calculate and store labels
		node.componentLabels = node.labels.applyTransform(node.component.Labels)
		resolvedUsage.storeLabels(node.componentKey, node.componentLabels)

		// Create new map with resolution keys for component
		node.discoveryTreeNode[node.component.Name] = NestedParameterMap{}

		// Calculate and store discovery params
		err := node.calculateAndStoreDiscoveryParams(resolvedUsage)
		if err != nil {
			return err
		}

		if node.component.Code != nil {
			// Print information that we are starting to resolve dependency on code
			node.debugResolvingDependencyOnCode()

			// Evaluate code params
			err := node.calculateAndStoreCodeParams(resolvedUsage)
			if err != nil {
				return err
			}
		} else if node.component.Service != "" {
			// Print information that we are starting to resolve dependency on another service
			node.debugResolvingDependencyOnService()

			// Create a child node for dependency resolution
			nodeNext := node.createChildNode()

			// Resolve dependency recursively
			err := usage.resolveDependency(nodeNext, resolvedUsage)
			if err != nil {
				return err
			}

			// if a dependency has not been fulfilled, then exit
			if !nodeNext.resolved {
				return node.cannotResolveDependency()
			}
		} else {
			node.errorInvalidComponent()
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
func (service *Service) dfsComponentSort(u *ServiceComponent, colors map[string]int) bool {
	colors[u.Name] = 1

	for _, vName := range u.Dependencies {
		v, exists := service.getComponentsMap()[vName]
		if !exists {
			debug.WithFields(log.Fields{
				"service":   service.Name,
				"component": vName,
			}).Fatal("Service dependency points to non-existing component")
		}
		if vColor, ok := colors[v.Name]; !ok {
			// not visited yet -> visit and exit if a cycle was found
			if service.dfsComponentSort(v, colors) {
				return true
			}
		} else if vColor == 1 {
			return true
		}
	}

	service.componentsOrdered = append(service.componentsOrdered, u)
	colors[u.Name] = 2
	return false
}

// Sorts all components in a topological way
func (service *Service) getComponentsSortedTopologically() ([]*ServiceComponent, error) {
	if service.componentsOrdered == nil {
		// Initiate colors
		colors := make(map[string]int)

		// Dfs
		var cycle = false
		for _, c := range service.Components {
			if _, ok := colors[c.Name]; !ok {
				if service.dfsComponentSort(c, colors) {
					cycle = true
					break
				}
			}
		}

		if cycle {
			return nil, errors.New("Component cycle detected in service " + service.Name)
		}
	}

	return service.componentsOrdered, nil
}

// Helper to get a service
func (policy *Policy) getService(serviceName string) (*Service, error) {
	// Locate the service
	service := policy.Services[serviceName]
	if service == nil {
		return nil, errors.New("Service " + serviceName + " not found")
	}
	return service, nil
}
