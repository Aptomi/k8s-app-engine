package slinga

import (
	"errors"
	"github.com/golang/glog"
)

/*
	Core engine for Slinga processing and evaluation
*/

// ResolveUsage evaluates all recorded Dependencies ("<user> needs <service> with <labels>") and calculates allocations
func (usage *ServiceUsageState) ResolveUsage(users *GlobalUsers) error {
	for serviceName, dependencies := range usage.Dependencies.Dependencies {
		for _, d := range dependencies {
			user := users.Users[d.UserID]

			// take user labels
			labels := user.getLabelSet()

			// combine them with dependency labels
			labels = labels.addLabels(d.getLabelSet())

			// resolve usage via applying policy
			resKey, err := usage.resolveWithLabels(user, serviceName, labels, d.Trace, 0)
			if err != nil {
				return err
			}
			d.ResolvesTo = resKey
		}
	}
	return nil
}

// Evaluate "<user> needs <service>" statement
func (usage *ServiceUsageState) resolveWithLabels(user User, serviceName string, labels LabelSet, trace bool, depth int) (string, error) {

	// Resolving allocations for service
	glog.Infof("Resolving allocations for service %s (user = %s, labels = %s)", serviceName, user.Name, labels)

	usage.tracing.do(trace).log(depth,"[Dependency]")
	usage.tracing.do(trace).log(depth,"User: %s (ID = %s)", user.Name, user.ID)
	usage.tracing.do(trace).log(depth + 1,"Labels: %s", labels)
	usage.tracing.do(trace).log(depth,"Service: %s", serviceName)

	// Policy
	policy := usage.Policy

	// Locate the service
	service, err := policy.getService(serviceName)
	if err != nil {
		usage.tracing.do(trace).log(depth,"Error while trying to look up service %s (%v)", serviceName, err)
		return "", err
	}

	// Process service and transform labels
	labels = labels.applyTransform(service.Labels)
	usage.tracing.do(trace).log(depth + 1,"New labels = %s", labels)

	// Match the context
	context, err := policy.getMatchedContext(*service, user, labels, depth, usage.tracing.do(trace))
	if err != nil {
		usage.tracing.do(trace).log(depth,"Error while matching context for service %s (%v)", serviceName, err)
		return "", err
	}
	// If no matching context is found, let's just exit
	if context == nil {
		usage.tracing.do(trace).log(depth,"No context matched for service %s", serviceName)
		return "", nil
	}

	// Process context and transform labels
	labels = labels.applyTransform(context.Labels)
	usage.tracing.do(trace).log(depth,"Context: %s", context.Name)
	usage.tracing.do(trace).log(depth + 1,"New labels = %s", labels)

	// Match the allocation
	allocation, err := policy.getMatchedAllocation(*service, user, *context, labels, depth, usage.tracing.do(trace))
	if err != nil {
		usage.tracing.do(trace).log(depth,"Error while matching allocation for service %s, context %s (%v)", serviceName, context.Name, err)
		return "", err
	}
	// If no matching allocation is found, let's just exit
	if allocation == nil {
		usage.tracing.do(trace).log(depth,"No allocation matched for service %s, context %s", serviceName, context.Name)
		return "", nil
	}

	// Process allocation and transform labels
	labels = labels.applyTransform(allocation.Labels)
	usage.tracing.do(trace).log(depth,"Allocation: %s", allocation.NameResolved)
	usage.tracing.do(trace).log(depth + 1,"New labels = %s", labels)

	// Now, sort all components in topological order
	componentsOrdered, err := service.getComponentsSortedTopologically()
	if err != nil {
		return "", err
	}

	componentDepMap := make(map[string]string)

	// Resolve every component
	for _, component := range componentsOrdered {
		// Process component and transform labels
		componentLabels := labels.applyTransform(component.Labels)

		// Is it a code?
		if component.Code != nil {
			// glog.Infof("Processing dependency on code execution: %s (in %s)", component.Name, service.Name)
		} else if component.Service != "" {
			glog.Infof("Processing dependency on another service: %s -> %s (in %s)", component.Name, component.Service, service.Name)
			usage.tracing.do(trace).newline()

			_, err := usage.resolveWithLabels(user, component.Service, componentLabels, trace, depth + 1)
			if err != nil {
				return "", err
			}
		} else {
			glog.Fatalf("Invalid component: %s (in %s)", component.Name, service.Name)
		}

		// Record usage of a given component
		componentKey := usage.recordUsage(user, service, context, allocation, component, componentLabels)
		componentDepMap[component.Name] = componentKey
	}

	// Record usage of a given service
	serviceKey := usage.recordUsage(user, service, context, allocation, nil, labels)

	usage.ComponentInstanceMap[serviceKey] = componentDepMap

	return context.Name + "#" + allocation.NameResolved, nil
}

// Topologically sort components and return true if there is a cycle detected
func (service *Service) dfsComponentSort(u *ServiceComponent, colors map[string]int) bool {
	colors[u.Name] = 1

	for _, vName := range u.Dependencies {
		v, exists := service.getComponentsMap()[vName]
		if !exists {
			glog.Fatalf("Invalid dependency in service %s: %s", service.Name, vName)
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
func (policy *Policy) getMatchedContext(service Service, user User, labels LabelSet, depth int, tracing *ServiceUsageTracing) (*Context, error) {
	// Locate the list of contexts for service
	contexts, ok := policy.Contexts[service.Name]
	if !ok || len(contexts) <= 0 {
		tracing.log(depth + 1, "No contexts found")
		return nil, errors.New("No contexts found for " + service.Name)
	}

	// See which context matches
	var contextMatched *Context
	for _, c := range contexts {
		m := c.matches(labels)
		tracing.log(depth + 1, "[%t] Testing context '%s': %s", m, c.Name, c.Criteria)
		if m {
			contextMatched = c
			break
		}
	}

	if contextMatched != nil {
		glog.Infof("Matched context: '%s' (service = %s, user = %s)", contextMatched.Name, service.Name, user.Name)
	} else {
		glog.Infof("No context matched (service = %s, user = %s)", service.Name, user.Name)
	}
	return contextMatched, nil
}

// Helper to get a matched allocation
func (policy *Policy) getMatchedAllocation(service Service, user User, context Context, labels LabelSet, depth int, tracing *ServiceUsageTracing) (*Allocation, error) {
	if len(context.Allocations) <= 0 {
		tracing.log(depth + 1, "No allocations found")
		return nil, errors.New("No allocations found for " + service.Name)
	}

	// See which allocation matches
	var allocationMatched *Allocation
	for _, a := range context.Allocations {
		m := a.matches(labels)
		tracing.log(depth + 1, "[%t] Testing allocation '%s': %s", m, a.Name, a.Criteria)
		if m {
			allocationMatched = a
			break
		}
	}

	// Check errors and resolve allocation name (it can be dynamic, depending on user labels)
	if allocationMatched != nil {
		err := allocationMatched.resolveName(user)
		if err != nil {
			glog.Infof("Cannot resolve name for an allocation: '%s' (context = %s, service = %s, user = %s)", allocationMatched.Name, context.Name, service.Name, user.Name)
			return nil, nil
		}
		glog.Infof("Matched allocation: '%s' -> '%s' (context = %s, service = %s, user = %s)", allocationMatched.Name, allocationMatched.NameResolved, context.Name, service.Name, user.Name)
	} else {
		glog.Infof("No allocation matched (context = %s, service = %s, user = %s)", context.Name, service.Name, user.Name)
	}

	return allocationMatched, nil
}
