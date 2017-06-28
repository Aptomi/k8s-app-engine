package slinga

import (
	. "github.com/Frostman/aptomi/pkg/slinga/language"
	. "github.com/Frostman/aptomi/pkg/slinga/log"
	. "github.com/Frostman/aptomi/pkg/slinga/util"
	log "github.com/Sirupsen/logrus"
)

// This is a special internal structure that gets used by the engine, while we traverse the policy graph for a given dependency
// It gets incrementally populated with data, as policy evaluation goes on for a given dependency
type resolutionNode struct {
	// whether we successfully resolved this node or not
	resolved bool

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

	// reference to the context that was matched
	allocation *Allocation

	// reference to the current node in discovery tree for components announcing their discovery properties
	// component1...component2...component3 -> component instance key
	DiscoveryTree NestedParameterMap

	discoveryTreeNode NestedParameterMap

	// reference to the current component in discovery tree
	component *ServiceComponent

	// reference to the current component key
	componentKey string

	// reference to the current labels during component processing
	componentLabels LabelSet

	// reference to the current service key
	serviceKey string

	// reference to the last key we arrived with, so we can reconstruct graph edges between keys
	arrivalKey string

	// reference to rule log writer
	ruleLogWriter *RuleLogWriter
}

// Creates a new resolution node as a starting point for resolving a particular dependency
func (state *ServiceUsageState) newResolutionNode(dependency *Dependency) *resolutionNode {

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

		data:          data,
		ruleLogWriter: NewRuleLogWriter(data, dependency),

		depth:      0,
		dependency: dependency,
		user:       user,

		// we start with the service specified in the dependency
		serviceName: dependency.Service,

		// combining user labels and dependency labels
		labels: user.GetLabelSet().AddLabels(user.GetSecretSet()).AddLabels(dependency.GetLabelSet()),

		// empty discovery tree
		discoveryTreeNode: NestedParameterMap{},
	}
}

// Creates a new resolution node (as we are processing dependency on another service)
func (node *resolutionNode) createChildNode() *resolutionNode {
	return &resolutionNode{
		resolved: false,

		data:          node.data,
		ruleLogWriter: NewRuleLogWriter(node.data, node.dependency),

		depth:      node.depth + 1,
		dependency: node.dependency,
		user:       node.user,

		// we take the current component we are iterating over, and get its service name
		serviceName: node.component.Service,

		// we take current processed labels for the component we are iterating over
		labels: node.componentLabels,

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
			"service":    node.service.Name,
			"component":  node.component.Name,
			"context":    node.context.Name,
			"allocation": node.allocation.NameResolved,
		}).Info("Processing dependency on code execution")
	} else if node.component.Service != "" {
		Debug.WithFields(log.Fields{
			"service":          node.service.Name,
			"component":        node.component.Name,
			"context":          node.context.Name,
			"allocation":       node.allocation.NameResolved,
			"dependsOnService": node.component.Service,
		}).Info("Processing dependency on another service")
	} else {
		Debug.WithFields(log.Fields{
			"service":   node.service.Name,
			"component": node.component.Name,
		}).Panic("Invalid component (not code and not service")
	}
}

