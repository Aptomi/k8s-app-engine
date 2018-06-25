package resolve

import (
	"fmt"

	"github.com/Aptomi/aptomi/pkg/event"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/plugin/k8s"
	"github.com/Aptomi/aptomi/pkg/runtime"
	"github.com/Aptomi/aptomi/pkg/util"
)

// This is a special internal structure that gets used by the engine, while we traverse the policy graph for a given claim
// It gets incrementally populated with data, as policy evaluation goes on for a given claim
type resolutionNode struct {
	// pointer to the policy resolver
	resolver *PolicyResolver

	// pointer to event log (local to the node)
	eventLog *event.Log

	// combined event logs from all resolution nodes in the subtree
	eventLogsCombined []*event.Log

	// new instance of PolicyResolution, where resolution resolution will be stored
	resolution *PolicyResolution

	// depth we are currently on (as we are traversing policy graph), with initial claim being on depth 0
	depth int

	// reference to initial claim
	claim *lang.Claim

	// reference to user who requested this claim
	user *lang.User

	// reference to the namespace & service we are currently resolving
	namespace   string
	serviceName string
	service     *lang.Service

	// reference to the current set of labels
	labels *lang.LabelSet

	// reference to the context & the corresponding bundle that were matched
	context *lang.Context
	bundle  *lang.Bundle

	// reference to the allocation keys that were resolved
	allocationKeysResolved []string

	// reference to the current node in discovery tree for components announcing their discovery properties
	// component1...component2...component3 -> component instance key
	discoveryTreeNode util.NestedParameterMap

	// reference to the current component in discovery tree
	component *lang.BundleComponent

	// reference to the current component key
	componentKey *ComponentInstanceKey

	// reference to the current bundle key
	bundleKey *ComponentInstanceKey

	// reference to the last key we arrived with, so we can reconstruct graph edges between keys
	arrivalKey *ComponentInstanceKey

	// path that we traveled so far (to detect cycles)
	path []string
}

// Creates a new empty resolution node
func (resolver *PolicyResolver) newResolutionNode() *resolutionNode {
	eventLog := event.NewLog(resolver.eventLog.GetLevel(), resolver.eventLog.GetScope())
	return &resolutionNode{
		resolver:          resolver,
		eventLog:          eventLog,
		eventLogsCombined: []*event.Log{eventLog},

		resolution: NewPolicyResolution(),

		depth: 0,

		// empty discovery tree
		discoveryTreeNode: util.NestedParameterMap{},

		// empty path
		path: []string{},
	}
}

// Initialized a newly created resolution node as a starting point for resolving a particular claim.
// Adds claim and user labels into it.
func (resolver *PolicyResolver) initResolutionNode(node *resolutionNode, claim *lang.Claim) {
	// populate user, claim
	node.claim = claim
	user := resolver.externalData.UserLoader.LoadUserByName(claim.User)
	node.user = user

	// start with the namespace & service specified in the claim
	node.namespace = claim.Namespace
	node.serviceName = claim.Service

	// create a starting set of labels, combining user labels and claim labels
	node.labels = lang.NewLabelSet(claim.Labels)
	if user != nil {
		node.labels.AddLabels(user.Labels)
	}
}

// Creates a new resolution node (as we are processing claim on another bundle)
func (node *resolutionNode) createChildNode() *resolutionNode {
	eventLog := event.NewLog(node.eventLog.GetLevel(), node.eventLog.GetScope())
	return &resolutionNode{
		resolver:          node.resolver,
		eventLog:          eventLog,
		eventLogsCombined: []*event.Log{eventLog},

		resolution: node.resolution,

		depth: node.depth + 1,
		claim: node.claim,
		user:  node.user,

		// we take the current component we are iterating over, and get its service name
		namespace:   node.namespace,
		serviceName: node.component.Service,

		// proceed with the current set of labels
		labels: lang.NewLabelSet(node.labels.Labels),

		// move further by the discovery tree via component name link
		discoveryTreeNode: node.discoveryTreeNode.GetNestedMap(node.component.Name),

		// remember the last arrival key
		arrivalKey: node.componentKey,

		// copy path
		path: util.CopySliceOfStrings(node.path),
	}
}

