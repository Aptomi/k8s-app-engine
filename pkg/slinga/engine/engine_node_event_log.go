package engine

import (
	. "github.com/Aptomi/aptomi/pkg/slinga/eventlog"
	. "github.com/Aptomi/aptomi/pkg/slinga/language"
)

func (node *resolutionNode) logStartResolvingDependency() {
	fields := Fields{
		"dependency": node.dependency,
		"user":       node.user,
		"dependsOn":  node.serviceName,
	}
	if node.depth == 0 {
		// at the top of the tree, when we resolve a root-level dependency
		node.eventLog.WithFields(fields).Infof("Resolving dependency: '%s' -> '%s'", node.user.Name, node.dependency.Service)
	} else {
		// recursively processing sub-dependencies
		node.eventLog.WithFields(fields).Infof("Continuing to resolve dependency: '%s' -> '%s' (processing '%s')", node.user.Name, node.dependency.Service, node.serviceName)
	}

	node.logLabels(node.labels)
}

func (node *resolutionNode) logLabels(labelSet LabelSet) {
	node.eventLog.WithFields(Fields{
		"labels": labelSet.Labels,
	}).Infof("Labels calculated: %d", len(labelSet.Labels))
}

func (node *resolutionNode) logServiceFound(service *Service) {
	node.eventLog.WithFields(Fields{
		"service": service,
	}).Debugf("Service found in policy: '%s'", service.Name)
}

func (node *resolutionNode) logServiceNotFoundError(serviceName string) {
	node.eventLog.WithFields(Fields{
		"serviceName": serviceName,
	}).Errorf("Service not found in policy: '%s'", serviceName)
}

func (node *resolutionNode) logCannotResolveInstance() {
	if node.component == nil {
		node.eventLog.WithFields(Fields{
			"serviceName": node.serviceName,
			"context":     node.context,
		}).Warningf("Cannot resolve service instance: service '%s', context '%s'", node.serviceName, getContextNameUnsafe(node.context))
	} else {
		node.eventLog.WithFields(Fields{
			"serviceName": node.serviceName,
			"context":     node.context,
			"component":   node.component,
		}).Warningf("Cannot resolve component instance: service '%s', context '%s', component '%s'", node.serviceName, getContextNameUnsafe(node.context), getComponentNameUnsafe(node.component))
	}

	// There may be a situation when service key has not been resolved yet. If so, we should create a fake one to attach logs to
	if node.serviceKey == nil {
		// Create service key
		node.serviceKey = node.createComponentKey(nil)

		// Once instance is figured out, make sure to attach rule logs to that instance
		node.eventLog.AttachToInstance(node.serviceKey.GetKey())
	}
}

func (node *resolutionNode) logStartMatchingContexts() {
	node.eventLog.WithFields(Fields{
		"dependency": node.dependency,
		"user":       node.user,
	}).Infof("Resolving context for service: '%s'. Contexts defined: %d", node.service.Name, len(node.state.Policy.Contexts))
}

func (node *resolutionNode) logContextMatched(contextMatched *Context) {
	node.eventLog.WithFields(Fields{
		"service": node.service.Name,
		"context": contextMatched.Name,
		"user":    node.user.Name,
	}).Infof("Found matching context for service: '%s'. Context: '%s'", node.service.Name, contextMatched.Name)
}

func (node *resolutionNode) logContextNotMatched() {
	node.eventLog.WithFields(Fields{
		"service": node.service.Name,
		"user":    node.user.Name,
	}).Warningf("Unable to find matching context for service: '%s'", node.service.Name)
}

func (node *resolutionNode) logTestedContextCriteria(context *Context, matched bool) {
	node.eventLog.WithFields(Fields{
		"service": node.service,
		"context": context,
		"matched": matched,
	}).Debugf("Testing context match for service: '%s'. Context: '%s'. Result: %t", node.service.Name, context.Name, matched)
}

