package resolve

import (
	"fmt"
	sysruntime "runtime"
	"runtime/debug"
	"sync"

	"github.com/Aptomi/aptomi/pkg/event"
	"github.com/Aptomi/aptomi/pkg/external"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/lang/expression"
	"github.com/Aptomi/aptomi/pkg/lang/template"
	"github.com/Aptomi/aptomi/pkg/util"
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

// NewPolicyResolver creates a new policy resolver. You must call policy.Validate() before calling this method, to
// ensure that the policy is valid.
func NewPolicyResolver(policy *lang.Policy, externalData *external.Data, eventLog *event.Log) *PolicyResolver {
	// Check that the policy is valid
	policyValidationErr := policy.Validate()
	if policyValidationErr != nil {
		panic(fmt.Sprintf("can't create resolver because policy is invalid: %s", policyValidationErr))
	}

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
// The method resolves all recorded claims for consuming contracts ("instantiate <contract> with <labels>"), calculating
// which components have to be allocated and with which parameters. Once PolicyResolution (desired state) is calculated,
// it can be rendered by the engine diff/apply by deploying and configuring required components in the cloud.
//
// As a result, status of every dependency will be stored in resolution state.
func (resolver *PolicyResolver) ResolveAllDependencies() *PolicyResolution {
	// Allocate semaphore, making sure we don't run more than MaxConcurrentGoRoutines go routines at the same time
	var semaphore = make(chan int, MaxConcurrentGoRoutines)
	var wg sync.WaitGroup
	dependencies := resolver.policy.GetObjectsByKind(lang.DependencyObject.Kind)

	// Resolve every declared dependency
	for _, d := range dependencies {
		// Start go routine for resolving a given dependency
		wg.Add(1)
		semaphore <- 1
		go func(d *lang.Dependency) {
			defer wg.Done()
			node, resolveErr := resolver.resolveDependency(d)
			resolver.combineData(node, resolveErr)
			<-semaphore
		}(d.(*lang.Dependency))
	}

	// Wait for all go routines to end
	wg.Wait()

	// Once all components are resolved, print information about them into event log
	for _, instance := range resolver.resolution.ComponentInstanceMap {
		if instance.Metadata.Key.IsComponent() {
			resolver.logComponentParams(instance)
		}
	}

	return resolver.resolution
}

// Resolves a single dependency and returns an error if it cannot be resolved
func (resolver *PolicyResolver) resolveDependency(d *lang.Dependency) (node *resolutionNode, resolveErr error) {
	// make sure we are converting panics into errors
	defer func() {
		if err := recover(); err != nil {
			resolveErr = fmt.Errorf("panic: %s\n%s", err, string(debug.Stack()))
			node.eventLog.NewEntry().Error(resolveErr)
		}
	}()

	// create new resolution node
	node = resolver.newResolutionNode()

	// populate resolution node with data (e.g. construct initial set of labels)
	resolver.initResolutionNode(node, d)

	// resolve it
	resolveErr = resolver.resolveNode(node)
	return node, resolveErr
}

// Combines resolution data into the overall state of the world. If there is a conflict, it will return an error
func (resolver *PolicyResolver) combineData(node *resolutionNode, resolutionErr error) {
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

	// if there was no resolution error, combine component data
	if resolutionErr == nil {
		// aggregate component instance data
		resolver.resolution.AppendData(node.resolution)
	}
}

// Evaluate evaluates and resolves a single dependency ("<user> needs <service> with <labels>") and calculates component allocations
// Returns error only if there is an issue with the given dependency and it cannot be resolved
func (resolver *PolicyResolver) resolveNode(node *resolutionNode) (resolveErr error) {
	recursiveError := false

	// if this function returns an error, it needs to be logged
	defer func() {
		if resolveErr != nil {
			// Log resolution error before we exit
			if !recursiveError {
				node.eventLog.NewEntry().Error(printCauseDetailsOnDebug(resolveErr, node.eventLog))
			}

			// Log that service or component instance cannot be resolved
			node.logCannotResolveInstance()
		}
	}()

	// Error variable that we will be reusing
	var err error

	// Indicate that we are starting to resolve dependency
	node.objectResolved(node.dependency)
	node.logStartResolvingDependency()

	// Locate the user
	err = node.checkUserExists()
	if err != nil {
		return err
	}

	// Locate the contract (it should be always be present, as policy has been validated)
	node.contract = node.getContract(resolver.policy)
	node.namespace = node.contract.Namespace
	node.objectResolved(node.contract)

	// Process service and transform labels
	node.transformLabels(node.labels, node.contract.ChangeLabels)

	// Match the context
	node.context, err = node.getMatchedContext(resolver.policy)
	if err != nil {
		return err
	}

	// Check that service, which current context is implemented with, exists
	node.service, err = node.getMatchedService(resolver.policy)
	if err != nil {
		return err
	}
	node.objectResolved(node.service)

	// Process context and transform labels
	node.transformLabels(node.labels, node.context.ChangeLabels)

	// Resolve allocation keys for the context
	node.allocationKeysResolved, err = node.resolveAllocationKeys(resolver.policy)
	if err != nil {
		return err
	}

	// Process global rules before processing service key and dependent component keys
	ruleResult, err := node.processRules()
	if err != nil {
		return err
	}
	// Create service key
	node.serviceKey, err = node.createComponentKey(nil)
	if err != nil {
		return err
	}

	// Check if we've been there already and therefore hit a service cycle
	cycle := util.ContainsString(node.path, node.serviceKey.GetKey())
	node.path = append(node.path, node.serviceKey.GetKey())
	if cycle {
		return node.errorServiceCycleDetected()
	}

	// Store labels for service
	node.resolution.RecordLabels(node.serviceKey, node.labels)

	// Store edge (last component instance -> service instance)
	node.resolution.StoreEdge(node.arrivalKey, node.serviceKey)

	// Now, sort all components in topological order (it should always succeed, as policy has been validated)
	componentsOrdered, err := node.service.GetComponentsSortedTopologically()
	if err != nil {
		return err
	}

	// Iterate over all service components and resolve them recursively
	// Note that discovery variables can refer to other variables announced by dependents in the discovery tree
	for _, node.component = range componentsOrdered {
		// Check if component criteria holds
		componentMatch, componentMatchErr := node.componentMatches(node.component)
		if componentMatchErr != nil {
			return err
		}

		// If component criteria doesn't hold, do not proceed further
		if !componentMatch {
			continue
		}

		// Create component key and check that we were able to form it
		node.componentKey, err = node.createComponentKey(node.component)
		if err != nil {
			return err
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
			return err
		}

		// Print information that we are starting to resolve dependency (on code, or on service)
		node.logResolvingDependencyOnComponent()

		if node.component.Code != nil {
			// Evaluate code params
			err := node.calculateAndStoreCodeParams()
			if err != nil {
				return err
			}
		} else if node.component.Contract != "" {
			// Create a child node for dependency resolution
			nodeNext := node.createChildNode()

			// Resolve dependency on another contract recursively
			err := resolver.resolveNode(nodeNext)

			// Combine event logs first
			node.eventLogsCombined = append(node.eventLogsCombined, nodeNext.eventLogsCombined...)

			// Then return an error, if there was one
			if err != nil {
				recursiveError = true
				return err
			}
		}

		// Record usage of a given component instance
		node.logInstanceSuccessfullyResolved(node.componentKey)
		node.resolution.RecordResolved(node.componentKey, node.dependency, node.depth, ruleResult)
	}

	// Mark note as resolved and record usage of a given service instance
	node.logInstanceSuccessfullyResolved(node.serviceKey)
	node.resolution.RecordResolved(node.serviceKey, node.dependency, node.depth, ruleResult)

	return nil
}
