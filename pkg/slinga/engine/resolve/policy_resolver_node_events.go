package resolve

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/slinga/errors"
	"github.com/Aptomi/aptomi/pkg/slinga/event"
	"github.com/Aptomi/aptomi/pkg/slinga/lang"
	"github.com/Aptomi/aptomi/pkg/slinga/object"
	"strings"
)

/*
	Non-critical errors. If any of them occur, the corresponding dependency will not be fulfilled
	and engine will move on to processing other dependencies
*/

func (node *resolutionNode) errorUserDoesNotExist() error {
	return errors.NewErrorWithDetails(
		fmt.Sprintf("Dependency refers to non-existing user: "+node.dependency.UserID),
		errors.Details{},
	)
}

func (node *resolutionNode) errorContractDoesNotExist() error {
	return errors.NewErrorWithDetails(
		fmt.Sprintf("Dependency refers to non-existing contract: "+node.dependency.Contract),
		errors.Details{},
	)
}

func (node *resolutionNode) errorDependencyNotAllowedByRules() error {
	userName := node.dependency.UserID
	if node.user != nil {
		userName = node.user.Name
	}
	return errors.NewErrorWithDetails(
		fmt.Sprintf("Rules do not allow dependency: '%s' -> '%s' (processing '%s', tree depth %d)", userName, node.dependency.Contract, node.contractName, node.depth),
		errors.Details{},
	)
}

/*
	Critical errors. If one of them occurs, engine will report an error and fail policy processing
	all together
*/

func (node *resolutionNode) errorClusterDoesNotExist() error {
	var err *errors.ErrorWithDetails
	if label, ok := node.labels.Labels[lang.LabelCluster]; ok {
		err = errors.NewErrorWithDetails(
			fmt.Sprintf("Cluster '%s/%s' doesn't exist in policy", object.SystemNS, label),
			errors.Details{},
		)
	} else {
		err = errors.NewErrorWithDetails(
			fmt.Sprintf("Engine needs cluster defined, but 'cluster' label is not set"),
			errors.Details{},
		)
	}
	return NewCriticalError(err)
}

func (node *resolutionNode) errorServiceDoesNotExist() error {
	var name string
	if node.context.Allocation != nil {
		name = node.context.Allocation.Service
	} else {
		name = "allocation block is empty"
	}
	err := errors.NewErrorWithDetails(
		fmt.Sprintf("Unable to find service definition: %s", name),
		errors.Details{},
	)
	return NewCriticalError(err)
}

func (node *resolutionNode) errorServiceOwnerDoesNotExist(service *lang.Service) error {
	err := errors.NewErrorWithDetails(
		fmt.Sprintf("Owner doesn't exist for service '%s': %s", service.Name, service.Owner),
		errors.Details{},
	)
	return NewCriticalError(err)
}

func (node *resolutionNode) errorServiceIsNotInSameNamespaceAsContract(service *lang.Service) error {
	err := errors.NewErrorWithDetails(
		fmt.Sprintf("Service '%s' is not in the same namespace as contract %s", object.GetKey(service), object.GetKey(node.contract)),
		errors.Details{},
	)
	return NewCriticalError(err)
}

func (node *resolutionNode) errorWhenTestingContext(context *lang.Context, cause error) error {
	err := errors.NewErrorWithDetails(
		fmt.Sprintf("Error while trying to match context '%s' for contract '%s': %s", context.Name, node.contract.Name, cause.Error()),
		errors.Details{
			"context": context,
			"cause":   cause,
		},
	)
	return NewCriticalError(err)
}

func (node *resolutionNode) errorWhenProcessingRule(rule *lang.Rule, cause error) error {
	err := errors.NewErrorWithDetails(
		fmt.Sprintf("Error while processing rule '%s' on contract '%s', context '%s', service '%s': %s", rule.Name, node.contract.Name, node.context.Name, node.service.Name, cause),
		errors.Details{
			"context": node.context,
			"rule":    rule,
			"labels":  node.labels.Labels,
			"cause":   cause,
		},
	)
	return NewCriticalError(err)
}

func (node *resolutionNode) errorWhenResolvingAllocationKeys(cause error) error {
	err := errors.NewErrorWithDetails(
		fmt.Sprintf("Error while resolving allocation keys for contract '%s', context '%s': %s", node.contract.Name, node.context.Name, cause.Error()),
		errors.Details{
			"cause": cause,
		},
	)
	return NewCriticalError(err)
}

func (node *resolutionNode) errorWhenDoingTopologicalSort(cause error) error {
	componentNames := []string{}
	for _, component := range node.service.Components {
		componentNames = append(componentNames, component.Name)
	}
	err := errors.NewErrorWithDetails(
		fmt.Sprintf("Failed to topologically sort components within a service '%s': %s", node.service.Name, cause.Error()),
		errors.Details{
			"cause":          cause,
			"componentNames": componentNames,
		},
	)
	return NewCriticalError(err)
}