// As the resolution goes on, this method is called when objects become resolved and available in the context
// Right now we only call it for claim, service, and bundle
func (node *resolutionNode) objectResolved(object runtime.Storable) {
	node.eventLog.AddFixedField(object.GetKind()+"Id", runtime.KeyForStorable(object))
}

// Helper to check that user exists
func (node *resolutionNode) checkUserExists() error {
	if node.user == nil {
		return node.errorUserDoesNotExist()
	}
	return nil
}

// Helper to get a service
func (node *resolutionNode) getService(policy *lang.Policy) (*lang.Service, error) {
	serviceObj, err := policy.GetObject(lang.ServiceObject.Kind, node.serviceName, node.namespace)
	if serviceObj == nil || err != nil {
		panic(fmt.Sprintf("Can't get service '%s/%s': %s", node.namespace, node.serviceName, err))
	}
	service := serviceObj.(*lang.Service) // nolint: errcheck

	// User should have permissions to consume the service according to the ACL
	userView := node.resolver.policy.View(node.user)
	canConsume, err := userView.CanConsume(service)
	if !canConsume {
		return nil, node.userNotAllowedToConsumeService(err)
	}

	node.logServiceFound(service)
	return service, nil
}

// Helper to get a matched context
func (node *resolutionNode) getMatchedContext(policy *lang.Policy) (*lang.Context, error) {
	// Locate the list of contexts for bundle
	node.logStartMatchingContexts()

	// Find matching context
	contextualData := node.getContextualDataForContextExpression()
	var contextMatched *lang.Context
	for _, context := range node.service.Contexts {
		// Check if context matches (based on criteria)
		matched, err := context.Matches(contextualData, node.resolver.expressionCache)
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

	if contextMatched == nil {
		return nil, node.errorContextNotMatched()
	}

	node.logContextMatched(contextMatched)
	return contextMatched, nil
}

// Helper to get a matched bundle
func (node *resolutionNode) getMatchedBundle(policy *lang.Policy) (*lang.Bundle, error) {
	bundleObj, err := policy.GetObject(lang.BundleObject.Kind, node.context.Allocation.Bundle, node.namespace)
	if bundleObj == nil || err != nil {
		panic(fmt.Sprintf("Can't get bundle '%s/%s': %s", node.namespace, node.context.Allocation.Bundle, err))
	}

	bundle := bundleObj.(*lang.Bundle) // nolint: errcheck

	// Bundle should be located in the same namespace as service
	if bundle.Namespace != node.service.Namespace {
		return nil, node.errorBundleIsNotInSameNamespaceAsService(bundle)
	}

	node.logBundleFound(bundle)
	return bundle, nil
}

// Helper to resolve allocation keys
func (node *resolutionNode) resolveAllocationKeys(policy *lang.Policy) ([]string, error) {
	// Resolve allocation keys (they can be dynamic, depending on user labels)
	result, err := node.context.ResolveKeys(node.getContextualDataForContextAllocationTemplate(), node.resolver.templateCache)
	if err != nil {
		return nil, node.errorWhenResolvingAllocationKeys(err)
	}

	node.logAllocationKeysSuccessfullyResolved(result)
	return result, nil
}

// checks if component criteria holds or not (i.e. whether component should be included or excluded from processing)
func (node *resolutionNode) componentMatches(component *lang.BundleComponent) (bool, error) {
	contextualData := node.getContextualDataForComponentCriteria()
	matched, err := component.Matches(contextualData, node.resolver.expressionCache)
	if err != nil {
		// Propagate error up
		return false, node.errorWhenTestingComponent(component, err)
	}
	if !matched {
		node.logComponentNotMatched(component)
	}
	return matched, nil
}

// createComponentKey creates a component key
func (node *resolutionNode) createComponentKey(component *lang.BundleComponent) (*ComponentInstanceKey, error) {
	targetLabel := node.labels.Labels[lang.LabelTarget]
	if len(targetLabel) <= 0 {
		return nil, node.errorTargetNotSet()
	}

	target := lang.NewTarget(targetLabel)
	cluster, err := target.GetCluster(node.resolver.policy, node.namespace)
	if err != nil {
		return nil, node.errorClusterLookup(target.ClusterName, err)
	}

	// handle default namespace for kubernetes clusters
	if len(target.Suffix) <= 0 && cluster.Type == "kubernetes" {
		k8sClusterConfig := &k8s.ClusterConfig{}
		err := cluster.ParseConfigInto(k8sClusterConfig)

		// if it's a k8s cluster, let's grab default namespace from it
		if err == nil {
			target.Suffix = k8sClusterConfig.DefaultNamespace
		}

		// if it's still empty, use default
		if len(target.Suffix) <= 0 {
			target.Suffix = "default"
		}
	}

	return NewComponentInstanceKey(
		cluster,
		target.Suffix,
		node.service,
		node.context,
		node.allocationKeysResolved,
		node.bundle,
		component,
	), nil
}

func (node *resolutionNode) transformLabels(labels *lang.LabelSet, operations lang.LabelOperations) {
	changedLabels := labels.ApplyTransform(operations)
	if changedLabels {
		node.logLabels(labels, "after transform")
	}
}

func (node *resolutionNode) processRulesWithinNamespace(policyNamespace *lang.PolicyNamespace, result *lang.RuleActionResult) error {
	if policyNamespace == nil {
		return nil
	}

	rules := lang.GetRulesSortedByWeight(policyNamespace.Rules)
	contextualData := node.getContextualDataForRuleExpression()
	for _, rule := range rules {
		matched, err := rule.Matches(contextualData, node.resolver.expressionCache)
		if err != nil {
			return node.errorWhenProcessingRule(rule, err)
		}
		node.logTestedRuleMatch(rule, matched)
		if matched {
			rule.ApplyActions(result)

			// if a claim has been rejected, handle it right away and return that we cannot resolve it
			if result.RejectClaim {
				return node.errorClaimNotAllowedByRules()
			}
			if result.ChangedLabelsOnLastApply {
				node.logLabels(result.Labels, "after transform")
			}
		}
	}

	node.logRulesProcessingResult(policyNamespace, result)
	return nil
}

func (node *resolutionNode) processRules() (*lang.RuleActionResult, error) {
	result := lang.NewRuleActionResult(node.labels)

	// process rules within the current namespace
	var err = node.processRulesWithinNamespace(node.resolver.policy.Namespace[node.namespace], result)
	if err != nil {
		return nil, err
	}

	// process rules globally (within system namespace)
	err = node.processRulesWithinNamespace(node.resolver.policy.Namespace[runtime.SystemNS], result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (node *resolutionNode) calculateAndStoreCodeParams() error {
	componentCodeParams, err := util.ProcessParameterTree(node.component.Code.Params, node.getContextualDataForCodeDiscoveryTemplate(), node.resolver.templateCache, util.ModeEvaluate)
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
	componentDiscoveryParams, err := util.ProcessParameterTree(node.component.Discovery, node.getContextualDataForCodeDiscoveryTemplate(), node.resolver.templateCache, util.ModeEvaluate)
	if err != nil {
		return node.errorWhenProcessingDiscoveryParams(err)
	}

	err = node.resolution.RecordDiscoveryParams(node.componentKey, componentDiscoveryParams)
	if err != nil {
		return node.errorWhenProcessingDiscoveryParams(err)
	}

	// Populate discovery tree (allow this component to announce its discovery properties in the discovery tree)
	node.discoveryTreeNode.GetNestedMap(node.component.Name)["instance"] = util.EscapeName(node.componentKey.GetDeployName())
	for k, v := range componentDiscoveryParams {
		node.discoveryTreeNode.GetNestedMap(node.component.Name)[k] = v
	}

	return nil
}
