package resolve

import (
	. "github.com/Aptomi/aptomi/pkg/slinga/eventlog"
	. "github.com/Aptomi/aptomi/pkg/slinga/language"
	"github.com/Aptomi/aptomi/pkg/slinga/object"
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

	// reference to the namespace & contract we are currently resolving
	namespace    string
	contractName string
	contract     *Contract

	// reference to the current set of labels
	labels *LabelSet

	// reference to the context & the corresponding service that were matched
	context *Context
	service *Service

	// reference to the allocation keys that were resolved
	allocationKeysResolved []string

	// reference to the current node in discovery tree for components announcing their discovery properties
	// component1...component2...component3 -> component instance key
	discoveryTreeNode NestedParameterMap

	// reference to the current component in discovery tree
	component *ServiceComponent

	// reference to the current component key
	componentKey *ComponentInstanceKey

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
	labels := NewLabelSet(dependency.Labels)
	if user != nil {
		labels.AddLabels(user.Labels)
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

		// we start with the namespace & contract specified in the dependency
		namespace:    dependency.Namespace,
		contractName: dependency.Contract,

		// start with the generated set of labels
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

		// we take the current component we are iterating over, and get its contract name
		namespace:    node.namespace,
		contractName: node.component.Contract,

		// proceed with the current set of labels
		labels: NewLabelSet(node.labels.Labels),

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
// - contract
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

// Helper to get a contract
func (node *resolutionNode) getContract(policy *Policy) (*Contract, error) {
	contractObj, err := policy.GetObject(ContractObject.Kind, node.contractName, node.namespace)
	if err != nil {
		return nil, node.errorContractDoesNotExist()
	}
	if contractObj == nil {
		return nil, node.errorContractDoesNotExist()
	}
	contract := contractObj.(*Contract)
	node.logContractFound(contract)
	return contract, nil
}

// Helper to get a matched context
func (node *resolutionNode) getMatchedContext(policy *Policy) (*Context, error) {
	// Locate the list of contexts for service
	node.logStartMatchingContexts()

	// Find matching context
	contextualDataForExpression := node.getContextualDataForContextExpression()
	var contextMatched *Context
	for _, context := range node.contract.Contexts {
		// Check if context matches (based on criteria)
		matched, err := context.Matches(contextualDataForExpression, node.resolver.expressionCache)
		if err != nil {
			// Propagate error up
			return nil, node.errorWhenTestingContext(context, err)
		}
		node.logTestedContextCriteria(context, matched)
		if matched {
			contextMatched = context
			break
		}
	}

	if contextMatched != nil {
		node.logContextMatched(contextMatched)
	} else {
		node.logContextNotMatched()
	}

	return contextMatched, nil
}

// Helper to get a matched service
func (node *resolutionNode) getMatchedService(policy *Policy) (*Service, error) {
	if node.context.Allocation == nil {
		return nil, node.errorServiceDoesNotExist()
	}

	serviceObj, err := policy.GetObject(ServiceObject.Kind, node.context.Allocation.Service, node.namespace)
	if err != nil {
		return nil, node.errorServiceDoesNotExist()
	}
	if serviceObj == nil {
		return nil, node.errorServiceDoesNotExist()
	}

	service := serviceObj.(*Service)
	serviceOwner := node.resolver.externalData.UserLoader.LoadUserByID(service.Owner)

	// If a service has no owner, it is considered a malformed policy, so let's return an error
	if serviceOwner == nil {
		return nil, node.errorServiceOwnerDoesNotExist(service)
	}

	// Service should be located in the same namespace as contract
	if service.Namespace != node.contract.Namespace {
		return nil, node.errorServiceIsNotInSameNamespaceAsContract(service)
	}

	node.logServiceFound(service)
	return service, nil
}

// Helper to resolve allocation keys
func (node *resolutionNode) resolveAllocationKeys(policy *Policy) ([]string, error) {
	// If there is no allocation, there are no keys to resolve
	if node.context.Allocation == nil {
		return nil, nil
	}

	// Resolve allocation keys (they can be dynamic, depending on user labels)
	result, err := node.context.ResolveKeys(node.getContextualDataForContextAllocationTemplate(), node.resolver.templateCache)
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
func (node *resolutionNode) createComponentKey(component *ServiceComponent) (*ComponentInstanceKey, error) {
	clusterObj, err := node.resolver.policy.GetObject(ClusterObject.Kind, node.labels.Labels[LabelCluster], object.SystemNS)
	if err != nil {
		return nil, node.errorClusterDoesNotExist()
	}
	if clusterObj == nil {
		return nil, node.errorClusterDoesNotExist()
	}

	return NewComponentInstanceKey(
		clusterObj.(*Cluster),
		node.contract,
		node.context,
		node.allocationKeysResolved,
		node.service,
		component,
	), nil
}

func (node *resolutionNode) transformLabels(labels *LabelSet, operations LabelOperations) {
	changedLabels := labels.ApplyTransform(operations)
	if changedLabels {
		node.logLabels(labels, "after transform")
	}
}

func (node *resolutionNode) processRulesWithinNamespace(policyNamespace *PolicyNamespace, result *RuleActionResult) error {
	if policyNamespace == nil {
		return nil
	}

	rules := policyNamespace.Rules.GetRulesSortedByWeight()
	contextualDataForRule := node.getContextualDataForRuleExpression()
	for _, rule := range rules {
		matched, err := rule.Matches(contextualDataForRule, node.resolver.expressionCache)
		if err != nil {
			return node.errorWhenProcessingRule(rule, err)
		}
		node.logTestedRuleMatch(rule, matched)
		if matched {
			rule.ApplyActions(result)

			// if a dependency has been rejected, handle it right away and return that we cannot resolve it
			if len(rule.Actions.Dependency) > 0 {
				if !result.AllowDependency {
					return node.errorDependencyNotAllowedByRules()
				}
			}
			if result.ChangedLabelsOnLastApply {
				node.logLabels(result.Labels, "after transform")
			}
		}
	}

	if !result.AllowDependency {
		return node.errorDependencyNotAllowedByRules()
	}

	node.logRulesProcessingResult(policyNamespace, result)
	return nil
}

func (node *resolutionNode) processRules() (*RuleActionResult, error) {
	result := NewRuleActionResult(node.labels)
	var err error

	// process rules within the current namespace
	err = node.processRulesWithinNamespace(node.resolver.policy.Namespace[node.namespace], result)
	if err != nil {
		return nil, err
	}

	// process rules globally (within system namespace)
	err = node.processRulesWithinNamespace(node.resolver.policy.Namespace[object.SystemNS], result)
	if err != nil {
		return nil, err
	}
	return result, nil
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

	// Populate discovery tree (allow this component to announce its discovery properties in the discovery tree)
	node.discoveryTreeNode.GetNestedMap(node.component.Name)["instance"] = EscapeName(node.componentKey.GetKey())
	for k, v := range componentDiscoveryParams {
		node.discoveryTreeNode.GetNestedMap(node.component.Name)[k] = v
	}

	return nil
}