func (node *resolutionNode) errorWhenProcessingCodeParams(cause error) error {
	err := errors.NewErrorWithDetails(
		fmt.Sprintf("Error when processing code params for service '%s', contract '%s', context '%s', component '%s': %s", node.service.Name, node.contract.Name, node.context.Name, node.component.Name, cause.Error()),
		errors.Details{
			"component":       node.component,
			"contextual_data": node.getContextualDataForCodeDiscoveryTemplate(),
			"cause":           cause,
		},
	)
	return NewCriticalError(err)
}

func (node *resolutionNode) errorWhenProcessingDiscoveryParams(cause error) error {
	err := errors.NewErrorWithDetails(
		fmt.Sprintf("Error when processing discovery params for service '%s', contract '%s', context '%s', component '%s': %s", node.service.Name, node.contract.Name, node.context.Name, node.component.Name, cause.Error()),
		errors.Details{
			"component":       node.component,
			"contextual_data": node.getContextualDataForCodeDiscoveryTemplate(),
			"cause":           cause,
		},
	)
	return NewCriticalError(err)
}

func (node *resolutionNode) errorServiceCycleDetected() error {
	err := errors.NewErrorWithDetails(
		fmt.Sprintf("Error when processing policy, service cycle detected: %s", node.path),
		errors.Details{
			"path": node.path,
		},
	)
	return NewCriticalError(err)
}

/*
	Event log - report debug/info/warning messages
*/

func (node *resolutionNode) logStartResolvingDependency() {
	userName := node.dependency.UserID
	if node.user != nil {
		userName = node.user.Name
	}
	if node.depth == 0 {
		// at the top of the tree, when we resolve a root-level dependency
		node.eventLog.WithFields(event.Fields{}).Infof("Resolving top-level dependency: '%s' -> '%s'", userName, node.dependency.Contract)
	} else {
		// recursively processing sub-dependencies
		node.eventLog.WithFields(event.Fields{}).Infof("Resolving dependency: '%s' -> '%s' (processing '%s', tree depth %d)", userName, node.dependency.Contract, node.contractName, node.depth)
	}

	node.logLabels(node.labels, "initial")
}

func (node *resolutionNode) logLabels(labelSet *lang.LabelSet, scope string) {
	secretCnt := 0
	if node.user != nil {
		secretCnt = len(node.resolver.externalData.SecretLoader.LoadSecretsByUserID(node.user.ID))
	}
	node.eventLog.WithFields(event.Fields{
		"labels": labelSet.Labels,
	}).Infof("Labels (%s): %s and %d secrets", scope, labelSet.Labels, secretCnt)
}

func (node *resolutionNode) logContractFound(contract *lang.Contract) {
	node.eventLog.WithFields(event.Fields{
		"contract": contract,
	}).Debugf("Contract found in policy: '%s'", contract.Name)
}

func (node *resolutionNode) logServiceFound(service *lang.Service) {
	node.eventLog.WithFields(event.Fields{
		"service": service,
	}).Debugf("Service found in policy: '%s'", service.Name)
}

func (node *resolutionNode) logStartMatchingContexts() {
	contextNames := []string{}
	for _, context := range node.contract.Contexts {
		contextNames = append(contextNames, context.Name)
	}
	node.eventLog.WithFields(event.Fields{}).Infof("Picking context within contract '%s'. Trying contexts: %s", node.contract.Name, contextNames)
}

func (node *resolutionNode) logContextMatched(contextMatched *lang.Context) {
	node.eventLog.WithFields(event.Fields{}).Infof("Found matching context within contract '%s': %s", node.contract.Name, contextMatched.Name)
}

func (node *resolutionNode) logContextNotMatched() {
	node.eventLog.WithFields(event.Fields{}).Warningf("Unable to find matching context within contract: '%s'", node.contract.Name)
}

func (node *resolutionNode) logTestedContextCriteria(context *lang.Context, matched bool) {
	node.eventLog.WithFields(event.Fields{
		"context": context,
	}).Debugf("Trying context '%s' within contract '%s'. Matched = %t", context.Name, node.contract.Name, matched)
}

func (node *resolutionNode) logRulesProcessingResult(policyNamespace *lang.PolicyNamespace, result *lang.RuleActionResult) {
	node.eventLog.WithFields(event.Fields{
		"result": result,
	}).Debugf("Rules processed within namespace '%s' for context '%s' within contract '%s'. Dependency allowed", policyNamespace.Name, node.context.Name, node.contract.Name)
}

