package engine

import (
	. "github.com/Aptomi/aptomi/pkg/slinga/language"
	. "github.com/Aptomi/aptomi/pkg/slinga/log"
	. "github.com/Aptomi/aptomi/pkg/slinga/util"
	log "github.com/Sirupsen/logrus"
	"errors"
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
func (state *ServiceUsageState) newResolutionNode(dependency *Dependency, cache *EngineCache) *resolutionNode {

	user := state.userLoader.LoadUserByID(dependency.UserID)
	if user == nil {
		// Resolving allocations for service
		Debug.WithFields(log.Fields{
			"dependency": dependency,
		}).Panic("Dependency refers to non-existing user")
	}

	data := newServiceUsageData()
	return &resolutionNode{
		resolved: false,

		state: state,
		cache: cache,

		data:          data,
		ruleLogWriter: NewRuleLogWriter(data, dependency),

		depth:      0,
		dependency: dependency,
		user:       user,

		// we start with the service specified in the dependency
		serviceName: dependency.Service,

		// combining user labels, dependency labels, and making service.Name available to the engine as special variable
		labels: user.GetLabelSet().
			AddLabels(user.GetSecretSet()).
			AddLabels(dependency.GetLabelSet()).
			ApplyTransform(NewLabelOperationsSetSingleLabel("service.Name", dependency.Service)),

		// empty discovery tree
		discoveryTreeNode: NestedParameterMap{},
	}
}

// Creates a new resolution node (as we are processing dependency on another service)
func (node *resolutionNode) createChildNode() *resolutionNode {
	return &resolutionNode{
		resolved: false,

		state: node.state,
		cache: node.cache,

		data:          node.data,
		ruleLogWriter: NewRuleLogWriter(node.data, node.dependency),

		depth:      node.depth + 1,
		dependency: node.dependency,
		user:       node.user,

		// we take the current component we are iterating over, and get its service name
		serviceName: node.component.Service,

		// we take current processed labels for the component we are iterating over, and making service.Name available to the engine as special variable
		labels: node.componentLabels.
			ApplyTransform(NewLabelOperationsSetSingleLabel("service.Name", node.component.Service)),

		// move further by the discovery tree via component name link
		discoveryTreeNode: node.discoveryTreeNode.GetNestedMap(node.component.Name),

		// remember the last arrival key
		arrivalKey: node.componentKey,
	}
}

func (node *resolutionNode) debugResolvingDependencyStart() {
	Debug.WithFields(log.Fields{
		"dependency": node.dependency,
		"user":       node.user.Name,
		"labels":     node.labels,
		"dependsOn":  node.serviceName,
	}).Info("Resolving dependency")

	node.ruleLogWriter.addRuleLogEntry(entryResolvingDependencyStart(node.serviceName, node.user, node.dependency))
	node.ruleLogWriter.addRuleLogEntry(entryLabels(node.labels))
}

func (node *resolutionNode) debugResolvingDependencyEnd() {
	Debug.WithFields(log.Fields{
		"dependency": node.dependency,
		"user":       node.user.Name,
		"labels":     node.labels,
		"dependsOn":  node.serviceName,
	}).Info("Successfully resolved dependency")

	node.ruleLogWriter.addRuleLogEntry(entryResolvingDependencyEnd(node.serviceName, node.user, node.dependency))
}

func (node *resolutionNode) debugResolvingDependencyOnComponent() {
	if node.component.Code != nil {
		Debug.WithFields(log.Fields{
			"service":    node.service.GetName(),
			"component":  node.component.Name,
			"context":    node.context.GetName(),
			"allocation": node.context.Allocation.Name,
		}).Info("Processing dependency on code execution")
	} else if node.component.Service != "" {
		Debug.WithFields(log.Fields{
			"service":          node.service.GetName(),
			"component":        node.component.Name,
			"context":          node.context.GetName(),
			"allocation":       node.context.Allocation.Name,
			"dependsOnService": node.component.Service,
		}).Info("Processing dependency on another service")
	} else {
		Debug.WithFields(log.Fields{
			"service":   node.service.GetName(),
			"component": node.component.Name,
		}).Panic("Invalid component (not code and not service")
	}
}

func (node *resolutionNode) cannotResolve() error {
	Debug.WithFields(log.Fields{
		"service":      node.serviceName,
		"componentObj": node.component,
		"contextObj":   node.context,
	}).Info("Cannot resolve instance")

	// There may be a situation when service key has not been resolved yet. If so, we should create a fake one to attach logs to
	if node.serviceKey == nil {
		// Create service key
		node.serviceKey = node.createComponentKey(nil)

		// Once instance is figured out, make sure to attach rule logs to that instance
		node.ruleLogWriter.attachToInstance(node.serviceKey)
	}

	return nil
}

// Helper to get a matched service
func (node *resolutionNode) getMatchedService(policy *Policy) *Service {
	service := policy.Services[node.serviceName]
	node.ruleLogWriter.addRuleLogEntry(entryServiceMatched(node.serviceName, service != nil))
	return service
}

// Helper to get a matched context
func (node *resolutionNode) getMatchedContext(policy *Policy) *Context {
	// Locate the list of contexts for service
	node.ruleLogWriter.addRuleLogEntry(entryContextsFound(len(policy.Contexts) > 0))

	// Find matching context
	var contextMatched *Context
	for _, context := range policy.Contexts {
		// Get matching context
		matched := context.Matches(node.getContextualDataForExpression(), node.cache.expressionCache)
		node.ruleLogWriter.addRuleLogEntry(entryContextCriteriaTesting(context, matched))
		if matched {
			// Match is only valid if there are no global rule voilations context
			labels := node.transformLabels(node.labels, context.ChangeLabels)
			cluster := node.getCluster(policy, labels, context)
			globalRuleViolations := node.hasGlobalRuleViolations(policy, context, labels, cluster)
			node.ruleLogWriter.addRuleLogEntry(entryContextGlobalRulesNoViolations(context, !globalRuleViolations))
			if !globalRuleViolations {
				contextMatched = context
				break
			}
		}
	}

	if contextMatched != nil {
		Debug.WithFields(log.Fields{
			"service": node.service.GetName(),
			"context": contextMatched.GetName(),
			"user":    node.user.Name,
		}).Info("Matched context")
	} else {
		Debug.WithFields(log.Fields{
			"service": node.service.GetName(),
			"user":    node.user.Name,
		}).Info("No context matched")
	}

	node.ruleLogWriter.addRuleLogEntry(entryContextMatched(node.service, contextMatched))

	return contextMatched
}

// Helper to get a matched allocation
func (node *resolutionNode) resolveAllocationKeys(policy *Policy) ([]string, error) {
	// If there is no allocation, exit
	if node.context.Allocation == nil {
		return nil, errors.New("No allocation present")
	}

	// Resolve allocation keys (they can be dynamic, depending on user labels)
	result, err := node.context.ResolveKeys(node.getContextualDataForAllocationTemplate(), node.cache.templateCache)
	if err != nil {
		Debug.WithFields(log.Fields{
			"service":    node.service.GetName(),
			"context":    node.context.GetName(),
			"allocation": node.context.Allocation.Name,
			"keys":       node.context.Allocation.Keys,
			"user":       node.user.Name,
			"error":      err,
		}).Info("Cannot resolve one of the keys within an allocation")
		return nil, err
	}

	if len(node.context.Allocation.Keys) > 0 {
		Debug.WithFields(log.Fields{
			"service":    node.service.GetName(),
			"context":    node.context.GetName(),
			"allocation": node.context.Allocation.Name,
			"user":       node.user.Name,
		}).Info("All allocation keys resolved")
		node.ruleLogWriter.addRuleLogEntry(entryAllocationKeysResolved(node.service, node.context, result))
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

func (node *resolutionNode) getCluster(policy *Policy, labels LabelSet, context *Context) *Cluster {
	// todo(slukjanov): temp hack - expecting that cluster is always passed through the label "cluster"
	var cluster *Cluster
	if clusterLabel, ok := labels.Labels["cluster"]; ok {
		if cluster, ok = policy.Clusters[clusterLabel]; !ok {
			Debug.WithFields(log.Fields{
				"service": node.service,
				"context": context,
				"labels":  labels.Labels,
			}).Panic("Can't find cluster (based on label 'cluster')")
		}
	}
	return cluster
}

func (node *resolutionNode) transformLabels(labels LabelSet, operations LabelOperations) LabelSet {
	result := labels.ApplyTransform(operations)
	if !labels.Equal(result) {
		node.ruleLogWriter.addRuleLogEntry(entryLabels(result))
	}
	return result
}

func (node *resolutionNode) hasGlobalRuleViolations(policy *Policy, context *Context, labels LabelSet, cluster *Cluster) bool {
	globalRules := policy.Rules
	if rules, ok := globalRules.Rules["dependency"]; ok {
		for _, rule := range rules {
			matched := rule.FilterServices.Match(labels, node.user, cluster, node.cache.expressionCache)
			node.ruleLogWriter.addRuleLogEntry(entryContextGlobalRuleTesting(context, rule, matched))
			if matched {
				for _, action := range rule.Actions {
					if action.Type == "dependency" && action.Content == "forbid" {
						return true
					}
				}
			}
		}
	}

	return false
}

func (node *resolutionNode) calculateAndStoreCodeParams() error {
	componentCodeParams, err := evaluateParameterTree(node.component.Code.Params, node.getContextualDataForCodeDiscoveryTemplate(), node.cache.templateCache)
	node.data.recordCodeParams(node.componentKey, componentCodeParams)
	return err
}

func (node *resolutionNode) calculateAndStoreDiscoveryParams() error {
	componentDiscoveryParams, err := evaluateParameterTree(node.component.Discovery, node.getContextualDataForCodeDiscoveryTemplate(), node.cache.templateCache)
	node.data.recordDiscoveryParams(node.componentKey, componentDiscoveryParams)

	// Populate discovery tree (allow this component to announce its discovery properties in the discovery tree)
	node.discoveryTreeNode.GetNestedMap(node.component.Name)["instance"] = EscapeName(node.componentKey.GetKey())
	for k, v := range componentDiscoveryParams {
		node.discoveryTreeNode.GetNestedMap(node.component.Name)[k] = v
	}

	return err
}
