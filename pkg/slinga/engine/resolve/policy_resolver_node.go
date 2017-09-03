package resolve

import (
	. "github.com/Aptomi/aptomi/pkg/slinga/eventlog"
	. "github.com/Aptomi/aptomi/pkg/slinga/language"
	. "github.com/Aptomi/aptomi/pkg/slinga/util"
)

// This is a special internal structure that gets used by the engine, while we traverse the policy graph for a given dependency
// It gets incrementally populated with data, as policy evaluation goes on for a given dependency
type resolutionNode struct {
	// whether we successfully resolved this node or not
	resolved bool

	// pointer to the policy resolver
	resolver *PolicyResolver

	// pointer to event log (local to the node)
	eventLog *EventLog

	// combined event logs from all resolution nodes in the subtree
	eventLogsCombined []*EventLog

	// new instance of PolicyResolution, where resolution resolution will be stored
	resolution *PolicyResolution

	// depth we are currently on (as we are traversing policy graph), with initial dependency being on depth 0
	depth int

	// reference to initial dependency
	dependency *Dependency

	// reference to user who requested this dependency
	user *User

	// reference to the service we are currently resolving
	serviceName string
	service     *Service

	// reference to the current set of labels
	labels LabelSet

	// reference to the context that was matched
	context *Context

	// reference to the allocation keys that were resolved
	allocationKeysResolved []string

	// reference to the current node in discovery tree for components announcing their discovery properties
	// component1...component2...component3 -> component instance key
	discoveryTreeNode NestedParameterMap

	// reference to the current component in discovery tree
	component *ServiceComponent

	// reference to the current component key
	componentKey *ComponentInstanceKey

	// reference to the current labels during component processing
	componentLabels LabelSet

	// reference to the current service key
	serviceKey *ComponentInstanceKey

	// reference to the last key we arrived with, so we can reconstruct graph edges between keys
	arrivalKey *ComponentInstanceKey

	// path that we traveled so far (to detect cycles)
	path []string
}

// Creates a new resolution node as a starting point for resolving a particular dependency
func (resolver *PolicyResolver) newResolutionNode(dependency *Dependency) *resolutionNode {

	// combining user labels and dependency labels
	user := resolver.externalData.UserLoader.LoadUserByID(dependency.UserID)
	labels := dependency.GetLabelSet()
	if user != nil {
		labels = labels.AddLabels(user.GetLabelSet())
	}

	eventLog := NewEventLog()
	return &resolutionNode{
		resolved: false,

		resolver:          resolver,
		eventLog:          eventLog,
		eventLogsCombined: []*EventLog{eventLog},

		resolution: NewPolicyResolution(),

		depth:      0,
		dependency: dependency,
		user:       user,

		// we start with the service specified in the dependency
		serviceName: dependency.Service,

		// combining user labels and dependency labels
		labels: labels,

		// empty discovery tree
		discoveryTreeNode: NestedParameterMap{},

		// empty path
		path: []string{},
	}
}

// Creates a new resolution node (as we are processing dependency on another service)
func (node *resolutionNode) createChildNode() *resolutionNode {
	eventLog := NewEventLog()
	return &resolutionNode{
		resolved: false,

		resolver:          node.resolver,
		eventLog:          eventLog,
		eventLogsCombined: []*EventLog{eventLog},

		resolution: node.resolution,

		depth:      node.depth + 1,
		dependency: node.dependency,
		user:       node.user,

		// we take the current component we are iterating over, and get its service name
		serviceName: node.component.Service,

		// we take current processed labels for the component
		labels: node.componentLabels,

		// move further by the discovery tree via component name link
		discoveryTreeNode: node.discoveryTreeNode.GetNestedMap(node.component.Name),

		// remember the last arrival key
		arrivalKey: node.componentKey,

		// copy path
		path: CopySliceOfStrings(node.path),
	}
}

// This method is called by the main engine resolution engine when an error happens
// If analyzes error type, writes the corresponding messages into the log
// And makes a decision whether to swallow the error, or fail policy processing
func (node *resolutionNode) cannotResolveInstance(err error) error {
	var criticalError *CriticalError
	var isCriticalError bool = false

	// Log critical error as error in the event log
	if err != nil {
		criticalError, isCriticalError = err.(*CriticalError)
		if isCriticalError {
			if !criticalError.IsLogged() {
				// Log it
				node.eventLog.LogError(err)

				// Mark this error as processed. So that when we go up the recursion stack, we don't log it multiple times
				criticalError.SetLoggedFlag()
			}
		} else {
			// Log it
			node.eventLog.LogErrorAsWarning(err)
		}
	}

	// Log that service or component instance cannot be resolved
	node.logCannotResolveInstance()

	// There may be a situation when service key has not been resolved yet. If so, we should create a fake one to attach event log to
	if node.serviceKey == nil {
		// Create service key
		node.serviceKey = node.createComponentKey(nil)

		// Once instance is figured out, make sure to attach event logs to that instance
		node.objectResolved(node.serviceKey)
	}

	// If it's a critical error, return it
	if isCriticalError {
		return err
	}

	// Otherwise, tell engine to swallow it
	return nil
}

// As the resolution goes on, this method is called when objects become resolved and available in the context
// Right now it gets called for as the following get resolved:
// - dependency
// - user
// - service
// - context
// - serviceKey
func (node *resolutionNode) objectResolved(object interface{}) {
	node.eventLog.AttachTo(object)
}

// Helper to check that user exists
func (node *resolutionNode) checkUserExists() error {
	if node.user == nil {
		return node.errorUserDoesNotExist()
	}
	return nil
}

