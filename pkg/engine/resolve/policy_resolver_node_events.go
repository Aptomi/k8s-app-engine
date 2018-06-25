package resolve

import (
	"fmt"

	"github.com/Aptomi/aptomi/pkg/errors"
	"github.com/Aptomi/aptomi/pkg/event"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/runtime"
	"github.com/davecgh/go-spew/spew"
)

/*
	Non-critical errors - if any of them occur, the corresponding claim will not be fulfilled
	and engine will move on to processing other claims
*/

func (node *resolutionNode) errorUserDoesNotExist() error {
	return fmt.Errorf("claim '%s/%s' refers to non-existing user: %s", node.claim.Metadata.Namespace, node.claim.Name, node.claim.User)
}

func (node *resolutionNode) errorClaimNotAllowedByRules() error {
	return fmt.Errorf("rules do not allow claim '%s/%s' ('%s' -> '%s'): processing '%s', tree depth %d", node.claim.Metadata.Namespace, node.claim.Name, node.claim.User, node.claim.Contract, node.contractName, node.depth)
}

func (node *resolutionNode) userNotAllowedToConsumeBundle(err error) error {
	return fmt.Errorf("user '%s' not allowed to consume bundle: %s", node.claim.User, err)
}

func (node *resolutionNode) errorTargetNotSet() error {
	return fmt.Errorf("not sure where components should be deployed: label 'target' is not set (claim '%s', contract '%s', bundle '%s')", node.claim.Name, node.contract.Name, node.bundle.Name)
}

func (node *resolutionNode) errorClusterLookup(clusterName string, cause error) error {
	return fmt.Errorf("cluster '%s' lookup error: %s (claim '%s', contract '%s', bundle '%s')", clusterName, cause, node.claim.Name, node.contract.Name, node.bundle.Name)
}

func (node *resolutionNode) errorBundleIsNotInSameNamespaceAsContract(bundle *lang.Bundle) error {
	return fmt.Errorf("bundle '%s' is not in the same namespace as contract '%s'", runtime.KeyForStorable(bundle), runtime.KeyForStorable(node.contract))
}

func (node *resolutionNode) errorWhenTestingContext(context *lang.Context, cause error) error {
	return fmt.Errorf("error while trying to match context '%s' for contract '%s': %s", context.Name, node.contract.Name, printCauseDetailsOnDebug(cause, node.eventLog))
}

func (node *resolutionNode) errorContextNotMatched() error {
	return fmt.Errorf("unable to find matching context within contract: '%s'", node.contract.Name)
}

func (node *resolutionNode) errorWhenTestingComponent(component *lang.BundleComponent, cause error) error {
	return fmt.Errorf("error while checking component criteria '%s' for bundle '%s': %s", component.Name, node.bundle.Name, printCauseDetailsOnDebug(cause, node.eventLog))
}

func (node *resolutionNode) errorWhenProcessingRule(rule *lang.Rule, cause error) error {
	return fmt.Errorf("error while processing rule '%s' on contract '%s', context '%s', bundle '%s': %s", rule.Name, node.contract.Name, node.context.Name, node.bundle.Name, printCauseDetailsOnDebug(cause, node.eventLog))
}

func (node *resolutionNode) errorWhenResolvingAllocationKeys(cause error) error {
	return fmt.Errorf("error while resolving allocation keys for contract '%s', context '%s': %s", node.contract.Name, node.context.Name, printCauseDetailsOnDebug(cause, node.eventLog))
}

func (node *resolutionNode) errorWhenProcessingCodeParams(cause error) error {
	return fmt.Errorf("error when processing code params for bundle '%s', contract '%s', context '%s', component '%s': %s", node.bundle.Name, node.contract.Name, node.context.Name, node.component.Name, printCauseDetailsOnDebug(cause, node.eventLog))
}

