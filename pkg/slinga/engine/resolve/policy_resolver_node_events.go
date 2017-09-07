package resolve

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/slinga/errors"
	. "github.com/Aptomi/aptomi/pkg/slinga/eventlog"
	. "github.com/Aptomi/aptomi/pkg/slinga/language"
	"strings"
)

/*
	Non-critical errors. If one of them occurs, the corresponding dependency will not be fulfilled
	and engine will move on to processing other dependencies
*/

func (node *resolutionNode) errorUserDoesNotExist() error {
	return errors.NewErrorWithDetails(
		fmt.Sprintf("Dependency refers to non-existing user: "+node.dependency.UserID),
		errors.Details{},
	)
}

/*
	Critical errors. If one of them occurs, engine will report an error and fail policy processing
	all together
*/

func (node *resolutionNode) errorServiceDoesNotExist() error {
	err := errors.NewErrorWithDetails(
		fmt.Sprintf("Unable to find service definition: %s", node.serviceName),
		errors.Details{},
	)
	return NewCriticalError(err)
}

func (node *resolutionNode) errorServiceOwnerDoesNotExist() error {
	err := errors.NewErrorWithDetails(
		fmt.Sprintf("Unable to find service owner for service '%s': %s", node.serviceName, node.resolver.policy.Services[node.serviceName].Owner),
		errors.Details{},
	)
	return NewCriticalError(err)
}

func (node *resolutionNode) errorWhenTestingContext(context *Context, cause error) error {
	err := errors.NewErrorWithDetails(
		fmt.Sprintf("Error while trying to match context '%s' for service '%s': %s", context.Name, node.service.Name, cause.Error()),
		errors.Details{
			"context": context,
			"cause":   cause,
		},
	)
	return NewCriticalError(err)
}

func (node *resolutionNode) errorGettingClusterForGlobalRules(context *Context, labelSet *LabelSet, cause error) error {
	err := errors.NewErrorWithDetails(
		fmt.Sprintf("Can't evaluate global rules for service '%s', context '%s' due to cluster error: %s", node.service.Name, context.Name, cause.Error()),
		errors.Details{
			"context": context,
			"labels":  labelSet,
			"cause":   cause,
		},
	)
	return NewCriticalError(err)
}

func (node *resolutionNode) errorWhenTestingGlobalRule(context *Context, rule *Rule, labelSet *LabelSet, cause error) error {
	err := errors.NewErrorWithDetails(
		fmt.Sprintf("Error while testing global rule '%s' on service '%s', context '%s': %s", rule.Name, node.service.Name, context.Name, cause),
		errors.Details{
			"context": context,
			"rule":    rule,
			"labels":  labelSet,
			"cause":   cause,
		},
	)
	return NewCriticalError(err)
}

func (node *resolutionNode) errorWhenResolvingAllocationKeys(cause error) error {
	err := errors.NewErrorWithDetails(
		fmt.Sprintf("Error while resolving allocation keys for service '%s', context '%s': %s", node.service.Name, node.context.Name, cause.Error()),
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
		fmt.Sprintf("Error when processing code params for service '%s', context '%s', component '%s': %s", node.service.Name, node.context.Name, node.component.Name, cause.Error()),
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
		fmt.Sprintf("Error when processing discovery params for service '%s', context '%s', component '%s': %s", node.service.Name, node.context.Name, node.component.Name, cause.Error()),
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
		fmt.Sprintf("Error when processing policy, cycle detected: %s", node.path),
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
		node.eventLog.WithFields(Fields{}).Infof("Resolving top-level dependency: '%s' -> '%s'", userName, node.dependency.Service)
	} else {
		// recursively processing sub-dependencies
		node.eventLog.WithFields(Fields{}).Infof("Resolving dependency: '%s' -> '%s' (processing '%s', tree depth %d)", userName, node.dependency.Service, node.serviceName, node.depth)
	}

	node.logLabels(node.labels, "initial")
}

func (node *resolutionNode) logLabels(labelSet *LabelSet, scope string) {
	secretCnt := 0
	if node.user != nil {
		secretCnt = len(node.resolver.externalData.SecretLoader.LoadSecretsByUserID(node.user.ID))
	}
	node.eventLog.WithFields(Fields{
		"labels": labelSet.Labels,
	}).Infof("Labels (%s): %s and %d secrets", scope, labelSet.Labels, secretCnt)
}

func (node *resolutionNode) logServiceFound(service *Service) {
	node.eventLog.WithFields(Fields{
		"service": service,
	}).Debugf("Service found in policy: '%s'", service.Name)
}

func (node *resolutionNode) logStartMatchingContexts() {
	contextNames := []string{}
	for _, context := range node.resolver.policy.Contexts {
		contextNames = append(contextNames, context.Name)
	}
	node.eventLog.WithFields(Fields{}).Infof("Resolving context for service '%s'. Trying contexts: %s", node.service.Name, contextNames)
}

func (node *resolutionNode) logContextMatched(contextMatched *Context) {
	node.eventLog.WithFields(Fields{}).Infof("Found matching context for service '%s': %s", node.service.Name, contextMatched.Name)
}

