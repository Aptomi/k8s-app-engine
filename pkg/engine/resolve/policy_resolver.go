package resolve

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/event"
	"github.com/Aptomi/aptomi/pkg/external"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/lang/expression"
	"github.com/Aptomi/aptomi/pkg/lang/template"
	"github.com/Aptomi/aptomi/pkg/runtime"
	"github.com/Aptomi/aptomi/pkg/util"
	sysruntime "runtime"
	"sync"
)

// MaxConcurrentGoRoutines is the number of concurrently running goroutines for policy evaluation and processing.
// We don't necessarily want to run a lot of them due to CPU/memory constraints and due to the fact that there is minimal
// io wait time in policy processing (goroutines are mostly busy doing calculations as opposed to waiting).
var MaxConcurrentGoRoutines = sysruntime.NumCPU()

// PolicyResolver is a core of Aptomi for policy resolution and translating all service consumption declarations
// into a single PolicyResolution object which represents desired state of components running in a cloud.
type PolicyResolver struct {
	/*
		Input objects
	*/

	// Policy
	policy *lang.Policy

	// External data
	externalData *external.Data

	/*
		Cache
	*/

	// Expression cache
	expressionCache *expression.Cache

	// Template cache
	templateCache *template.Cache

	/*
		Calculated objects (aggregated over all dependencies)
	*/

	combineMutex sync.Mutex

	// Reference to the calculated PolicyResolution
	resolution *PolicyResolution

	// Buffered event log - gets populated during policy resolution
	eventLog *event.Log
}

// NewPolicyResolver creates a new policy resolver
func NewPolicyResolver(policy *lang.Policy, externalData *external.Data, eventLog *event.Log) *PolicyResolver {
	return &PolicyResolver{
		policy:          policy,
		externalData:    externalData,
		expressionCache: expression.NewCache(),
		templateCache:   template.NewCache(),
		resolution:      NewPolicyResolution(),
		eventLog:        eventLog,
	}
}

// ResolveAllDependencies takes policy as input and calculates PolicyResolution (desired state) as output.
//
// It resolves all recorded service consumption declarations ("<user> needs <contract> with <labels>"), calculating
// which component have to be allocated and with which parameters. Once PolicyResolution (desired state) is calculated,
// it can be rendered by the engine diff/apply by deploying/configuring required components/containers in the cloud.
func (resolver *PolicyResolver) ResolveAllDependencies() (*PolicyResolution, error) {
	// Run policy validation before resolution, just in case
	err := resolver.policy.Validate()
	if err != nil {
		return nil, err
	}

	// Allocate semaphore
	var semaphore = make(chan int, MaxConcurrentGoRoutines)
	dependencies := resolver.policy.GetObjectsByKind(lang.DependencyObject.Kind)
	var errs = make(chan error, len(dependencies))

	// Run every declared dependency via policy and resolve it
	for _, d := range dependencies {
		// resolve dependency via applying policy
		semaphore <- 1
		go func(d *lang.Dependency) {
			node, resolveErr := resolver.resolveDependency(d)
			errs <- resolver.combineData(node, resolveErr)
			<-semaphore
		}(d.(*lang.Dependency))
	}

	// Wait for all go routines to end
	errFound := 0
	for i := 0; i < len(dependencies); i++ {
		resolveErr := <-errs
		if resolveErr != nil {
			errFound++
		}
	}

	// See if there were any errors
	if errFound > 0 {
		return nil, fmt.Errorf("errors occurred during policy resolution: %d", errFound)
	}

	// Once all components are resolved, print information about them into event log
	for _, instance := range resolver.resolution.ComponentInstanceMap {
		if instance.Metadata.Key.IsComponent() {
			resolver.logComponentCodeParams(instance)
			resolver.logComponentDiscoveryParams(instance)
		}
	}

	return resolver.resolution, nil
}

// Resolves a single dependency
func (resolver *PolicyResolver) resolveDependency(d *lang.Dependency) (node *resolutionNode, resolveErr error) {
	// create new resolution node
	node = resolver.newResolutionNode()

	// make sure we are converting panics into errors
	defer func() {
		if err := recover(); err != nil {
			resolveErr = fmt.Errorf("panic: %s", err)
			node.eventLog.LogError(resolveErr)
		}
	}()

	// populate resolution node with data (e.g. put initial set of labels from user & dependency)
	resolver.initResolutionNode(node, d)

	// resolve it
	resolveErr = resolver.resolveNode(node)
	return node, resolveErr
}

// Combines resolution data into the overall state of the world
func (resolver *PolicyResolver) combineData(node *resolutionNode, resolutionErr error) error {
	// put a lock
	resolver.combineMutex.Lock()

	// aggregate logs in the end, especially if resolutionErr occurred
	defer func() {
		if node != nil {
			for _, eventLog := range node.eventLogsCombined {
				resolver.eventLog.Append(eventLog)
			}
		}
		resolver.combineMutex.Unlock()
	}()

	// if there was a resolution error, return it
	if resolutionErr != nil {
		return resolutionErr
	}

	// if node is nil (likely, panic happened), return
	if node == nil {
		return resolutionErr
	}

	// exit if dependency has not been fulfilled. otherwise, proceed to data aggregation
	if !node.resolved || node.serviceKey == nil {
		return nil
	}

	// add a record for dependency resolution
	resolver.resolution.DependencyInstanceMap[runtime.KeyForStorable(node.dependency)] = node.serviceKey.GetKey()

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

	// Locate the contract (it should be always be present, as policy has been validated)
	node.contract = node.getContract(resolver.policy)
	node.namespace = node.contract.Namespace
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

	// Process global rules before processing service key and dependent component keys
	ruleResult, err := node.processRules()
	if err != nil {
		// Return an error in case of rule processing error
		return node.cannotResolveInstance(err)
	}
	// Create service key
	node.serviceKey, err = node.createComponentKey(nil)
	if err != nil {
		// Return an error in case of malformed policy or policy processing error
		return node.cannotResolveInstance(err)
	}
	node.objectResolved(node.serviceKey)

	// Check if we've been there already
	cycle := util.ContainsString(node.path, node.serviceKey.GetKey())
	node.path = append(node.path, node.serviceKey.GetKey())
	if cycle {
		err = node.errorServiceCycleDetected()
		return node.cannotResolveInstance(err)
	}

	// Store labels for service
	node.resolution.RecordLabels(node.serviceKey, node.labels)

	// Store edge (last component instance -> service instance)
	node.resolution.StoreEdge(node.arrivalKey, node.serviceKey)

	// Now, sort all components in topological order (it should always succeed, as policy has been validated)
	componentsOrdered, err := node.service.GetComponentsSortedTopologically()
	if err != nil {
		// Return an error in case of failed component topological sort
		return node.cannotResolveInstance(err)
	}

	// Iterate over all service components and resolve them recursively
	// Note that discovery variables can refer to other variables announced by dependents in the discovery tree
	for _, node.component = range componentsOrdered {
		// Create key
		node.componentKey, err = node.createComponentKey(node.component)
		if err != nil {
			// Return an error in case of malformed policy or policy processing error
			return node.cannotResolveInstance(err)
		}

		// Store edge (service instance -> component instance)
		node.resolution.StoreEdge(node.serviceKey, node.componentKey)

		// Calculate and store labels for component
		node.resolution.RecordLabels(node.componentKey, node.labels)

		// Create new map with resolution keys for component
		node.discoveryTreeNode[node.component.Name] = util.NestedParameterMap{}

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