func (node *resolutionNode) cannotResolve() error {
	Debug.WithFields(log.Fields{
		"service":       node.serviceName,
		"componentObj":  node.component,
		"contextObj":    node.context,
		"allocationObj": node.allocation,
	}).Info("Cannot resolve instance")

	// There may be a situation when service key has not been resolved yet. If so, we should create a fake one to attach logs to
	if len(node.serviceKey) <= 0 {
		// Create service key
		node.serviceKey = createServiceUsageKey(node.serviceName, node.context, node.allocation, nil)

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
func (node *resolutionNode) getMatchedContext(policy *Policy) (*Context, error) {
	// Locate the list of contexts for service
	contexts := policy.Contexts[node.service.Name]
	node.ruleLogWriter.addRuleLogEntry(entryContextsFound(node.service, len(contexts) > 0))

	// Find matching context
	var contextMatched *Context
	for _, context := range contexts {
		matched := context.Matches(node.labels)
		node.ruleLogWriter.addRuleLogEntry(entryContextCriteriaTesting(context, matched))
		if matched {
			contextMatched = context
			break
		}
	}

	if contextMatched != nil {
		Debug.WithFields(log.Fields{
			"service": node.service.Name,
			"context": contextMatched.Name,
			"user":    node.user.Name,
		}).Info("Matched context")
	} else {
		Debug.WithFields(log.Fields{
			"service": node.service.Name,
			"user":    node.user.Name,
		}).Info("No context matched")
	}

	node.ruleLogWriter.addRuleLogEntry(entryContextMatched(node.service, contextMatched))

	return contextMatched, nil
}

// Helper to get a matched allocation
func (node *resolutionNode) getMatchedAllocation(policy *Policy) (*Allocation, error) {
	node.ruleLogWriter.addRuleLogEntry(entryAllocationsFound(node.service, node.context, len(node.context.Allocations) > 0))

	// Find matching allocation
	var allocationMatched *Allocation
	for _, allocation := range node.context.Allocations {
		matched := allocation.Matches(node.labels)
		node.ruleLogWriter.addRuleLogEntry(entryAllocationCriteriaTesting(allocation, matched))
		if !matched {
			continue
		}

		// use labels for allocation
		labels := node.transformLabels(node.labels, allocation.Labels)

		// todo(slukjanov): temp hack - expecting that cluster is always passed through the label "cluster"
		var cluster *Cluster
		if clusterLabel, ok := labels.Labels["cluster"]; ok {
			if cluster, ok = policy.Clusters[clusterLabel]; !ok {
				Debug.WithFields(log.Fields{
					"allocation": allocation,
					"labels":     labels.Labels,
				}).Panic("Can't find cluster for allocation (based on label 'cluster')")
			}
		}

		matched = node.allowsAllocation(policy, allocation, labels, cluster)
		node.ruleLogWriter.addRuleLogEntry(entryAllocationGlobalRulesNoViolations(allocation, matched))
		if matched {
			allocationMatched = allocation
			break
		}
	}

	// Check errors and resolve allocation name (it can be dynamic, depending on user labels)
	if allocationMatched != nil {
		err := allocationMatched.ResolveName(node.user, node.labels)
		if err != nil {
			Debug.WithFields(log.Fields{
				"service":    node.service.Name,
				"context":    node.context.Name,
				"allocation": allocationMatched.Name,
				"user":       node.user.Name,
				"error":      err,
			}).Panic("Cannot resolve name for an allocation")
		}
		Debug.WithFields(log.Fields{
			"service":            node.service.Name,
			"context":            node.context.Name,
			"allocation":         allocationMatched.Name,
			"allocationResolved": allocationMatched.NameResolved,
			"user":               node.user.Name,
		}).Info("Matched allocation")
	} else {
		Debug.WithFields(log.Fields{
			"service": node.service.Name,
			"context": node.context.Name,
			"user":    node.user.Name,
		}).Info("No allocation matched")
	}

	node.ruleLogWriter.addRuleLogEntry(entryAllocationMatched(node.service, node.context, allocationMatched))

	return allocationMatched, nil
}

func (node *resolutionNode) transformLabels(labels LabelSet, operations *LabelOperations) LabelSet {
	result := labels.ApplyTransform(operations)
	if !labels.Equal(result) {
		node.ruleLogWriter.addRuleLogEntry(entryLabels(result))
	}
	return result
}

func (node *resolutionNode) allowsAllocation(policy *Policy, allocation *Allocation, labels LabelSet, cluster *Cluster) bool {
	globalRules := policy.Rules
	if rules, ok := globalRules.Rules["dependency"]; ok {
		for _, rule := range rules {
			matched := rule.FilterServices.Match(labels, node.user, cluster)
			node.ruleLogWriter.addRuleLogEntry(entryAllocationGlobalRuleTesting(allocation, rule, matched))
			if matched {
				for _, action := range rule.Actions {
					if action.Type == "dependency" && action.Content == "forbid" {
						return false
					}
				}
			}
		}
	}

	return true
}

func (node *resolutionNode) calculateAndStoreCodeParams() error {
	componentCodeParams, err := ProcessTemplateParams(node.component.Code.Params, node.componentKey, node.componentLabels, node.user, node.discoveryTreeNode)
	node.data.recordCodeParams(node.componentKey, componentCodeParams)
	return err
}

func (node *resolutionNode) calculateAndStoreDiscoveryParams() error {
	componentDiscoveryParams, err := ProcessTemplateParams(node.component.Discovery, node.componentKey, node.componentLabels, node.user, node.discoveryTreeNode)
	node.data.recordDiscoveryParams(node.componentKey, componentDiscoveryParams)

	// Populate discovery tree (allow this component to announce its discovery properties in the discovery tree)
	node.discoveryTreeNode.GetNestedMap(node.component.Name)["instance"] = EscapeName(node.componentKey)
	for k, v := range componentDiscoveryParams {
		node.discoveryTreeNode.GetNestedMap(node.component.Name)[k] = v
	}

	return err
}