func (node *resolutionNode) logTestedRuleMatch(rule *lang.Rule, match bool) {
	node.eventLog.WithFields(event.Fields{
		"rule":  rule,
		"match": match,
	}).Debugf("Testing if rule '%s' applies in context '%s' within contract '%s'. Result: %t", rule.Name, node.context.Name, node.contract.Name, match)
}

func (node *resolutionNode) logAllocationKeysSuccessfullyResolved(resolvedKeys []string) {
	if len(resolvedKeys) > 0 {
		node.eventLog.WithFields(event.Fields{
			"keys":         node.context.Allocation.Keys,
			"keysResolved": resolvedKeys,
		}).Infof("Allocation keys successfully resolved for context '%s' within contract '%s': %s", node.context.Name, node.contract.Name, resolvedKeys)
	}
}

func (node *resolutionNode) logResolvingDependencyOnComponent() {
	if node.component.Code != nil {
		node.eventLog.WithFields(event.Fields{}).Infof("Processing dependency on component with code: %s (%s)", node.component.Name, node.component.Code.Type)
	} else if node.component.Contract != "" {
		node.eventLog.WithFields(event.Fields{}).Infof("Processing dependency on another contract: %s", node.component.Contract)
	} else {
		node.eventLog.WithFields(event.Fields{}).Warningf("Skipping unknown component (not code and not contract): %s", node.component.Name)
	}
}

func (node *resolutionNode) logInstanceSuccessfullyResolved(cik *ComponentInstanceKey) {
	fields := event.Fields{
		"user":       node.user.Name,
		"dependency": node.dependency,
		"key":        cik,
	}
	if node.depth == 0 && cik.IsService() {
		// at the top of the tree, when we resolve a root-level dependency
		node.eventLog.WithFields(fields).Infof("Successfully resolved dependency '%s' -> '%s': %s", node.user.Name, node.dependency.Contract, cik.GetKey())
	} else if cik.IsService() {
		// resolved service instance
		node.eventLog.WithFields(fields).Infof("Successfully resolved service instance '%s' -> '%s': %s", node.user.Name, node.contract.Name, cik.GetKey())
	} else {
		// resolved component instance
		node.eventLog.WithFields(fields).Infof("Successfully resolved component instance '%s' -> '%s' (component '%s'): %s", node.user.Name, node.contract.Name, node.component.Name, cik.GetKey())
	}
}

func (node *resolutionNode) logCannotResolveInstance() {
	if node.service == nil {
		node.eventLog.WithFields(event.Fields{}).Warningf("Cannot resolve instance: contract '%s'", node.contractName)
	} else if node.component == nil {
		node.eventLog.WithFields(event.Fields{}).Warningf("Cannot resolve instance: contract '%s', service '%s'", node.contractName, node.service.Name)
	} else {
		node.eventLog.WithFields(event.Fields{}).Warningf("Cannot resolve instance: contract '%s', service '%s', component '%s'", node.contractName, node.service.Name, node.component.Name)
	}
}

func (resolver *PolicyResolver) logComponentCodeParams(instance *ComponentInstance) {
	serviceObj, err := resolver.policy.GetObject(lang.ServiceObject.Kind, instance.Metadata.Key.ServiceName, instance.Metadata.Key.Namespace)
	if err != nil {
		panic(fmt.Sprintf("Fatal error while getting service '%s/%s' from the policy: %s", instance.Metadata.Key.ServiceName, instance.Metadata.Key.Namespace, err))
	}
	code := serviceObj.(*lang.Service).GetComponentsMap()[instance.Metadata.Key.ComponentName].Code
	if code != nil {
		paramsTemplate := code.Params
		params := instance.CalculatedCodeParams
		diff := strings.TrimSpace(paramsTemplate.Diff(params))
		if len(diff) > 0 {
			resolver.eventLog.WithFields(event.Fields{
				"params": diff,
			}).Debugf("Calculated code params for component '%s'", instance.Metadata.Key.GetKey())
		}
	}
}

func (resolver *PolicyResolver) logComponentDiscoveryParams(instance *ComponentInstance) {
	serviceObj, err := resolver.policy.GetObject(lang.ServiceObject.Kind, instance.Metadata.Key.ServiceName, instance.Metadata.Key.Namespace)
	if err != nil {
		panic(fmt.Sprintf("Fatal error while getting service '%s/%s' from the policy: %s", instance.Metadata.Key.ServiceName, instance.Metadata.Key.Namespace, err))
	}
	paramsTemplate := serviceObj.(*lang.Service).GetComponentsMap()[instance.Metadata.Key.ComponentName].Discovery
	params := instance.CalculatedDiscovery
	diff := strings.TrimSpace(paramsTemplate.Diff(params))
	if len(diff) > 0 {
		resolver.eventLog.WithFields(event.Fields{
			"params": diff,
		}).Debugf("Calculated discovery params for component '%s'", instance.Metadata.Key.GetKey())
	}
}
