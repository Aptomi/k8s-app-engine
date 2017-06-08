package slinga

import (
	"errors"
	log "github.com/Sirupsen/logrus"
)

/*
	Core engine for Slinga processing and evaluation
*/

type node struct {
	// TODO: complete the struct
	User *User
}

// ResolveUsage evaluates all recorded Dependencies ("<user> needs <service> with <labels>") and calculates allocations
func (usage *ServiceUsageState) ResolveUsage() error {

	// Run every declared dependency via policy and resolve it
	for serviceName, dependencies := range usage.Dependencies.Dependencies {
		for _, d := range dependencies {
			user := usage.users.Users[d.UserID]

			// take user labels
			labels := user.getLabelSet()

			// combine them with dependency labels
			labels = labels.addLabels(d.getLabelSet())

			// see if it needs to be traced (addl debug output on console)
			tracing.setEnable(d.Trace)

			// resolve usage via applying policy
			resKey, err := usage.resolveWithLabels(user, serviceName, labels, usage.ResolvedUsage.DiscoveryTree, 0)

			// TODO: if a dependency cannot be fulfilled, we need to handle it correctly.
			// i.e. usages should be recorded in different context (not in usage.DiscoveryTree and not even in usage) and not applied

			// disable tracing
			tracing.setEnable(false)

			// see if there is an error
			if err != nil {
				return err
			}

			// record high-level service resolution
			d.ResolvesTo = resKey
		}
	}
	return nil
}

// Evaluate "<user> needs <service>" statement
func (usage *ServiceUsageState) resolveWithLabels(user User, serviceName string, labels LabelSet, discoveryTreeNode NestedParameterMap, depth int) (string, error) {

	// Resolving allocations for service
	debug.WithFields(log.Fields{
		"service": serviceName,
		"user":    user.Name,
		"labels":  labels,
	}).Info("Resolving allocations for service")

	tracing.Printf(depth, "[Dependency]")
	tracing.Printf(depth, "User: %s (ID = %s)", user.Name, user.ID)
	tracing.Printf(depth+1, "Labels: %s", labels)
	tracing.Printf(depth, "Service: %s", serviceName)

	// Policy
	policy := usage.Policy

	// Locate the service
	service, err := policy.getService(serviceName)
	if err != nil {
		tracing.Printf(depth+1, "Error while trying to look up service %s (%v)", serviceName, err)
		return "", err
	}

	// Process service and transform labels
	labels = labels.applyTransform(service.Labels)
	tracing.Printf(depth+1, "New labels = %s", labels)

	// Match the context
	context, err := policy.getMatchedContext(*service, user, labels, depth)
	if err != nil {
		tracing.Printf(depth+1, "Error while matching context for service %s (%v)", serviceName, err)
		return "", err
	}
	// If no matching context is found, let's just exit
	if context == nil {
		tracing.Printf(depth+1, "No context matched for service %s", serviceName)
		return "", nil
	}

	// Process context and transform labels
	labels = labels.applyTransform(context.Labels)
	tracing.Printf(depth, "Context: %s", context.Name)
	tracing.Printf(depth+1, "New labels = %s", labels)

	// Match the allocation
	allocation, err := policy.getMatchedAllocation(*service, user, *context, labels, depth)
	if err != nil {
		tracing.Printf(depth, "Error while matching allocation for service %s, context %s (%v)", serviceName, context.Name, err)
		return "", err
	}
	// If no matching allocation is found, let's just exit
	if allocation == nil {
		tracing.Printf(depth, "No allocation matched for service %s, context %s", serviceName, context.Name)
		return "", nil
	}

	// Process allocation and transform labels
	labels = labels.applyTransform(allocation.Labels)
	tracing.Printf(depth, "Allocation: %s", allocation.NameResolved)
	tracing.Printf(depth+1, "New labels = %s", labels)

	// Now, sort all components in topological order
	componentsOrdered, err := service.getComponentsSortedTopologically()
	if err != nil {
		return "", err
	}

	// Iterate over all service components and resolve them recursively
	// Note that discovery variables can refer to other variables announced by dependents in the discovery tree
	for _, component := range componentsOrdered {
		// Create key
		componentKey := usage.createServiceUsageKey(service, context, allocation, component)

		// Calculate and store labels
		componentLabels := labels.applyTransform(component.Labels)
		usage.storeLabels(componentKey, componentLabels)

		// Calculate and store discovery params
		componentDiscoveryParams, err := component.processTemplateParams(component.Discovery, componentKey, componentLabels, user, discoveryTreeNode, "discovery", depth)
		if err != nil {
			return "", err
		}
		usage.storeDiscoveryParams(componentKey, componentDiscoveryParams)

		// Create new map with resolution keys for component
		discoveryTreeNode[component.Name] = NestedParameterMap{}

		if component.Code != nil {
			// Evaluate code params
			debug.WithFields(log.Fields{
				"service":    service.Name,
				"component":  component.Name,
				"context":    context.Name,
				"allocation": allocation.NameResolved,
			}).Info("Processing dependency on code execution")

			// Populate discovery tree (allow this component to announce its discovery properties in the discovery tree)
			discoveryTreeNode.getNestedMap(component.Name)["instance"] = EscapeName(componentKey)
			for k, v := range componentDiscoveryParams {
				discoveryTreeNode.getNestedMap(component.Name)[k] = v
			}

			componentCodeParams, err := component.processTemplateParams(component.Code.Params, componentKey, componentLabels, user, discoveryTreeNode, "code", depth)
			if err != nil {
				return "", err
			}
			usage.storeCodeParams(componentKey, componentCodeParams)
		} else if component.Service != "" {
			debug.WithFields(log.Fields{
				"service":          service.Name,
				"component":        component.Name,
				"context":          context.Name,
				"allocation":       allocation.NameResolved,
				"dependsOnService": component.Service,
			}).Info("Processing dependency on another service")

			tracing.Println()

			// resolve dependency recursively
			resolvedKey, err := usage.resolveWithLabels(user, component.Service, componentLabels, discoveryTreeNode.getNestedMap(component.Name), depth+1)
			if err != nil {
				return "", err
			}

			// if a dependency has not been matched
			if len(resolvedKey) <= 0 {
				debug.WithFields(log.Fields{
					"service":          service.Name,
					"component":        component.Name,
					"context":          context.Name,
					"allocation":       allocation.NameResolved,
					"dependsOnService": component.Service,
				}).Info("Cannot fulfill dependency on another service")
				return "", nil
			}
		} else {
			debug.WithFields(log.Fields{
				"service":   service.Name,
				"component": component.Name,
			}).Fatal("Invalid component (not code and not service")
		}

		// Record usage of a given component
		usage.recordUsage(componentKey, user)
	}

	// Record usage of a given service
	serviceKey := usage.createServiceUsageKey(service, context, allocation, nil)
	usage.recordUsage(serviceKey, user)

	return context.Name + "#" + allocation.NameResolved, nil
}