func (node *resolutionNode) logTestedContextCriteriaError(context *Context, err error) {
	node.eventLog.WithFields(Fields{
		"service": node.service,
		"context": context,
		"err":     err,
	}).Errorf("Error while testing context match for service: '%s'. Context: '%s'. Error: %s", node.service.Name, context.Name, err.Error())
}

func (node *resolutionNode) logTestedGlobalRuleViolations(context *Context, labelSet LabelSet, noViolations bool) {
	fields := Fields{
		"service": node.service,
		"context": context,
	}
	if noViolations {
		node.eventLog.WithFields(fields).Debugf("No global rule violations found for service: '%s', context: %s", node.service.Name, context.Name)
	} else {
		node.eventLog.WithFields(fields).Debugf("Detected global rule violation for service: '%s', context: %s", node.service.Name, context.Name)
	}
}

func (node *resolutionNode) logTestedGlobalRuleViolationsError(context *Context, labelSet LabelSet, err error) {
	node.eventLog.WithFields(Fields{
		"service": node.service,
		"context": context,
	}).Errorf("Error while checking global rule violations found for service: '%s', context '%s'. Error: %s", node.service.Name, context.Name, string(err.Error()))
}

func (node *resolutionNode) logTestedGlobalRuleMatch(context *Context, rule *Rule, labelSet LabelSet, match bool) {
	node.eventLog.WithFields(Fields{
		"service": node.service,
		"context": context,
		"rule":    rule,
		"labels":  labelSet,
		"match":   match,
	}).Debugf("Testing if global rule '%s' applies to service '%s', context '%s'. Result: ", rule.Name, node.service.Name, context.Name, match)
}

func (node *resolutionNode) logTestedGlobalRuleMatchError(context *Context, rule *Rule, labelSet LabelSet, err error) {
	node.eventLog.WithFields(Fields{
		"service": node.service,
		"context": context,
		"rule":    rule,
		"labels":  labelSet,
	}).Errorf("Error while testing global rule '%s' on service '%s', context '%s'. Error: %s", rule.Name, node.service.Name, context.Name, err)
}

func (node *resolutionNode) logAllocationKeysSuccessfullyResolved(resolvedKeys []string) {
	node.eventLog.WithFields(Fields{
		"service": node.service.Name,
		"context": node.context.Name,
		"keys":    node.context.Allocation.Keys,
	}).Debugf("Allocation keys successfully resolved for service '%s', context '%s': %v", node.service.Name, node.context.Name, resolvedKeys)
}

func (node *resolutionNode) logResolvingAllocationKeysError(err error) {
	node.eventLog.WithFields(Fields{
		"service": node.service.Name,
		"context": node.context.Name,
		"keys":    node.context.Allocation.Keys,
	}).Errorf("Error while resolving allocation keys for service '%s', context '%s'. Error: %s", node.service.Name, node.context.Name, err)
}

func (node *resolutionNode) logFailedTopologicalSort(err error) {
	node.eventLog.WithFields(Fields{
		"service":    node.service.Name,
		"components": node.service.Components,
	}).Errorf("Failed to topologically sort components within a service: '%s'. Error: %s", node.service.Name, err.Error())
}

func (node *resolutionNode) logResolvingDependencyOnComponent() {
	if node.component.Code != nil {
		node.eventLog.WithFields(Fields{
			"service":   node.service.Name,
			"component": node.component.Name,
			"context":   node.context.Name,
		}).Infof("Processing dependency on component/code: %s (%s)", node.component.Name, node.component.Code.Type)
	} else if node.component.Service != "" {
		node.eventLog.WithFields(Fields{
			"service":   node.service.Name,
			"component": node.component.Name,
			"context":   node.context.Name,
			"dependsOn": node.component.Service,
		}).Infof("Processing dependency on another service: %s", node.component.Service)
	} else {
		node.eventLog.WithFields(Fields{
			"service":   node.service.Name,
			"component": node.component.Name,
		}).Errorf("Invalid component (not code and not service")
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
		node.eventLog.WithFields(fields).Infof("Successfully resolved component instance '%s' -> '%s' ('%s'): %s", node.user.Name, node.service.Name, node.component.Name, cik.GetKey())
	}
}
