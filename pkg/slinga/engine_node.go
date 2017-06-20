package slinga

import (
	"errors"
	log "github.com/Sirupsen/logrus"
)

// This is a special internal structure that gets used by the engine, while we traverse the policy graph for a given dependency
// It gets incrementally populated with data, as policy evaluation goes on for a given dependency
type resolutionNode struct {
	// whether we successfully resolved this node or not
	resolved bool

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
}

// Creates a new resolution node as a starting point for resolving a particular dependency
func (usage *ServiceUsageState) newResolutionNode(d *Dependency) *resolutionNode {
	user := usage.users.Users[d.UserID]
	if user == nil {
		// Resolving allocations for service
		debug.WithFields(log.Fields{
			"dependency": d,
		}).Fatal("Dependency refers to non-existing user")
	}
	return &resolutionNode{
		resolved:   false,
		depth:      0,
		dependency: d,
		user:       user,

		// we start with the service specified in the dependency
		serviceName: d.Service,

		// combining user labels and dependency labels
		labels: user.getLabelSet().addLabels(user.getSecretSet()).addLabels(d.getLabelSet()),

		// empty discovery tree
		discoveryTreeNode: NestedParameterMap{},
	}
}

// Creates a new resolution node (as we are processing dependency on another service)
func (node *resolutionNode) createChildNode() *resolutionNode {
	return &resolutionNode{
		resolved:   false,
		depth:      node.depth + 1,
		dependency: node.dependency,
		user:       node.user,

		// we take the current component we are iterating over, and get its service name
		serviceName: node.component.Service,

		// we take current processed labels for the component we are iterating over
		labels: node.componentLabels,

		// move further by the discovery tree via component name link
		discoveryTreeNode: node.discoveryTreeNode.getNestedMap(node.component.Name),
	}
}

func (node *resolutionNode) debugResolvingDependency() {
	// Resolving allocations for service
	debug.WithFields(log.Fields{
		"service": node.serviceName,
		"user":    node.user.Name,
		"labels":  node.labels,
	}).Info("Resolving allocations for service")

	trace.Printf(node.depth, "[Dependency]")
	trace.Printf(node.depth, "User: %s (ID = %s)", node.user.Name, node.user.ID)
	trace.Printf(node.depth+1, "Labels: %s", node.labels)
	trace.Printf(node.depth, "Service: %s", node.serviceName)
}

func (node *resolutionNode) debugResolvedContext() {
	trace.Printf(node.depth, "Context: %s", node.context.Name)
}

func (node *resolutionNode) debugResolvedAllocation() {
	trace.Printf(node.depth, "Allocation: %s", node.allocation.NameResolved)
}

func (node *resolutionNode) debugResolvingDependencyOnComponent() {
	if node.component.Code != nil {
		debug.WithFields(log.Fields{
			"service":    node.service.Name,
			"component":  node.component.Name,
			"context":    node.context.Name,
			"allocation": node.allocation.NameResolved,
		}).Info("Processing dependency on code execution")
	} else if node.component.Service != "" {
		debug.WithFields(log.Fields{
			"service":          node.service.Name,
			"component":        node.component.Name,
			"context":          node.context.Name,
			"allocation":       node.allocation.NameResolved,
			"dependsOnService": node.component.Service,
		}).Info("Processing dependency on another service")

		trace.Println()
	} else {
		debug.WithFields(log.Fields{
			"service":   node.service.Name,
			"component": node.component.Name,
		}).Fatal("Invalid component (not code and not service")
	}
}

func (node *resolutionNode) cannotResolveDependency() error {
	debug.WithFields(log.Fields{
		"service":          node.service.Name,
		"component":        node.component.Name,
		"context":          node.context.Name,
		"allocation":       node.allocation.NameResolved,
		"dependsOnService": node.component.Service,
	}).Info("Cannot fulfill dependency on another service")

	return nil
}

func (node *resolutionNode) getMatchedService(policy *Policy) (*Service, error) {
	// Locate the service
	service := policy.Services[node.serviceName]
	if service == nil {
		trace.Printf(node.depth+1, "Error while trying to look up service %s (not found)", node.serviceName)
		return nil, errors.New("Service " + node.serviceName + " not found")
	}
	return service, nil
}

