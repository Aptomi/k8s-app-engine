package engine

import (
	. "github.com/Aptomi/aptomi/pkg/slinga/util"
	. "github.com/Aptomi/aptomi/pkg/slinga/eventlog"
)

/*
	Core engine for Slinga processing and evaluation
*/

// ResolveAllDependencies evaluates and resolves all recorded dependencies ("<user> needs <service> with <labels>"), calculating component allocations
func (state *ServiceUsageState) ResolveAllDependencies() error {

	// Run every declared dependency via policy and resolve it
	cache := NewEngineCache()
	eventLog := NewEventLog()

	// TODO: create new event log for every node. attach to different instances. then merge then
	for _, dependencies := range state.Policy.Dependencies.DependenciesByService {
		for _, d := range dependencies {
			// create resolution node
			node := state.newResolutionNode(d, cache, eventLog)

			// resolve usage via applying policy
			err := state.resolveDependency(node)

			// see if there is an error
			if err != nil {
				return err
			}

			// record and put usages in the right place
			d.Resolved = node.resolved
			d.ServiceKey = node.serviceKey.GetKey()
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
// Returns error only if there is an issue with the policy (e.g. it's malformed)
// Returns nil if there is no error (it may be that nothing was still matched though)
// If you want to check for successful resolution, use node.resolved flag
func (state *ServiceUsageState) resolveDependency(node *resolutionNode) error {
	// Error variable that we will be reusing
	var err error

	// Indicate that we are starting to resolve dependency
	node.logStartResolvingDependency()

	// Locate the user
	err = node.checkUserExists()
	if err != nil {
		// If consumer is not present, let's just say that this dependency cannot be fulfilled
		node.logCannotResolveInstance()
		return nil
	}

	// Locate the service
	node.service, err = node.getMatchedService(state.Policy)
	if err != nil {
		// Return a policy processing error in case service is not present in policy
		node.logCannotResolveInstance()
		return err
	}

	// Process service and transform labels
	node.labels = node.transformLabels(node.labels, node.service.ChangeLabels)

	// Match the context
	node.context, err = node.getMatchedContext(state.Policy)
	if err != nil {
		// Return a policy processing error in case of context resolution failure
		node.logCannotResolveInstance()
		return err
	}

	// If no matching context is found, the dependency cannot be resolved
	if node.context == nil {
		// This is considered a normal scenario (context not found), so no error is returned
		node.logCannotResolveInstance()
		return nil
	}

	// Process context and transform labels
	node.labels = node.transformLabels(node.labels, node.context.ChangeLabels)

	// Resolve allocation keys for the context
	node.allocationKeysResolved, err = node.resolveAllocationKeys(state.Policy)
	if err != nil {
		// Return an error in case of malformed policy or policy processing error
		node.logCannotResolveInstance()
		return err
	}

	// Create service key
	node.serviceKey = node.createComponentKey(nil)

	// Once instance is figured out, make sure to attach rule logs to that instance
	node.ruleLogWriter.attachToInstance(node.serviceKey)

	// Store labels for service
	node.data.recordLabels(node.serviceKey, node.labels)

	// Store edge (last component instance -> service instance)
	node.data.storeEdge(node.arrivalKey, node.serviceKey)

	// Now, sort all components in topological order
	componentsOrdered, err := node.sortServiceComponents()
	if err != nil {
		// Return an error in case of failed component topological sort
		node.logCannotResolveInstance()
		return err
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
		node.data.recordLabels(node.componentKey, node.componentLabels)

		// Create new map with resolution keys for component
		node.discoveryTreeNode[node.component.Name] = NestedParameterMap{}

		// Calculate and store discovery params
		err := node.calculateAndStoreDiscoveryParams()
		if err != nil {
			node.logCannotResolveInstance()
			return err
		}

		// Print information that we are starting to resolve dependency (on code, or on service)
		node.logResolvingDependencyOnComponent()

		if node.component.Code != nil {
			// Evaluate code params
			err := node.calculateAndStoreCodeParams()
			if err != nil {
				node.logCannotResolveInstance()
				return err
			}
		} else if node.component.Service != "" {
			// Create a child node for dependency resolution
			nodeNext := node.createChildNode()

			// Resolve dependency on another service recursively
			err := state.resolveDependency(nodeNext)
			if err != nil {
				node.logCannotResolveInstance()
				return err
			}

			// If a sub-dependency has not been fulfilled, then exit
			if !nodeNext.resolved {
				// This is considered a normal scenario (sub-dependency not fulfilled), so no error is returned
				node.logCannotResolveInstance()
				return nil
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

