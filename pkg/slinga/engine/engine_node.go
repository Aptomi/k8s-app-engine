package engine

import (
	. "github.com/Aptomi/aptomi/pkg/slinga/language"
	. "github.com/Aptomi/aptomi/pkg/slinga/util"
	. "github.com/Aptomi/aptomi/pkg/slinga/eventlog"
	"fmt"
)

// This is a special internal structure that gets used by the engine, while we traverse the policy graph for a given dependency
// It gets incrementally populated with data, as policy evaluation goes on for a given dependency
type resolutionNode struct {
	// whether we successfully resolved this node or not
	resolved bool

	// reference to a cache
	cache *EngineCache

	// pointer to ServiceUsageState
	state *ServiceUsageState

	// pointer to event log
	eventLog *EventLog

	// new instance of ServiceUsageData, where resolution data will be stored
	data *ServiceUsageData

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

	// reference to rule log writer
	ruleLogWriter *RuleLogWriter
}

// Creates a new resolution node as a starting point for resolving a particular dependency
func (state *ServiceUsageState) newResolutionNode(dependency *Dependency, cache *EngineCache, eventLog *EventLog) *resolutionNode {

	// combining user labels and dependency labels
	user := state.userLoader.LoadUserByID(dependency.UserID)
	labels := dependency.GetLabelSet()
	if user != nil {
		labels = labels.AddLabels(user.GetLabelSet())
		labels = labels.AddLabels(user.GetSecretSet())
	}

	data := newServiceUsageData()
	return &resolutionNode{
		resolved: false,

		state:    state,
		cache:    cache,
		eventLog: eventLog,

		data:          data,
		ruleLogWriter: NewRuleLogWriter(data, dependency),

		depth:      0,
		dependency: dependency,
		user:       user,

		// we start with the service specified in the dependency
		serviceName: dependency.Service,

		// combining user labels and dependency labels
		labels: labels,

		// empty discovery tree
		discoveryTreeNode: NestedParameterMap{},
	}
}

// Creates a new resolution node (as we are processing dependency on another service)
func (node *resolutionNode) createChildNode() *resolutionNode {
	return &resolutionNode{
		resolved: false,

		state:    node.state,
		cache:    node.cache,
		eventLog: node.eventLog, // TODO: create new instance

		data:          node.data,
		ruleLogWriter: NewRuleLogWriter(node.data, node.dependency),

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
	}
}

// Helper to check that user exists
func (node *resolutionNode) checkUserExists() error {
	if node.user == nil {
		return fmt.Errorf("Dependency refers to non-existing user: " + node.dependency.UserID)
	}
	return nil
}

// Helper to get a matched service
func (node *resolutionNode) getMatchedService(policy *PolicyNamespace) (*Service, error) {
	service := policy.Services[node.serviceName]
	if service == nil {
		// This is considered a malformed policy, so let's return an error
		node.logServiceNotFoundError(node.serviceName)
		return nil, fmt.Errorf("Service not found: %s", node.serviceName)
	}
	node.logServiceFound(service)
	return service, nil
}

// Helper to get a matched context
func (node *resolutionNode) getMatchedContext(policy *PolicyNamespace) (*Context, error) {
	// Locate the list of contexts for service
	node.logStartMatchingContexts()

	// Find matching context
	var contextMatched *Context
	for _, context := range policy.Contexts {

		// Check if context matches (based on criteria)
		matched, err := context.Matches(node.getContextualDataForExpression(), node.cache.expressionCache)
		if err != nil {
			// Propagate error up
			node.logTestedContextCriteriaError(context, err)
			return nil, err
		}
		node.logTestedContextCriteria(context, matched)

		// If criteria matches, then check global rules as well
		if matched {
			// Match is only valid if there are no global rule voilations for the current context
			labels := node.transformLabels(node.labels, context.ChangeLabels)

			// Lookup cluster from a label
			cluster, err := getCluster(policy, labels)
			if err != nil {
				// Propagate error up
				node.logTestedGlobalRuleViolationsError(context, labels, err)
				return nil, err
			}

			// Check for rule voilations
			hasViolations, err := node.hasGlobalRuleViolations(policy, context, labels, cluster)
			if err != nil {
				// Propagate error up
				node.logTestedGlobalRuleViolationsError(context, labels, err)
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
	result, err := node.context.ResolveKeys(node.getContextualDataForAllocationTemplate(), node.cache.templateCache)
	if err != nil {
		node.logResolvingAllocationKeysError(err)
		return nil, err
	}

	if len(result) > 0 {
		node.logAllocationKeysSuccessfullyResolved(result)
	}
	return result, nil
}

func (node *resolutionNode) sortServiceComponents() ([]*ServiceComponent, error) {
	result, err := node.service.GetComponentsSortedTopologically()
	if err != nil {
		node.logFailedTopologicalSort(err)
		return result, err
	}
	return result, nil
}

// createComponentKey creates a component key
func (node *resolutionNode) createComponentKey(component *ServiceComponent) *ComponentInstanceKey {
	// TODO: write about creating component instance key
	return NewComponentInstanceKey(
		node.serviceName,
		node.context,
		node.allocationKeysResolved,
		component,
	)
}

func (node *resolutionNode) transformLabels(labels LabelSet, operations LabelOperations) LabelSet {
	// TODO: write about applying transform
	result := labels.ApplyTransform(operations)
	if !labels.Equal(result) {
		node.logLabels(result)
	}
	return result
}

func (node *resolutionNode) hasGlobalRuleViolations(policy *PolicyNamespace, context *Context, labels LabelSet, cluster *Cluster) (bool, error) {
	globalRules := policy.Rules
	if rules, ok := globalRules.Rules["dependency"]; ok {
		for _, rule := range rules {
			matched, err := rule.FilterServices.Match(labels, node.user, cluster, node.cache.expressionCache)
			if err != nil {
				node.logTestedGlobalRuleMatchError(context, rule, labels, err)
				return true, err
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
	// TODO: write into event log

	componentCodeParams, err := evaluateParameterTree(node.component.Code.Params, node.getContextualDataForCodeDiscoveryTemplate(), node.cache.templateCache)
	node.data.recordCodeParams(node.componentKey, componentCodeParams)
	return err
}

func (node *resolutionNode) calculateAndStoreDiscoveryParams() error {
	// TODO: write into event log

	componentDiscoveryParams, err := evaluateParameterTree(node.component.Discovery, node.getContextualDataForCodeDiscoveryTemplate(), node.cache.templateCache)
	node.data.recordDiscoveryParams(node.componentKey, componentDiscoveryParams)

	// Populate discovery tree (allow this component to announce its discovery properties in the discovery tree)
	node.discoveryTreeNode.GetNestedMap(node.component.Name)["instance"] = EscapeName(node.componentKey.GetKey())
	for k, v := range componentDiscoveryParams {
		node.discoveryTreeNode.GetNestedMap(node.component.Name)[k] = v
	}

	return err
}

// Mark instance as resolved and save dependency that triggered its instantiation
func (node *resolutionNode) recordResolved(cik *ComponentInstanceKey, dependency *Dependency) {
	node.logInstanceSuccessfullyResolved(cik)
	node.data.recordResolved(cik, dependency)
}