func (node *resolutionNode) errorWhenProcessingDiscoveryParams(cause error) error {
	return fmt.Errorf("error when processing discovery params for bundle '%s', contract '%s', context '%s', component '%s': %s", node.bundle.Name, node.contract.Name, node.context.Name, node.component.Name, printCauseDetailsOnDebug(cause, node.eventLog))
}

func (node *resolutionNode) errorBundleCycleDetected() error {
	return fmt.Errorf("error when processing policy, bundle cycle detected: %s", node.path)
}

/*
	Event log - report debug/info/warning messages
*/

func (node *resolutionNode) logStartResolvingClaim() {
	if node.depth == 0 {
		// at the top of the tree, when we resolve a root-level claim
		node.eventLog.NewEntry().Infof("Resolving top-level claim '%s/%s' ('%s' -> '%s')", node.claim.Metadata.Namespace, node.claim.Name, node.claim.User, node.claim.Contract)
	} else {
		// recursively processing the rest of the tree
		node.eventLog.NewEntry().Infof("Resolving claim '%s/%s' ('%s' -> '%s'): processing '%s', tree depth %d", node.claim.Metadata.Namespace, node.claim.Name, node.claim.User, node.claim.Contract, node.contractName, node.depth)
	}

	node.logLabels(node.labels, "initial")
}

func (node *resolutionNode) logLabels(labelSet *lang.LabelSet, scope string) {
	secretCnt := 0
	if node.user != nil {
		secretCnt = len(node.resolver.externalData.SecretLoader.LoadSecretsByUserName(node.user.Name))
	}
	node.eventLog.NewEntry().Infof("Labels (%s): %s and %d secrets", scope, labelSet.Labels, secretCnt)
}

func (node *resolutionNode) logContractFound(contract *lang.Contract) {
	node.eventLog.NewEntry().Debugf("Contract found in policy: '%s'", contract.Name)
}

func (node *resolutionNode) logBundleFound(bundle *lang.Bundle) {
	node.eventLog.NewEntry().Debugf("Bundle found in policy: '%s'", bundle.Name)
}

func (node *resolutionNode) logStartMatchingContexts() {
	contextNames := []string{}
	for _, context := range node.contract.Contexts {
		contextNames = append(contextNames, context.Name)
	}
	node.eventLog.NewEntry().Infof("Picking context within contract '%s'. Trying contexts: %s", node.contract.Name, contextNames)
}

func (node *resolutionNode) logContextMatched(contextMatched *lang.Context) {
	node.eventLog.NewEntry().Infof("Found matching context within contract '%s': %s", node.contract.Name, contextMatched.Name)
}

func (node *resolutionNode) logComponentNotMatched(component *lang.BundleComponent) {
	node.eventLog.NewEntry().Infof("Component criteria evaluated to 'false', excluding it from processing: bundle '%s', component '%s'", node.bundle.Name, node.component.Name)
}

func (node *resolutionNode) logTestedContextCriteria(context *lang.Context, matched bool) {
	node.eventLog.NewEntry().Debugf("Trying context '%s' within contract '%s'. Matched = %t", context.Name, node.contract.Name, matched)
}

func (node *resolutionNode) logRulesProcessingResult(policyNamespace *lang.PolicyNamespace, result *lang.RuleActionResult) {
	node.eventLog.NewEntry().Debugf("Rules processed within namespace '%s' for context '%s' within contract '%s'", policyNamespace.Name, node.context.Name, node.contract.Name)
}

func (node *resolutionNode) logTestedRuleMatch(rule *lang.Rule, match bool) {
	node.eventLog.NewEntry().Debugf("Testing if rule '%s' applies in context '%s' within contract '%s'. Result: %t", rule.Name, node.context.Name, node.contract.Name, match)
}

func (node *resolutionNode) logAllocationKeysSuccessfullyResolved(resolvedKeys []string) {
	if len(resolvedKeys) > 0 {
		node.eventLog.NewEntry().Infof("Allocation keys successfully resolved for context '%s' within contract '%s': %s", node.context.Name, node.contract.Name, resolvedKeys)
	}
}