// Topologically sort components and return true if there is a cycle detected
func (service *Service) dfsComponentSort(u *ServiceComponent, colors map[string]int) bool {
	colors[u.Name] = 1

	for _, vName := range u.Dependencies {
		v, exists := service.getComponentsMap()[vName]
		if !exists {
			debug.WithFields(log.Fields{
				"service":   service.Name,
				"component": vName,
			}).Fatal("Service dependency points to non-existing component")
		}
		if vColor, ok := colors[v.Name]; !ok {
			// not visited yet -> visit and exit if a cycle was found
			if service.dfsComponentSort(v, colors) {
				return true
			}
		} else if vColor == 1 {
			return true
		}
	}

	service.componentsOrdered = append(service.componentsOrdered, u)
	colors[u.Name] = 2
	return false
}

// Sorts all components in a topological way
func (service *Service) getComponentsSortedTopologically() ([]*ServiceComponent, error) {
	if service.componentsOrdered == nil {
		// Initiate colors
		colors := make(map[string]int)

		// Dfs
		var cycle = false
		for _, c := range service.Components {
			if _, ok := colors[c.Name]; !ok {
				if service.dfsComponentSort(c, colors) {
					cycle = true
					break
				}
			}
		}

		if cycle {
			return nil, errors.New("Component cycle detected in service " + service.Name)
		}
	}

	return service.componentsOrdered, nil
}

// Helper to get a service
func (policy *Policy) getService(serviceName string) (*Service, error) {
	// Locate the service
	service := policy.Services[serviceName]
	if service == nil {
		return nil, errors.New("Service " + serviceName + " not found")
	}
	return service, nil
}

// Helper to get a matched context
func (policy *Policy) getMatchedContext(service Service, user User, labels LabelSet, depth int) (*Context, error) {
	// Locate the list of contexts for service
	contexts, ok := policy.Contexts[service.Name]
	if !ok || len(contexts) <= 0 {
		tracing.Printf(depth+1, "No contexts found")
		return nil, errors.New("No contexts found for " + service.Name)
	}

	// See which context matches
	var contextMatched *Context
	for _, c := range contexts {
		m := c.matches(labels)
		tracing.Printf(depth+1, "[%t] Testing context '%s': (criteria = %+v)", m, c.Name, c.Criteria)
		if m {
			contextMatched = c
			break
		}
	}

	if contextMatched != nil {
		debug.WithFields(log.Fields{
			"service": service.Name,
			"context": contextMatched.Name,
			"user":    user.Name,
		}).Info("Matched context")
	} else {
		debug.WithFields(log.Fields{
			"service": service.Name,
			"user":    user.Name,
		}).Info("No context matched")
	}
	return contextMatched, nil
}

// Helper to get a matched allocation
func (policy *Policy) getMatchedAllocation(service Service, user User, context Context, labels LabelSet, depth int) (*Allocation, error) {
	if len(context.Allocations) <= 0 {
		tracing.Printf(depth+1, "No allocations found")
		return nil, errors.New("No allocations found for " + service.Name)
	}

	// See which allocation matches
	var allocationMatched *Allocation
	for _, a := range context.Allocations {
		m := a.matches(labels)
		tracing.Printf(depth+1, "[%t] Testing allocation '%s': (criteria = %+v)", m, a.Name, a.Criteria)
		if m {
			allocationMatched = a
			break
		}
	}

	// Check errors and resolve allocation name (it can be dynamic, depending on user labels)
	if allocationMatched != nil {
		err := allocationMatched.resolveName(user, labels)
		if err != nil {
			debug.WithFields(log.Fields{
				"service":    service.Name,
				"context":    context.Name,
				"allocation": allocationMatched.Name,
				"user":       user.Name,
				"error":      err,
			}).Fatal("Cannot resolve name for an allocation")
		}
		debug.WithFields(log.Fields{
			"service":            service.Name,
			"context":            context.Name,
			"allocation":         allocationMatched.Name,
			"allocationResolved": allocationMatched.NameResolved,
			"user":               user.Name,
		}).Info("Matched allocation")
	} else {
		debug.WithFields(log.Fields{
			"service": service.Name,
			"context": context.Name,
			"user":    user.Name,
		}).Info("No allocation matched")
	}

	return allocationMatched, nil
}

type templateData struct {
	Labels    map[string]string
	User      User
	Discovery NestedParameterMap
}
