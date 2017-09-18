package resolve

import (
	"fmt"
	. "github.com/Aptomi/aptomi/pkg/slinga/eventlog"
	"github.com/Aptomi/aptomi/pkg/slinga/external"
	"github.com/Aptomi/aptomi/pkg/slinga/language"
	. "github.com/Aptomi/aptomi/pkg/slinga/language"
	"github.com/Aptomi/aptomi/pkg/slinga/language/expression"
	"github.com/Aptomi/aptomi/pkg/slinga/language/template"
	. "github.com/Aptomi/aptomi/pkg/slinga/util"
	"sync"
)

/*
	Core engine for policy resolution
	- takes policy an an input
	- calculates PolicyResolution as an output
*/

const THREAD_POOL_SIZE = 8

type PolicyResolver struct {
	/*
		Input objects
	*/

	// Policy
	policy *Policy

	// External data
	externalData *external.Data

	/*
		Cache
	*/

	// Expression cache
	expressionCache *expression.ExpressionCache

	// Template cache
	templateCache *template.TemplateCache

	/*
		Calculated objects (aggregated over all dependencies)
	*/

	combineMutex sync.Mutex

	// Reference to the calculated PolicyResolution
	resolution *PolicyResolution

	// Buffered event log - gets populated during policy resolution
	eventLog *EventLog
}

// NewPolicyResolver creates a new policy resolver
func NewPolicyResolver(policy *Policy, externalData *external.Data) *PolicyResolver {
	return &PolicyResolver{
		policy:          policy,
		externalData:    externalData,
		expressionCache: expression.NewExpressionCache(),
		templateCache:   template.NewTemplateCache(),
		resolution:      NewPolicyResolution(),
		eventLog:        NewEventLog(),
	}
}

// ResolveAllDependencies evaluates and resolves all recorded dependencies ("<user> needs <service> with <labels>"), calculating component allocations
func (resolver *PolicyResolver) ResolveAllDependencies() (*PolicyResolution, *EventLog, error) {
	var semaphore = make(chan int, THREAD_POOL_SIZE)
	var errs = make(chan error, len(resolver.policy.Dependencies.DependenciesByID))

	// Run every declared dependency via policy and resolve it
	for _, d := range resolver.policy.Dependencies.DependenciesByID {
		// resolve dependency via applying policy
		semaphore <- 1
		go func(d *Dependency) {
			node, err := resolver.resolveDependency(d)
			errs <- resolver.combineData(node, err)
			<-semaphore
		}(d)
	}

	// Wait for all go routines to end
	errFound := 0
	for i := 0; i < len(resolver.policy.Dependencies.DependenciesByID); i++ {
		err := <-errs
		if err != nil {
			errFound++
		}
	}

	// See if there were any errors
	if errFound > 0 {
		return nil, resolver.eventLog, fmt.Errorf("Errors during resolving policy: %d", errFound)
	}

	// Once all components are resolved, print information about them into event log
	for _, instance := range resolver.resolution.ComponentInstanceMap {
		if instance.Metadata.Key.IsComponent() {
			resolver.logComponentCodeParams(instance)
			resolver.logComponentDiscoveryParams(instance)
		}
	}

	return resolver.resolution, resolver.eventLog, nil
}

// Resolves a single dependency
func (resolver *PolicyResolver) resolveDependency(d *language.Dependency) (*resolutionNode, error) {
	// create resolution node and resolve it
	node := resolver.newResolutionNode(d)
	return node, resolver.resolveNode(node)
}

// Combines resolution data into the overall state of the world
func (resolver *PolicyResolver) combineData(node *resolutionNode, resolutionErr error) error {
	resolver.combineMutex.Lock()

	// aggregate logs in the end
	defer func() {
		for _, eventLog := range node.eventLogsCombined {
			resolver.eventLog.Append(eventLog)
		}
		resolver.combineMutex.Unlock()
	}()

	// if there was a resolution error, return it
	if resolutionErr != nil {
		return resolutionErr
	}

	// exit if dependency has not been fulfilled. otherwise, proceed to data aggregation
	if !node.resolved {
		return nil
	}

	// add a record for dependency resolution
	resolver.resolution.DependencyInstanceMap[node.dependency.GetID()] = node.serviceKey.GetKey()

	// append component instance data
	err := resolver.resolution.AppendData(node.resolution)
	if err != nil {
		node.eventLog.LogError(err)
		return err
	}

	return nil
}

// Evaluate evaluates and resolves a single dependency ("<user> needs <service> with <labels>") and calculates component allocations
// Returns error only if there is an issue with the policy (e.g. it's malformed)
// Returns nil if there is no error (it may be that nothing was still matched though)
// If you want to check for successful resolution, use node.resolved flag
func (resolver *PolicyResolver) resolveNode(node *resolutionNode) error {
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

	// Locate the contract
	node.contract, err = node.getContract(resolver.policy)
	if err != nil {
		// Return a policy processing error in case service is not present in policy
		return node.cannotResolveInstance(err)
	}
	node.objectResolved(node.contract)

	// Process service and transform labels
	node.transformLabels(node.labels, node.contract.ChangeLabels)

	// Match the context
	node.context, err = node.getMatchedContext(resolver.policy)
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

	// Check that service, which current context is implemented with, exists
	node.service, err = node.getMatchedService(resolver.policy)
	if err != nil {
		// Return a policy processing error in case of context resolution failure
		return node.cannotResolveInstance(err)
	}
	node.objectResolved(node.service)

	// Process context and transform labels
	node.transformLabels(node.labels, node.context.ChangeLabels)

	// Resolve allocation keys for the context
	node.allocationKeysResolved, err = node.resolveAllocationKeys(resolver.policy)
	if err != nil {
		// Return an error in case of malformed policy or policy processing error
		return node.cannotResolveInstance(err)
	}

	// Create service key
	node.serviceKey = node.createComponentKey(nil)
	node.objectResolved(node.serviceKey)

	// Check if we've been there already
	cycle := ContainsString(node.path, node.serviceKey.GetKey())
	node.path = append(node.path, node.serviceKey.GetKey())
	if cycle {
		err = node.errorServiceCycleDetected()
		return node.cannotResolveInstance(err)
	}

	// Store labels for service
	node.resolution.RecordLabels(node.serviceKey, node.labels)

	// Store edge (last component instance -> service instance)
	node.resolution.StoreEdge(node.arrivalKey, node.serviceKey)

	// Process global rules before processing components
	ruleResult, err := node.processRules(resolver.policy)
	if err != nil {
		// Return an error in case of rule processing error
		return node.cannotResolveInstance(err)
	}

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
		node.resolution.StoreEdge(node.serviceKey, node.componentKey)

		// Calculate and store labels for component
		node.resolution.RecordLabels(node.componentKey, node.labels)

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
		} else if node.component.Contract != "" {
			// Create a child node for dependency resolution
			nodeNext := node.createChildNode()

			// Resolve dependency on another contract recursively
			err := resolver.resolveNode(nodeNext)

			// Combine event logs
			node.eventLogsCombined = append(node.eventLogsCombined, nodeNext.eventLogsCombined...)

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
		node.logInstanceSuccessfullyResolved(node.componentKey)
		node.resolution.RecordResolved(node.componentKey, node.dependency, ruleResult)
	}

	// Mark note as resolved and record usage of a given service instance
	node.resolved = true
	node.logInstanceSuccessfullyResolved(node.serviceKey)
	node.resolution.RecordResolved(node.serviceKey, node.dependency, ruleResult)

	return nil
}
