package resolve

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/slinga/errors"
	. "github.com/Aptomi/aptomi/pkg/slinga/language"
	"github.com/Aptomi/aptomi/pkg/slinga/object"
	. "github.com/Aptomi/aptomi/pkg/slinga/util"
	"time"
)

// ComponentInstance is a struct that holds data for a given component instance, containing list of user IDs and calculated labels
// When adding new fields to this object, it's crucial to modify appendData() method as well (!)
type ComponentInstance struct {
	/*
		These fields get populated during policy resolution
	*/
	object.Metadata // todo it's temporarily

	// Key
	Key *ComponentInstanceKey

	// List of dependencies which are keeping this component instantiated
	DependencyIds map[string]bool

	// Calculated parameters for the component
	CalculatedLabels     *LabelSet
	CalculatedDiscovery  NestedParameterMap
	CalculatedCodeParams NestedParameterMap

	// Incoming and outgoing graph edges (instance: key -> true) as we are traversing the graph
	EdgesIn  map[string]bool
	EdgesOut map[string]bool

	/*
		These fields get populated during apply and desired -> actual state reconciliation
	*/

	// When this instance was created & last updated on
	CreatedOn time.Time
	UpdatedOn time.Time
}

// Creates a new component instance
func newComponentInstance(cik *ComponentInstanceKey) *ComponentInstance {
	return &ComponentInstance{
		Key:                  cik,
		DependencyIds:        make(map[string]bool),
		CalculatedLabels:     NewLabelSet(make(map[string]string)),
		CalculatedDiscovery:  NestedParameterMap{},
		CalculatedCodeParams: NestedParameterMap{},
		EdgesIn:              make(map[string]bool),
		EdgesOut:             make(map[string]bool),
	}
}

// GetRunningTime returns the time for long component has been running
func (instance *ComponentInstance) GetRunningTime() time.Duration {
	return time.Since(instance.CreatedOn)
}

func (instance *ComponentInstance) addDependency(dependencyID string) {
	instance.DependencyIds[dependencyID] = true
}

func (instance *ComponentInstance) addCodeParams(codeParams NestedParameterMap) error {
	if len(instance.CalculatedCodeParams) == 0 {
		// Record code parameters
		instance.CalculatedCodeParams = codeParams
	} else if !instance.CalculatedCodeParams.DeepEqual(codeParams) {
		// Same component instance, different code parameters
		return errors.NewErrorWithDetails(
			fmt.Sprintf("Invalid policy. Conflicting code parameters for component instance: %s", instance.Key.GetKey()),
			errors.Details{
				"instance":             instance.Key,
				"code_params_existing": instance.CalculatedCodeParams,
				"code_params_new":      codeParams,
				"diff":                 instance.CalculatedCodeParams.Diff(codeParams),
			},
		)
	}
	return nil
}

func (instance *ComponentInstance) addDiscoveryParams(discoveryParams NestedParameterMap) error {
	if len(instance.CalculatedDiscovery) == 0 {
		// Record discovery parameters
		instance.CalculatedDiscovery = discoveryParams
	} else if !instance.CalculatedDiscovery.DeepEqual(discoveryParams) {
		// Same component instance, different discovery parameters
		return errors.NewErrorWithDetails(
			fmt.Sprintf("Invalid policy. Conflicting discovery parameters for component instance: %s", instance.Key.GetKey()),
			errors.Details{
				"instance":                  instance.Key,
				"discovery_params_existing": instance.CalculatedDiscovery,
				"discovery_params_new":      discoveryParams,
				"diff":                      instance.CalculatedDiscovery.Diff(discoveryParams),
			},
		)
	}
	return nil
}

func (instance *ComponentInstance) addLabels(labels *LabelSet) {
	// it's pretty typical for us to come with different labels to a component instance, let's combine them all
	instance.CalculatedLabels.AddLabels(labels.Labels)
}

func (instance *ComponentInstance) addEdgeIn(srcKey string) {
	instance.EdgesIn[srcKey] = true
}

func (instance *ComponentInstance) addEdgeOut(dstKey string) {
	instance.EdgesOut[dstKey] = true
}

func (instance *ComponentInstance) UpdateTimes(createdOn time.Time, updatedOn time.Time) {
	if time.Time.IsZero(instance.CreatedOn) || (!time.Time.IsZero(createdOn) && createdOn.Before(instance.CreatedOn)) {
		instance.CreatedOn = createdOn
	}
	if !time.Time.IsZero(updatedOn) && updatedOn.After(instance.UpdatedOn) {
		instance.UpdatedOn = updatedOn
	}
}

func (instance *ComponentInstance) appendData(ops *ComponentInstance) error {
	// List of dependencies which are keeping this component instantiated
	for dependencyID := range ops.DependencyIds {
		instance.addDependency(dependencyID)
	}

	// Combine labels
	instance.addLabels(ops.CalculatedLabels)

	// Combine code params and discovery params for
	var err = instance.addDiscoveryParams(ops.CalculatedDiscovery)
	if err != nil {
		return err
	}

	err = instance.addCodeParams(ops.CalculatedCodeParams)
	if err != nil {
		return err
	}

	// Incoming and outgoing graph edges (instance: key -> true) as we are traversing the graph
	for key := range ops.EdgesIn {
		instance.addEdgeIn(key)
	}
	for keyDst := range ops.EdgesOut {
		instance.addEdgeOut(keyDst)
	}

	return nil
}