// Helper to get a matched context
func (node *resolutionNode) getMatchedContext(policy *Policy) (*Context, error) {
	// Locate the list of contexts for service
	contexts, ok := policy.Contexts[node.service.Name]
	if !ok || len(contexts) <= 0 {
		trace.Printf(node.depth+1, "Error while matching context for service %s (no contexts found)", node.service.Name)
		return nil, errors.New("No contexts found for " + node.service.Name)
	}

	// See which context matches
	var contextMatched *Context
	for _, c := range contexts {
		m := c.matches(node.labels)
		trace.Printf(node.depth+1, "[%t] Testing context '%s': (criteria = %+v)", m, c.Name, c.Criteria)
		if m {
			contextMatched = c
			break
		}
	}

	if contextMatched != nil {
		debug.WithFields(log.Fields{
			"service": node.service.Name,
			"context": contextMatched.Name,
			"user":    node.user.Name,
		}).Info("Matched context")
	} else {
		trace.Printf(node.depth+1, "No context matched for service %s", node.service.Name)
		debug.WithFields(log.Fields{
			"service": node.service.Name,
			"user":    node.user.Name,
		}).Info("No context matched")
	}
	return contextMatched, nil
}

// Helper to get a matched allocation
func (node *resolutionNode) getMatchedAllocation(policy *Policy) (*Allocation, error) {
	if len(node.context.Allocations) <= 0 {
		trace.Printf(node.depth, "Error while matching allocation for service %s, context %s (no allocations found)", node.service.Name, node.context.Name)
		return nil, errors.New("No allocations found for " + node.service.Name)
	}

	// See which allocation matches
	var allocationMatched *Allocation
	for _, a := range node.context.Allocations {
		m := a.matches(node.labels)
		trace.Printf(node.depth+1, "[%t] Testing allocation '%s': (criteria = %+v)", m, a.Name, a.Criteria)
		if !m {
			continue
		}

		// use labels for allocation
		labels := node.transformLabels(node.labels, a.Labels)

		// todo(slukjanov): temp hack - expecting that cluster is always passed through the label "cluster"
		var cluster *Cluster
		if clusterLabel, ok := labels.Labels["cluster"]; ok {
			if cluster, ok = policy.Clusters[clusterLabel]; !ok {
				debug.WithFields(log.Fields{
					"allocation": a,
					"labels":     labels.Labels,
				}).Fatal("Can't find cluster for allocation (based on label 'cluster')")
			}
		}

		if policy.Rules.allowsAllocation(a, labels, node, cluster) {
			allocationMatched = a
			break
		}
	}

	// Check errors and resolve allocation name (it can be dynamic, depending on user labels)
	if allocationMatched != nil {
		err := allocationMatched.resolveName(node.user, node.labels)
		if err != nil {
			debug.WithFields(log.Fields{
				"service":    node.service.Name,
				"context":    node.context.Name,
				"allocation": allocationMatched.Name,
				"user":       node.user.Name,
				"error":      err,
			}).Fatal("Cannot resolve name for an allocation")
		}
		debug.WithFields(log.Fields{
			"service":            node.service.Name,
			"context":            node.context.Name,
			"allocation":         allocationMatched.Name,
			"allocationResolved": allocationMatched.NameResolved,
			"user":               node.user.Name,
		}).Info("Matched allocation")
	} else {
		trace.Printf(node.depth+1, "No allocation matched for service %s, context %s", node.service.Name, node.context.Name)
		debug.WithFields(log.Fields{
			"service": node.service.Name,
			"context": node.context.Name,
			"user":    node.user.Name,
		}).Info("No allocation matched")
	}

	return allocationMatched, nil
}

func (node *resolutionNode) transformLabels(labels LabelSet, operations *LabelOperations) LabelSet {
	result := labels.applyTransform(operations)
	if !labels.equal(result) {
		trace.Printf(node.depth+1, "Labels changed: %s", result)
	}
	return result
}

func (node *resolutionNode) calculateAndStoreCodeParams(resolvedUsage *ResolvedServiceUsageData) error {
	componentCodeParams, err := node.component.processTemplateParams(node.component.Code.Params, node.componentKey, node.componentLabels, node.user, node.discoveryTreeNode, "code", node.depth)
	resolvedUsage.storeCodeParams(node.componentKey, componentCodeParams)
	return err
}

func (node *resolutionNode) calculateAndStoreDiscoveryParams(resolvedUsage *ResolvedServiceUsageData) error {
	componentDiscoveryParams, err := node.component.processTemplateParams(node.component.Discovery, node.componentKey, node.componentLabels, node.user, node.discoveryTreeNode, "discovery", node.depth)
	resolvedUsage.storeDiscoveryParams(node.componentKey, componentDiscoveryParams)

	// Populate discovery tree (allow this component to announce its discovery properties in the discovery tree)
	node.discoveryTreeNode.getNestedMap(node.component.Name)["instance"] = EscapeName(node.componentKey)
	for k, v := range componentDiscoveryParams {
		node.discoveryTreeNode.getNestedMap(node.component.Name)[k] = v
	}

	return err
}
