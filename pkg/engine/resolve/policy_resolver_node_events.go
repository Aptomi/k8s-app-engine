package resolve

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/errors"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/runtime"
)

/*
	Non-critical errors - if any of them occur, the corresponding dependency will not be fulfilled
	and engine will move on to processing other dependencies
*/

func (node *resolutionNode) errorUserDoesNotExist() error {
	return fmt.Errorf("dependency '%s/%s' refers to non-existing user: %s", node.dependency.Metadata.Namespace, node.dependency.Name, node.dependency.User)
}

func (node *resolutionNode) errorDependencyNotAllowedByRules() error {
	return fmt.Errorf("rules do not allow dependency '%s/%s' ('%s' -> '%s'): processing '%s', tree depth %d", node.dependency.Metadata.Namespace, node.dependency.Name, node.dependency.User, node.dependency.Contract, node.contractName, node.depth)
}

func (node *resolutionNode) userNotAllowedToConsumeService(err error) error {
	return fmt.Errorf("user '%s' not allowed to consume service: %s", node.dependency.User, err)
}

func (node *resolutionNode) errorClusterDoesNotExist(clusterName string) error {
	if len(clusterName) > 0 {
		return fmt.Errorf("cluster '%s/%s' doesn't exist in policy", runtime.SystemNS, clusterName)
	}
	return fmt.Errorf("engine needs cluster defined, but cluster is not set")
}

func (node *resolutionNode) errorServiceIsNotInSameNamespaceAsContract(service *lang.Service) error {
	return fmt.Errorf("service '%s' is not in the same namespace as contract '%s'", runtime.KeyForStorable(service), runtime.KeyForStorable(node.contract))
}

func (node *resolutionNode) errorWhenTestingContext(context *lang.Context, cause error) error {
	return fmt.Errorf("error while trying to match context '%s' for contract '%s': %s", context.Name, node.contract.Name, node.printCauseDetailsOnDebug(cause))
}

func (node *resolutionNode) errorContextNotMatched() error {
	return fmt.Errorf("unable to find matching context within contract: '%s'", node.contract.Name)
}

func (node *resolutionNode) errorWhenTestingComponent(component *lang.ServiceComponent, cause error) error {
	return fmt.Errorf("error while checking component criteria '%s' for service '%s': %s", component.Name, node.service.Name, node.printCauseDetailsOnDebug(cause))
}

func (node *resolutionNode) errorWhenProcessingRule(rule *lang.Rule, cause error) error {
	return fmt.Errorf("error while processing rule '%s' on contract '%s', context '%s', service '%s': %s", rule.Name, node.contract.Name, node.context.Name, node.service.Name, node.printCauseDetailsOnDebug(cause))
}

func (node *resolutionNode) errorWhenResolvingAllocationKeys(cause error) error {
	return fmt.Errorf("error while resolving allocation keys for contract '%s', context '%s': %s", node.contract.Name, node.context.Name, node.printCauseDetailsOnDebug(cause))
}

func (node *resolutionNode) errorWhenProcessingCodeParams(cause error) error {
	return fmt.Errorf("error when processing code params for service '%s', contract '%s', context '%s', component '%s': %s", node.service.Name, node.contract.Name, node.context.Name, node.component.Name, node.printCauseDetailsOnDebug(cause))
}

func (node *resolutionNode) errorWhenProcessingDiscoveryParams(cause error) error {
	return fmt.Errorf("error when processing discovery params for service '%s', contract '%s', context '%s', component '%s': %s", node.service.Name, node.contract.Name, node.context.Name, node.component.Name, node.printCauseDetailsOnDebug(cause))
}

func (node *resolutionNode) errorServiceCycleDetected() error {
	return fmt.Errorf("error when processing policy, service cycle detected: %s", node.path)
}

/*
	Event log - report debug/info/warning messages
*/

