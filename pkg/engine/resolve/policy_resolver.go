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

// PolicyResolver is a core of Aptomi for policy resolution and translating all claims
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
		Calculated objects (aggregated over all claims)
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

// ResolveAllClaims takes policy as input and calculates PolicyResolution (desired state) as output.
//
// The method resolves all recorded claims for consuming services ("instantiate <service> with <labels>"), calculating
// which components have to be allocated and with which parameters. Once PolicyResolution (desired state) is calculated,
// it can be rendered by the engine diff/apply by deploying and configuring required components in the cloud.
//
// As a result, status of every claim will be stored in resolution state.
func (resolver *PolicyResolver) ResolveAllClaims() *PolicyResolution {
	// Allocate semaphore, making sure we don't run more than MaxConcurrentGoRoutines go routines at the same time
	var semaphore = make(chan int, MaxConcurrentGoRoutines)
	var wg sync.WaitGroup
	claims := resolver.policy.GetObjectsByKind(lang.ClaimObject.Kind)

	// Resolve every declared claim
	for _, claim := range claims {
		// Start go routine for resolving a given claim
		wg.Add(1)
		semaphore <- 1
		go func(c *lang.Claim) {
			defer wg.Done()
			node, resolveErr := resolver.resolveClaim(c)
			resolver.combineData(node, resolveErr)
			<-semaphore
		}(claim.(*lang.Claim))
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

// Resolves a single claim and returns an error if it cannot be resolved
func (resolver *PolicyResolver) resolveClaim(claim *lang.Claim) (node *resolutionNode, resolveErr error) {
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
	resolver.initResolutionNode(node, claim)

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

// Evaluate evaluates and resolves a single claim, as well as calculates component allocations.
// Returns error only if there is an issue with the given claim and it cannot be resolved
func (resolver *PolicyResolver) resolveNode(node *resolutionNode) (resolveErr error) { // nolint: gocyclo
	recursiveError := false

	// if this function returns an error, it needs to be logged
	defer func() {
		if resolveErr != nil {
			// Log resolution error before we exit
			if !recursiveError {
				node.eventLog.NewEntry().Error(printCauseDetailsOnDebug(resolveErr, node.eventLog))
			}

			// Log that bundle or component instance cannot be resolved
			node.logCannotResolveInstance()
		}
	}()

	// Error variable that we will be reusing
	var err error

	// Indicate that we are starting to resolve claim
	node.objectResolved(node.claim)
	node.logStartResolvingClaim()

	// Locate the user
	err = node.checkUserExists()
	if err != nil {
		return err
	}

	// Locate the service (it should be always be present, as policy has been validated. but user
	// may or may not have permissions to consume it)
	node.service, err = node.getService(resolver.policy)
	if err != nil {
		return err
	}

	node.namespace = node.service.Namespace
	node.objectResolved(node.service)

	// Process bundle and transform labels
	node.transformLabels(node.labels, node.service.ChangeLabels)

	// Match the context
	node.context, err = node.getMatchedContext(resolver.policy)
	if err != nil {
		return err
	}

	// Check that bundle, which current context is implemented with, exists
	node.bundle, err = node.getMatchedBundle(resolver.policy)
	if err != nil {
		return err
	}
	node.objectResolved(node.bundle)

	// Process context and transform labels
	node.transformLabels(node.labels, node.context.ChangeLabels)

	// Resolve allocation keys for the context
	node.allocationKeysResolved, err = node.resolveAllocationKeys(resolver.policy)
	if err != nil {
		return err
	}

	// Process global rules before processing bundle key and dependent component keys
	ruleResult, err := node.processRules()
	if err != nil {
		return err
	}
	// Create bundle key
	node.bundleKey, err = node.createComponentKey(nil)
	if err != nil {
		return err
	}

	// Check if we've been there already and therefore hit a bundle cycle
	cycle := util.ContainsString(node.path, node.bundleKey.GetKey())
	node.path = append(node.path, node.bundleKey.GetKey())
	if cycle {
		return node.errorBundleCycleDetected()
	}

	// Store labels for bundle
	node.resolution.RecordLabels(node.bundleKey, node.labels)

	// Store edge (last component instance -> bundle instance)
	node.resolution.StoreEdge(node.arrivalKey, node.bundleKey)

	// Now, sort all components in topological order (it should always succeed, as policy has been validated)
	componentsOrdered, err := node.bundle.GetComponentsSortedTopologically()
	if err != nil {
		return err
	}

	// Iterate over all bundle components and resolve them recursively
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

		// Store edge (bundle instance -> component instance)
		node.resolution.StoreEdge(node.bundleKey, node.componentKey)

		// Calculate and store labels for component
		node.resolution.RecordLabels(node.componentKey, node.labels)

		// Create new map with resolution keys for component
		node.discoveryTreeNode[node.component.Name] = util.NestedParameterMap{}

		// Calculate and store discovery params
		err := node.calculateAndStoreDiscoveryParams()
		if err != nil {
			return err
		}

		// Print information that we are starting to resolve claim (on code, or on bundle)
		node.logResolvingClaimOnComponent()

		if node.component.Code != nil {
			// Evaluate code params
			err := node.calculateAndStoreCodeParams()
			if err != nil {
				return err
			}
		} else if node.component.Service != "" {
			// Create a child node for claim resolution
			nodeNext := node.createChildNode()

			// Resolve claim on another service recursively
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
		node.resolution.RecordResolved(node.componentKey, node.claim, node.depth, ruleResult)
	}

	// Mark note as resolved and record usage of a given bundle instance
	node.logInstanceSuccessfullyResolved(node.bundleKey)
	node.resolution.RecordResolved(node.bundleKey, node.claim, node.depth, ruleResult)

	return nil
}
