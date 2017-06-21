package slinga

import "time"

// ComponentInstance is a usage data for a given component instance, containing list of user IDs and calculated labels
type ComponentInstance struct {
	// When this instance was created & last updated on
	CreatedOn time.Time
	UpdatedOn time.Time

	// List of dependencies which are keeping this component instantiated
	DependencyIds []string

	// Calculated parameters for the component
	CalculatedLabels     LabelSet
	CalculatedDiscovery  NestedParameterMap
	CalculatedCodeParams NestedParameterMap

	// Incoming and outgoing graph edges (instance: key -> true) as we are traversing the graph
	EdgesIn  map[string]bool
	EdgesOut map[string]bool

	// Rule evaluation log (dependency ID -> []*RuleLogEntry)
	RuleLog map[string][]*RuleLogEntry
}

// Creates a new component instance
func newComponentInstance() *ComponentInstance {
	return &ComponentInstance{
		CalculatedLabels:     LabelSet{},
		CalculatedDiscovery:  NestedParameterMap{},
		CalculatedCodeParams: NestedParameterMap{},
		EdgesIn:              make(map[string]bool),
		EdgesOut:             make(map[string]bool),
		RuleLog:              make(map[string][]*RuleLogEntry),
	}
}

// GetRunningTime returns the time for long component has been running
func (instance *ComponentInstance) GetRunningTime() time.Duration {
	return time.Since(instance.CreatedOn)
}