func (node *resolutionNode) logContextNotMatched() {
	node.eventLog.WithFields(Fields{}).Warningf("Unable to find matching context for service: '%s'", node.service.Name)
}

func (node *resolutionNode) logTestedContextCriteria(context *Context, matched bool) {
	node.eventLog.WithFields(Fields{
		"context": context,
	}).Debugf("Trying context '%s' for service '%s'. Matched = %t", context.Name, node.service.Name, matched)
}

func (node *resolutionNode) logTestedGlobalRuleViolations(context *Context, labelSet *LabelSet, noViolations bool) {
	fields := Fields{
		"context": context,
	}
	if noViolations {
		node.eventLog.WithFields(fields).Debugf("No global rule violations found for service: '%s', context: %s", node.service.Name, context.Name)
	} else {
		node.eventLog.WithFields(fields).Debugf("Detected global rule violation for service: '%s', context: %s", node.service.Name, context.Name)
	}
}

func (node *resolutionNode) logTestedGlobalRuleMatch(context *Context, rule *Rule, labelSet *LabelSet, match bool) {
	node.eventLog.WithFields(Fields{
		"context": context,
		"rule":    rule,
		"labels":  labelSet,
		"match":   match,
	}).Debugf("Testing if global rule '%s' applies to service '%s', context '%s'. Result: %t", rule.Name, node.service.Name, context.Name, match)
}

func (node *resolutionNode) logAllocationKeysSuccessfullyResolved(resolvedKeys []string) {
	if len(resolvedKeys) > 0 {
		node.eventLog.WithFields(Fields{
			"keys":         node.context.Allocation.Keys,
			"keysResolved": resolvedKeys,
		}).Infof("Allocation keys successfully resolved for service '%s', context '%s': %s", node.service.Name, node.context.Name, resolvedKeys)
	}
}

func (node *resolutionNode) logResolvingDependencyOnComponent() {
	if node.component.Code != nil {
		node.eventLog.WithFields(Fields{}).Infof("Processing dependency on component with code: %s (%s)", node.component.Name, node.component.Code.Type)
	} else if node.component.Service != "" {
		node.eventLog.WithFields(Fields{}).Infof("Processing dependency on another service: %s", node.component.Service)
	} else {
		node.eventLog.WithFields(Fields{}).Warningf("Skipping unknown component (not code and not service): %s", node.component.Name)
	}
}

func (node *resolutionNode) logInstanceSuccessfullyResolved(cik *ComponentInstanceKey) {
	fields := Fields{
		"user":       node.user.Name,
		"dependency": node.dependency,
		"key":        cik,
	}
	if node.depth == 0 && cik.IsService() {
		// at the top of the tree, when we resolve a root-level dependency
		node.eventLog.WithFields(fields).Infof("Successfully resolved dependency '%s' -> '%s': %s", node.user.Name, node.dependency.Service, cik.GetKey())
	} else if cik.IsService() {
		// resolved service instance
		node.eventLog.WithFields(fields).Infof("Successfully resolved service instance '%s' -> '%s': %s", node.user.Name, node.service.Name, cik.GetKey())
	} else {
		// resolved component instance
		node.eventLog.WithFields(fields).Infof("Successfully resolved component instance '%s' -> '%s' (component '%s'): %s", node.user.Name, node.service.Name, node.component.Name, cik.GetKey())
	}
}

func (node *resolutionNode) logCannotResolveInstance() {
	if node.component == nil {
		node.eventLog.WithFields(Fields{}).Warningf("Cannot resolve service instance: service '%s'", node.serviceName)
	} else {
		node.eventLog.WithFields(Fields{}).Warningf("Cannot resolve component instance: service '%s', component '%s'", node.serviceName, node.component.Name)
	}
}

func (resolver *PolicyResolver) logComponentCodeParams(instance *ComponentInstance) {
	code := resolver.policy.Services[instance.Metadata.Key.ServiceName].GetComponentsMap()[instance.Metadata.Key.ComponentName].Code
	if code != nil {
		paramsTemplate := code.Params
		params := instance.CalculatedCodeParams
		diff := strings.TrimSpace(paramsTemplate.Diff(params))
		if len(diff) > 0 {
			resolver.eventLog.WithFields(Fields{
				"params": diff,
			}).Debugf("Calculated code params for component '%s'", instance.Metadata.Key.GetKey())
		}
	}
}

func (resolver *PolicyResolver) logComponentDiscoveryParams(instance *ComponentInstance) {
	paramsTemplate := resolver.policy.Services[instance.Metadata.Key.ServiceName].GetComponentsMap()[instance.Metadata.Key.ComponentName].Discovery
	params := instance.CalculatedDiscovery
	diff := strings.TrimSpace(paramsTemplate.Diff(params))
	if len(diff) > 0 {
		resolver.eventLog.WithFields(Fields{
			"params": diff,
		}).Debugf("Calculated discovery params for component '%s'", instance.Metadata.Key.GetKey())
	}
}