func (node *resolutionNode) logResolvingClaimOnComponent() {
	if node.component.Code != nil {
		node.eventLog.NewEntry().Infof("Processing claim on component with code: %s (%s)", node.component.Name, node.component.Code.Type)
	} else if node.component.Contract != "" {
		node.eventLog.NewEntry().Infof("Processing claim on another contract: %s", node.component.Contract)
	} else {
		node.eventLog.NewEntry().Warningf("Skipping unknown component (not code and not contract): %s", node.component.Name)
	}
}

func (node *resolutionNode) logInstanceSuccessfullyResolved(cik *ComponentInstanceKey) {
	if node.depth == 0 && cik.IsBundle() {
		// at the top of the tree, when we resolve a root-level claim
		node.eventLog.NewEntry().Infof("Successfully resolved claim '%s/%s' ('%s' -> '%s'): %s", node.claim.Metadata.Namespace, node.claim.Name, node.user.Name, node.claim.Contract, cik.GetKey())
	} else if cik.IsBundle() {
		// resolved bundle instance
		node.eventLog.NewEntry().Infof("Successfully resolved bundle instance '%s' -> '%s': %s", node.user.Name, node.contract.Name, cik.GetKey())
	} else {
		// resolved component instance
		node.eventLog.NewEntry().Infof("Successfully resolved component instance '%s' -> '%s' (component '%s'): %s", node.user.Name, node.contract.Name, node.component.Name, cik.GetKey())
	}
}

func (node *resolutionNode) logCannotResolveInstance() {
	if node.bundle == nil {
		node.eventLog.NewEntry().Warningf("Cannot resolve instance: contract '%s'", node.contractName)
	} else if node.component == nil {
		node.eventLog.NewEntry().Warningf("Cannot resolve instance: contract '%s', bundle '%s'", node.contractName, node.bundle.Name)
	} else {
		node.eventLog.NewEntry().Warningf("Cannot resolve instance: contract '%s', bundle '%s', component '%s'", node.contractName, node.bundle.Name, node.component.Name)
	}
}

func (resolver *PolicyResolver) logComponentParams(instance *ComponentInstance) {
	bundleObj, err := resolver.policy.GetObject(lang.BundleObject.Kind, instance.Metadata.Key.BundleName, instance.Metadata.Key.Namespace)
	if err != nil {
		panic(fmt.Sprintf("error while getting bundle '%s/%s' from the policy: %s", instance.Metadata.Key.BundleName, instance.Metadata.Key.Namespace, err))
	}

	// if there is a conflict (e.g. components have different code params), turn this into an error
	if instance.Error != nil {
		resolver.eventLog.NewEntry().Error(printCauseDetailsOnDebug(instance.Error, resolver.eventLog))
	}

	code := bundleObj.(*lang.Bundle).GetComponentsMap()[instance.Metadata.Key.ComponentName].Code
	if code != nil {
		cs := spew.ConfigState{Indent: "\t"}

		// log code params
		resolver.eventLog.NewEntry().Debugf("Calculated final code params for component '%s': %s", instance.Metadata.Key.GetKey(), cs.Sdump(instance.CalculatedCodeParams))

		// log discovery params
		resolver.eventLog.NewEntry().Debugf("Calculated final discovery params for component '%s': %s", instance.Metadata.Key.GetKey(), cs.Sdump(instance.CalculatedDiscovery))
	}
}

// if the given argument is ErrorWithDetails, it logs its details on debug mode
func printCauseDetailsOnDebug(err error, eventLog *event.Log) error {
	errWithDetails, isErrorWithDetails := err.(*errors.ErrorWithDetails)
	if isErrorWithDetails {
		cs := spew.ConfigState{Indent: "\t"}
		eventLog.NewEntry().Debug(cs.Sdump(errWithDetails.Details()))
	}
	return err
}
