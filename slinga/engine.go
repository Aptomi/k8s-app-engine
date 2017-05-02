package slinga

import (
	"errors"
	"log"
)

/*
 	Core engine for Slinga processing and evaluation
  */

// Evaluate "<user> needs <service>" statement
func (state *GlobalState) resolve(user User, serviceName string) (interface{}, error) {

	// Locate the service
	service, ok := state.Services[serviceName]
	if !ok {
		return nil, errors.New("Service " + serviceName + " not found")
	}

	// Locate the list of contexts for service
	contexts, ok := state.Contexts[serviceName]
	if !ok {
		return nil, errors.New("No contexts found for " + serviceName)
	}

	// See which context matches
	labels := user.getLabelSet()
	var contextMatched *Context
	for _, c := range contexts {
		if c.matches(labels) {
			contextMatched = &c
			break
		}
	}

	if contextMatched == nil {
		log.Printf("No context matched (service = %s, user = %s)", service.Name, user.Name)
		return nil, nil
	}
	log.Printf("Matched context: '%s' (service = %s, user = %s)", contextMatched.Name, service.Name, user.Name)

	// Transform labels
	labels = labels.applyTransform(contextMatched.Labels)

	// See which allocation matches
	var allocationMatched *Allocation
	for _, a := range contextMatched.Allocations {
		if a.matches(labels) {
			allocationMatched = &a
			break
		}
	}

	if allocationMatched == nil {
		log.Printf("No allocation matched (context = %s, service = %s, user = %s)", contextMatched.Name, service.Name, user.Name)
		return nil, nil
	}
	err := allocationMatched.resolveName(user)
	if err != nil {
		log.Printf("Cannot resolve name for an allocation: '%s' (context = %s, service = %s, user = %s)", allocationMatched.Name, contextMatched.Name, service.Name, user.Name)
		return nil, nil
	}

	log.Printf("Matched allocation: '%s' -> '%s' (context = %s, service = %s, user = %s)", allocationMatched.Name, allocationMatched.NameResolved, contextMatched.Name, service.Name, user.Name)

	return nil, nil
}
