package engine

import (
	. "github.com/Aptomi/aptomi/pkg/slinga/language"
	. "github.com/Aptomi/aptomi/pkg/slinga/log"
	. "github.com/Aptomi/aptomi/pkg/slinga/util"
	log "github.com/Sirupsen/logrus"
	"time"
)

// ComponentInstance is a usage data for a given component instance, containing list of user IDs and calculated labels
// When adding new fields to this object, it's crucial to modify appendData() method as well (!)
type ComponentInstance struct {
	// Whether or not component instance has been resolved
	Resolved bool

	// Key
	Key *ComponentInstanceKey

	// When this instance was created & last updated on
	CreatedOn time.Time
	UpdatedOn time.Time

	// List of dependencies which are keeping this component instantiated
	DependencyIds map[string]bool

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
func newComponentInstance(cik *ComponentInstanceKey) *ComponentInstance {
	return &ComponentInstance{
		Key:                  cik,
		DependencyIds:        make(map[string]bool),
		CalculatedLabels:     NewLabelSetEmpty(),
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

func (instance *ComponentInstance) setResolved(resolved bool) {
	if !instance.Resolved {
		instance.Resolved = resolved
	} else if !resolved {
		Debug.WithFields(log.Fields{
			"key": instance.Key,
		}).Panic("Invalid action. Trying to unset Resolved flag for instance")
	}
}

func (instance *ComponentInstance) addDependency(dependencyID string) {
	instance.DependencyIds[dependencyID] = true
}

func (instance *ComponentInstance) addCodeParams(codeParams NestedParameterMap) {
	if len(instance.CalculatedCodeParams) == 0 {
		// Record code parameters
		instance.CalculatedCodeParams = codeParams
	} else if !instance.CalculatedCodeParams.DeepEqual(codeParams) {
		// Same component instance, different code parameters
		Debug.WithFields(log.Fields{
			"key":            instance.Key,
			"prevCodeParams": instance.CalculatedCodeParams,
			"nextCodeParams": codeParams,
		}).Panic("Invalid policy. Arrived to the same component with different code parameters")
	}
}

func (instance *ComponentInstance) addDiscoveryParams(discoveryParams NestedParameterMap) {
	if len(instance.CalculatedDiscovery) == 0 {
		// Record discovery parameters
		instance.CalculatedDiscovery = discoveryParams
	} else if !instance.CalculatedDiscovery.DeepEqual(discoveryParams) {
		// Same component instance, different discovery parameters
		Debug.WithFields(log.Fields{
			"key": instance.Key,
			"prevDiscoveryParams": instance.CalculatedDiscovery,
			"nextDiscoveryParams": discoveryParams,
		}).Panic("Invalid policy. Arrived to the same component with different discovery parameters")
	}
}

func (instance *ComponentInstance) addLabels(labels LabelSet) {
	// Unfortunately it's pretty typical for us to come with different labels to a component instance, let's combine them all
	instance.CalculatedLabels = instance.CalculatedLabels.AddLabels(labels)
}

func (instance *ComponentInstance) addRuleLogEntries(dependencyID string, entry ...*RuleLogEntry) {
	instance.RuleLog[dependencyID] = append(instance.RuleLog[dependencyID], entry...)
}

func (instance *ComponentInstance) addEdgeIn(srcKey string) {
	instance.EdgesIn[srcKey] = true
}

func (instance *ComponentInstance) addEdgeOut(dstKey string) {
	instance.EdgesOut[dstKey] = true
}

func (instance *ComponentInstance) updateTimes(createdOn time.Time, updatedOn time.Time) {
	if time.Time.IsZero(instance.CreatedOn) || (!time.Time.IsZero(createdOn) && createdOn.Before(instance.CreatedOn)) {
		instance.CreatedOn = createdOn
	}
	if !time.Time.IsZero(updatedOn) && updatedOn.After(instance.UpdatedOn) {
		instance.UpdatedOn = updatedOn
	}
}

func (instance *ComponentInstance) checkTimesAreEmpty() {
	if !time.Time.IsZero(instance.CreatedOn) || !time.Time.IsZero(instance.UpdatedOn) {
		// Same component instance, different code parameters
		Debug.WithFields(log.Fields{
			"key":       instance.Key,
			"createdOn": instance.CreatedOn,
			"updatedOn": instance.UpdatedOn,
		}).Panic("Expected zero times, but found non-zero times")
	}
}

func (instance *ComponentInstance) appendData(ops *ComponentInstance) {
	// Resolution flag
	instance.setResolved(ops.Resolved)

	// Times should not be initialized yet
	instance.checkTimesAreEmpty()

	// List of dependencies which are keeping this component instantiated
	for dependencyID := range ops.DependencyIds {
		instance.addDependency(dependencyID)
	}

	// Calculated parameters for the component
	instance.addLabels(ops.CalculatedLabels)
	instance.addDiscoveryParams(ops.CalculatedDiscovery)
	instance.addCodeParams(ops.CalculatedCodeParams)

	// Incoming and outgoing graph edges (instance: key -> true) as we are traversing the graph
	for key := range ops.EdgesIn {
		instance.addEdgeIn(key)
	}
	for keyDst := range ops.EdgesOut {
		instance.addEdgeOut(keyDst)
	}

	// Rule evaluation log (dependency ID -> []*RuleLogEntry)
	for dependencyID, entryList := range ops.RuleLog {
		instance.addRuleLogEntries(dependencyID, entryList...)
	}
}
