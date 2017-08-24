package engine

import (
	. "github.com/Aptomi/aptomi/pkg/slinga/util"
	"github.com/Aptomi/aptomi/pkg/slinga/language"
)

/*
	Core engine for Slinga processing and evaluation
*/

// ResolveAllDependencies evaluates and resolves all recorded dependencies ("<user> needs <service> with <labels>"), calculating component allocations
func (state *ServiceUsageState) ResolveAllDependencies() error {

	// Create cache
	cache := NewEngineCache()

	// Run every declared dependency via policy and resolve it
	for _, dependencies := range state.Policy.Dependencies.DependenciesByService {
		for _, d := range dependencies {
			// resolve usage via applying policy
			err := state.resolveDependency(d, cache)

			// see if there is an error
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Resolves a single dependency and puts resolution data into the overall state of the world
func (state *ServiceUsageState) resolveDependency(d *language.Dependency, cache *EngineCache) error {
	// create resolution node
	node := state.newResolutionNode(d, cache)

	// recursively resolve everything
	err := state.resolveNode(node)
	if err != nil {
		return err
	}

	// add dependency resolution data to the rest of the records
	d.Resolved = node.resolved
	d.ServiceKey = node.serviceKey.GetKey()

	var data *ServiceUsageData
	if node.resolved {
		data = state.ResolvedData
	} else {
		data = state.UnresolvedData
	}

	err = data.appendData(node.data)
	if err != nil {
		node.logError(err)
		return err
	}

	return nil
}

// Evaluate evaluates and resolves a single dependency ("<user> needs <service> with <labels>") and calculates component allocations
// Returns error only if there is an issue with the policy (e.g. it's malformed)
// Returns nil if there is no error (it may be that nothing was still matched though)
// If you want to check for successful resolution, use node.resolved flag
func (state *ServiceUsageState) resolveNode(node *resolutionNode) error {
	// Error variable that we will be reusing
	var err error

	// Indicate that we are starting to resolve dependency
	node.objectResolved(node.dependency)
	node.logStartResolvingDependency()

	// Locate the user
	err = node.checkUserExists()
	if err != nil {
		// If consumer is not present, let's just say that this dependency cannot be fulfilled
		return node.cannotResolveInstance(err)
	}
	node.objectResolved(node.user)

	// Locate the service
	node.service, err = node.getMatchedService(state.Policy)
	if err != nil {
		// Return a policy processing error in case service is not present in policy
		return node.cannotResolveInstance(err)
	}
	node.objectResolved(node.service)

	// Process service and transform labels
	node.labels = node.transformLabels(node.labels, node.service.ChangeLabels)

	// Match the context
	node.context, err = node.getMatchedContext(state.Policy)
	if err != nil {
		// Return a policy processing error in case of context resolution failure
		return node.cannotResolveInstance(err)
	}

	// If no matching context is found, the dependency cannot be resolved
	if node.context == nil {
		// This is considered a normal scenario (no matching context found), so no error is returned
		return node.cannotResolveInstance(nil)
	}
	node.objectResolved(node.context)

	// Process context and transform labels
	node.labels = node.transformLabels(node.labels, node.context.ChangeLabels)

	// Resolve allocation keys for the context
	node.allocationKeysResolved, err = node.resolveAllocationKeys(state.Policy)
	if err != nil {
		// Return an error in case of malformed policy or policy processing error
		return node.cannotResolveInstance(err)
	}

	// Create service key
	node.serviceKey = node.createComponentKey(nil)
	node.objectResolved(node.serviceKey)

	// Store labels for service
	node.recordLabels(node.serviceKey, node.labels)

	// Store edge (last component instance -> service instance)
	node.data.storeEdge(node.arrivalKey, node.serviceKey)

	// Now, sort all components in topological order
	componentsOrdered, err := node.sortServiceComponents()
	if err != nil {
		// Return an error in case of failed component topological sort
		return node.cannotResolveInstance(err)
	}

	// Iterate over all service components and resolve them recursively
	// Note that discovery variables can refer to other variables announced by dependents in the discovery tree
	for _, node.component = range componentsOrdered {
		// Create key
		node.componentKey = node.createComponentKey(node.component)

		// Store edge (service instance -> component instance)
		node.data.storeEdge(node.serviceKey, node.componentKey)

		// Calculate and store labels for component
		node.componentLabels = node.transformLabels(node.labels, node.component.ChangeLabels)
		node.recordLabels(node.componentKey, node.componentLabels)

		// Create new map with resolution keys for component
		node.discoveryTreeNode[node.component.Name] = NestedParameterMap{}

		// Calculate and store discovery params
		err := node.calculateAndStoreDiscoveryParams()
		if err != nil {
			return node.cannotResolveInstance(err)
		}

		// Print information that we are starting to resolve dependency (on code, or on service)
		node.logResolvingDependencyOnComponent()

		if node.component.Code != nil {
			// Evaluate code params
			err := node.calculateAndStoreCodeParams()
			if err != nil {
				return node.cannotResolveInstance(err)
			}
		} else if node.component.Service != "" {
			// Create a child node for dependency resolution
			nodeNext := node.createChildNode()

			// Resolve dependency on another service recursively
			err := state.resolveNode(nodeNext)
			if err != nil {
				return node.cannotResolveInstance(err)
			}

			// If a sub-dependency has not been fulfilled, then exit
			if !nodeNext.resolved {
				// This is considered a normal scenario (sub-dependency not fulfilled), so no error is returned
				return node.cannotResolveInstance(nil)
			}
		}

		// Record usage of a given component instance
		node.recordResolved(node.componentKey, node.dependency)
	}

	// Mark note as resolved and record usage of a given service instance
	node.resolved = true
	node.recordResolved(node.serviceKey, node.dependency)

	return nil
}