// Helper to get a matched service
func (node *resolutionNode) getMatchedService(policy *PolicyNamespace) (*Service, error) {
	service := policy.Services[node.serviceName]
	if service == nil {
		// This is considered a malformed policy, so let's return an error
		return nil, node.errorServiceDoesNotExist()
	}

	serviceOwner := node.resolver.externalData.UserLoader.LoadUserByID(service.Owner)
	if serviceOwner == nil {
		// This is considered a malformed policy, so let's return an error
		return nil, node.errorServiceOwnerDoesNotExist()
	}

	node.logServiceFound(service)
	return service, nil
}

// Helper to get a matched context
func (node *resolutionNode) getMatchedContext(policy *PolicyNamespace) (*Context, error) {
	// Locate the list of contexts for service
	node.logStartMatchingContexts()

	// Find matching context
	contextualDataForExpression := node.getContextualDataForExpression()
	var contextMatched *Context
	for _, context := range policy.Contexts {

		// Check if context matches (based on criteria)
		matched, err := context.Matches(contextualDataForExpression, node.resolver.expressionCache)
		if err != nil {
			// Propagate error up
			return nil, node.errorWhenTestingContext(context, err)
		}
		node.logTestedContextCriteria(context, matched)

		// If criteria matches, then check global rules as well
		if matched {
			// Match is only valid if there are no global rule violations for the current context
			labels := node.transformLabels(node.labels, context.ChangeLabels)

			// Lookup cluster from a label
			cluster, err := policy.GetClusterByLabels(labels)
			if err != nil {
				// Propagate error up
				return nil, node.errorGettingClusterForGlobalRules(context, labels, err)
			}

			// Check for rule violations
			hasViolations, err := node.hasGlobalRuleViolations(policy, context, labels, cluster)
			if err != nil {
				// Propagate error up. Don't wrap this error into anything else. It's already good
				return nil, err
			}
			node.logTestedGlobalRuleViolations(context, labels, !hasViolations)

			if !hasViolations {
				contextMatched = context
				break
			}
		}
	}

	if contextMatched != nil {
		node.logContextMatched(contextMatched)
	} else {
		node.logContextNotMatched()
	}

	return contextMatched, nil
}

// Helper to resolve allocation keys
func (node *resolutionNode) resolveAllocationKeys(policy *PolicyNamespace) ([]string, error) {
	// If there is no allocation, there are no keys to resolve
	if node.context.Allocation == nil {
		return nil, nil
	}

	// Resolve allocation keys (they can be dynamic, depending on user labels)
	result, err := node.context.ResolveKeys(node.getContextualDataForAllocationTemplate(), node.resolver.templateCache)
	if err != nil {
		return nil, node.errorWhenResolvingAllocationKeys(err)
	}

	node.logAllocationKeysSuccessfullyResolved(result)
	return result, nil
}

func (node *resolutionNode) sortServiceComponents() ([]*ServiceComponent, error) {
	result, err := node.service.GetComponentsSortedTopologically()
	if err != nil {
		return nil, node.errorWhenDoingTopologicalSort(err)
	}
	return result, nil
}

// createComponentKey creates a component key
func (node *resolutionNode) createComponentKey(component *ServiceComponent) *ComponentInstanceKey {
	return NewComponentInstanceKey(
		node.serviceName,
		node.context,
		node.allocationKeysResolved,
		component,
	)
}

func (node *resolutionNode) transformLabels(labels LabelSet, operations LabelOperations) LabelSet {
	result := labels.ApplyTransform(operations)
	if !labels.Equal(result) {
		node.logLabels(result, "after transform")
	}
	return result
}

func (node *resolutionNode) hasGlobalRuleViolations(policy *PolicyNamespace, context *Context, labels LabelSet, cluster *Cluster) (bool, error) {
	globalRules := policy.Rules
	if rules, ok := globalRules.Rules["dependency"]; ok {
		for _, rule := range rules {
			matched, err := rule.FilterServices.Match(labels, node.user, cluster, node.resolver.expressionCache)
			if err != nil {
				return true, node.errorWhenTestingGlobalRule(context, rule, labels, err)
			}

			node.logTestedGlobalRuleMatch(context, rule, labels, matched)

			if matched {
				for _, action := range rule.Actions {
					if action.Type == "dependency" && action.Content == "forbid" {
						return true, nil
					}
				}
			}
		}
	}

	return false, nil
}

func (node *resolutionNode) calculateAndStoreCodeParams() error {
	componentCodeParams, err := evaluateParameterTree(node.component.Code.Params, node.getContextualDataForCodeDiscoveryTemplate(), node.resolver.templateCache)
	if err != nil {
		return node.errorWhenProcessingCodeParams(err)
	}

	err = node.resolution.RecordCodeParams(node.componentKey, componentCodeParams)
	if err != nil {
		return node.errorWhenProcessingCodeParams(err)
	}

	node.logComponentCodeParams()

	return nil
}

func (node *resolutionNode) calculateAndStoreDiscoveryParams() error {
	componentDiscoveryParams, err := evaluateParameterTree(node.component.Discovery, node.getContextualDataForCodeDiscoveryTemplate(), node.resolver.templateCache)
	if err != nil {
		return node.errorWhenProcessingDiscoveryParams(err)
	}

	err = node.resolution.RecordDiscoveryParams(node.componentKey, componentDiscoveryParams)
	if err != nil {
		return node.errorWhenProcessingDiscoveryParams(err)
	}

	node.logComponentDiscoveryParams()

	// Populate discovery tree (allow this component to announce its discovery properties in the discovery tree)
	node.discoveryTreeNode.GetNestedMap(node.component.Name)["instance"] = EscapeName(node.componentKey.GetKey())
	for k, v := range componentDiscoveryParams {
		node.discoveryTreeNode.GetNestedMap(node.component.Name)[k] = v
	}

	return nil
}
