package slinga

import (
	"errors"
	"log"
)

/*
 	Core engine for Slinga processing and evaluation
  */

// Evaluate "<user> needs <service>" statement
func (state *Policy) resolve(user User, serviceName string) (ServiceUsageState, error) {
	result := NewServiceUsageState()
	result.recordDependency(user, serviceName)
	err := state.resolveWithLabels(user, serviceName, user.getLabelSet(), &result)
	return result, err
}

// Evaluate "<user> needs <service>" statement
func (state *Policy) resolveWithLabels(user User, serviceName string, labels LabelSet, result *ServiceUsageState) (error) {

	// Locate the service
	service, err := state.getService(serviceName)
	if err != nil {
		return err
	}

	// Match the context
	context, err := state.getMatchedContext(*service, user, labels)
	if err != nil {
		return err
	}
	// If no matching context is found, let's just exit
	if context == nil {
		return nil
	}

	// Process context and transform labels
	labels = labels.applyTransform(context.Labels)

	// Match the allocation
	allocation, err := state.getMatchedAllocation(*service, user, *context, labels)
	if err != nil {
		return err
	}
	// If no matching allocation is found, let's just exit
	if allocation == nil {
		return nil
	}

	// Process allocation and transform labels
	labels = labels.applyTransform(allocation.Labels)

	// Now, sort all components in topological order
	err = service.sortComponentsTopologically()
	if err != nil {
		return err
	}

	// Record usage of a given service
	result.recordUsage(user, service, context, allocation, nil)

	// Resolve every component
	for _, component := range service.ComponentsOrdered {
		// Process component and transform labels
		labels = labels.applyTransform(component.Labels)

		// Is it a code?
		if component.Code != "" {
			log.Println("Processing dependency on code execution: " + component.Name + " (in " + service.Name + ")")
		} else if component.Service != "" {
			log.Println("Processing dependency on another service: " + component.Name + " -> " + component.Service + " (in " + service.Name + ")")
			err := state.resolveWithLabels(user, component.Service, labels, result)
			if err != nil {
				return err
			}
		} else {
			log.Fatal("Invalid component: " + component.Name + " " + service.Name)
		}

		// Record usage of a given component
		result.recordUsage(user, service, context, allocation, &component)
	}

	return nil
}

// Topologically sort components and return true if there is a cycle detected
func (service *Service) dfsComponentSort(u ServiceComponent, colors map[string]int) bool {
	colors[u.Name] = 1

	for _, vName := range u.Dependencies {
		v, exists := service.ComponentsMap[vName]
		if !exists {
			log.Fatal("Invalid dependency in service " + service.Name + ": " + vName)
		}
		if vColor, ok := colors[v.Name]; !ok {
			// not visited yet -> visit and exit if a cycle was found
			if service.dfsComponentSort(v, colors) {
				return true;
			}
		} else if vColor == 1 {
			return true;
		}
	}

	service.ComponentsOrdered = append(service.ComponentsOrdered, u)
	colors[u.Name] = 2;
	return false;
}

// Orders all components in a topological way
func (service *Service) sortComponentsTopologically() error {
	// Put all components into map
	service.ComponentsMap = make(map[string]ServiceComponent)
	for _, c := range service.Components {
		service.ComponentsMap[c.Name] = c
	}

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
		return errors.New("Component cycle detected in service " + service.Name)
	}

	return nil
}

// Helper to get a service
func (state *Policy) getService(serviceName string) (*Service, error) {
	// Locate the service
	service, ok := state.Services[serviceName]
	if !ok {
		return nil, errors.New("Service " + serviceName + " not found")
	}
	return &service, nil
}

// Helper to get a matched context
func (state *Policy) getMatchedContext(service Service, user User, labels LabelSet) (*Context, error) {
	// Locate the list of contexts for service
	contexts, ok := state.Contexts[service.Name]
	if !ok {
		return nil, errors.New("No contexts found for " + service.Name)
	}

	// See which context matches
	var contextMatched *Context
	for _, c := range contexts {
		if c.matches(labels) {
			contextMatched = &c
			break
		}
	}

	if contextMatched != nil {
		log.Printf("Matched context: '%s' (service = %s, user = %s)", contextMatched.Name, service.Name, user.Name)
	} else {
		log.Printf("No context matched (service = %s, user = %s)", service.Name, user.Name)
	}
	return contextMatched, nil
}

// Helper to get a matched allocation
func (state *Policy) getMatchedAllocation(service Service, user User, context Context, labels LabelSet) (*Allocation, error) {
	// See which allocation matches
	var allocationMatched *Allocation
	for _, a := range context.Allocations {
		if a.matches(labels) {
			allocationMatched = &a
			break
		}
	}

	// Check errors and resolve allocation name (it can be dynamic, depending on user labels)
	if allocationMatched != nil {
		err := allocationMatched.resolveName(user)
		if err != nil {
			log.Printf("Cannot resolve name for an allocation: '%s' (context = %s, service = %s, user = %s)", allocationMatched.Name, context.Name, service.Name, user.Name)
			return nil, nil
		}
		log.Printf("Matched allocation: '%s' -> '%s' (context = %s, service = %s, user = %s)", allocationMatched.Name, allocationMatched.NameResolved, context.Name, service.Name, user.Name)
	} else {
		log.Printf("No allocation matched (context = %s, service = %s, user = %s)", context.Name, service.Name, user.Name)
	}

	return allocationMatched, nil
}