func (node *resolutionNode) logStartResolvingDependency() {
	if node.depth == 0 {
		// at the top of the tree, when we resolve a root-level dependency
		node.eventLog.NewEntry().Infof("Resolving top-level dependency '%s/%s' ('%s' -> '%s')", node.dependency.Metadata.Namespace, node.dependency.Name, node.dependency.User, node.dependency.Contract)
	} else {
		// recursively processing sub-dependencies
		node.eventLog.NewEntry().Infof("Resolving dependency '%s/%s' ('%s' -> '%s'): processing '%s', tree depth %d", node.dependency.Metadata.Namespace, node.dependency.Name, node.dependency.User, node.dependency.Contract, node.contractName, node.depth)
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

func (node *resolutionNode) logServiceFound(service *lang.Service) {
	node.eventLog.NewEntry().Debugf("Service found in policy: '%s'", service.Name)
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

func (node *resolutionNode) logComponentNotMatched(component *lang.ServiceComponent) {
	node.eventLog.NewEntry().Infof("Component criteria evaluated to 'false', excluding it from processing: service '%s', component '%s'", node.service.Name, node.component.Name)
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

func (node *resolutionNode) logResolvingDependencyOnComponent() {
	if node.component.Code != nil {
		node.eventLog.NewEntry().Infof("Processing dependency on component with code: %s (%s)", node.component.Name, node.component.Code.Type)
	} else if node.component.Contract != "" {
		node.eventLog.NewEntry().Infof("Processing dependency on another contract: %s", node.component.Contract)
	} else {
		node.eventLog.NewEntry().Warningf("Skipping unknown component (not code and not contract): %s", node.component.Name)
	}
}

func (node *resolutionNode) logInstanceSuccessfullyResolved(cik *ComponentInstanceKey) {
	if node.depth == 0 && cik.IsService() {
		// at the top of the tree, when we resolve a root-level dependency
		node.eventLog.NewEntry().Infof("Successfully resolved dependency '%s/%s' ('%s' -> '%s'): %s", node.dependency.Metadata.Namespace, node.dependency.Name, node.user.Name, node.dependency.Contract, cik.GetKey())
	} else if cik.IsService() {
		// resolved service instance
		node.eventLog.NewEntry().Infof("Successfully resolved service instance '%s' -> '%s': %s", node.user.Name, node.contract.Name, cik.GetKey())
	} else {
		// resolved component instance
		node.eventLog.NewEntry().Infof("Successfully resolved component instance '%s' -> '%s' (component '%s'): %s", node.user.Name, node.contract.Name, node.component.Name, cik.GetKey())
	}
}

func (node *resolutionNode) logCannotResolveInstance() {
	if node.service == nil {
		node.eventLog.NewEntry().Warningf("Cannot resolve instance: contract '%s'", node.contractName)
	} else if node.component == nil {
		node.eventLog.NewEntry().Warningf("Cannot resolve instance: contract '%s', service '%s'", node.contractName, node.service.Name)
	} else {
		node.eventLog.NewEntry().Warningf("Cannot resolve instance: contract '%s', service '%s', component '%s'", node.contractName, node.service.Name, node.component.Name)
	}
}

func (resolver *PolicyResolver) logComponentCodeParams(instance *ComponentInstance) {
	serviceObj, err := resolver.policy.GetObject(lang.ServiceObject.Kind, instance.Metadata.Key.ServiceName, instance.Metadata.Key.Namespace)
	if err != nil {
		panic(fmt.Sprintf("error while getting service '%s/%s' from the policy: %s", instance.Metadata.Key.ServiceName, instance.Metadata.Key.Namespace, err))
	}
	code := serviceObj.(*lang.Service).GetComponentsMap()[instance.Metadata.Key.ComponentName].Code
	if code != nil {
		resolver.eventLog.NewEntry().Debugf("Calculated final code params for component '%s': %v", instance.Metadata.Key.GetKey(), instance.CalculatedCodeParams)
	}
}

func (resolver *PolicyResolver) logComponentDiscoveryParams(instance *ComponentInstance) {
	serviceObj, err := resolver.policy.GetObject(lang.ServiceObject.Kind, instance.Metadata.Key.ServiceName, instance.Metadata.Key.Namespace)
	if err != nil {
		panic(fmt.Sprintf("error while getting service '%s/%s' from the policy: %s", instance.Metadata.Key.ServiceName, instance.Metadata.Key.Namespace, err))
	}
	code := serviceObj.(*lang.Service).GetComponentsMap()[instance.Metadata.Key.ComponentName].Code
	if code != nil {
		resolver.eventLog.NewEntry().Debugf("Calculated final discovery params for component '%s': %v", instance.Metadata.Key.GetKey(), instance.CalculatedDiscovery)
	}
}

// if the given argument is ErrorWithDetails, it logs its details on debug mode
func (node *resolutionNode) printCauseDetailsOnDebug(err error) error {
	errWithDetails, isErrorWithDetails := err.(*errors.ErrorWithDetails)
	if isErrorWithDetails {
		node.eventLog.NewEntry().Debugf("Error details: %v", errWithDetails.Details())
	}
	return err
}
