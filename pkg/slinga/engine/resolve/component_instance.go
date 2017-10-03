package resolve

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/slinga/errors"
	"github.com/Aptomi/aptomi/pkg/slinga/lang"
	"github.com/Aptomi/aptomi/pkg/slinga/object"
	"github.com/Aptomi/aptomi/pkg/slinga/util"
	"strconv"
	"time"
)

var ComponentInstanceObject = &object.Info{
	Kind:        "component-instance",
	Constructor: func() object.Base { return &ComponentInstance{} },
}

type ComponentInstanceMetadata struct {
	Key  *ComponentInstanceKey
	Kind string
}

const ALLOW_INGRESS = "allow_ingress"

// ComponentInstance is a struct that holds data for a given component instance, containing list of user IDs and calculated labels
// When adding new fields to this object, it's crucial to modify appendData() method as well (!)
type ComponentInstance struct {
	/*
		These fields get populated during policy resolution
	*/
	Metadata *ComponentInstanceMetadata

	// List of dependency keys which are keeping this component instantiated
	DependencyKeys map[string]bool

	// Calculated parameters for the component
	CalculatedLabels     *lang.LabelSet
	CalculatedDiscovery  util.NestedParameterMap
	CalculatedCodeParams util.NestedParameterMap

	// Incoming and outgoing graph edges (instance: key -> true) as we are traversing the graph
	EdgesIn  map[string]bool
	EdgesOut map[string]bool

	// Additional data recorded for use in plugins
	DataForPlugins map[string]string

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
		Metadata:             &ComponentInstanceMetadata{cik, ComponentInstanceObject.Kind},
		DependencyKeys:       make(map[string]bool),
		CalculatedLabels:     lang.NewLabelSet(make(map[string]string)),
		CalculatedDiscovery:  util.NestedParameterMap{},
		CalculatedCodeParams: util.NestedParameterMap{},
		EdgesIn:              make(map[string]bool),
		EdgesOut:             make(map[string]bool),
		DataForPlugins:       make(map[string]string),
	}
}

func (instance *ComponentInstance) GetKey() string {
	return instance.Metadata.Key.GetKey()
}

func (instance *ComponentInstance) GetNamespace() string {
	return object.SystemNS
}

func (instance *ComponentInstance) GetKind() string {
	return ComponentInstanceObject.Kind
}

func (instance *ComponentInstance) GetName() string {
	return instance.GetKey()
}

func (instance *ComponentInstance) GetGeneration() object.Generation {
	// we aren't storing multiple versions of ComponentInstance
	return 0
}

func (instance *ComponentInstance) SetGeneration(generation object.Generation) {
	panic("ComponentInstance isn't a versioned object")
}

// GetRunningTime returns the time for long component has been running
func (instance *ComponentInstance) GetRunningTime() time.Duration {
	return time.Since(instance.CreatedOn)
}

func (instance *ComponentInstance) addDependency(dependencyKey string) {
	instance.DependencyKeys[dependencyKey] = true
}

func (instance *ComponentInstance) addRuleInformation(result *lang.RuleActionResult) {
	instance.DataForPlugins[ALLOW_INGRESS] = strconv.FormatBool(result.AllowIngress)
}

func (instance *ComponentInstance) addCodeParams(codeParams util.NestedParameterMap) error {
	if len(instance.CalculatedCodeParams) == 0 {
		// Record code parameters
		instance.CalculatedCodeParams = codeParams
	} else if !instance.CalculatedCodeParams.DeepEqual(codeParams) {
		// Same component instance, different code parameters
		return errors.NewErrorWithDetails(
			fmt.Sprintf("Invalid policy. Conflicting code parameters for component instance: %s", instance.GetKey()),
			errors.Details{
				"instance":             instance.Metadata.Key,
				"code_params_existing": instance.CalculatedCodeParams,
				"code_params_new":      codeParams,
				"diff":                 instance.CalculatedCodeParams.Diff(codeParams),
			},
		)
	}
	return nil
}

func (instance *ComponentInstance) addDiscoveryParams(discoveryParams util.NestedParameterMap) error {
	if len(instance.CalculatedDiscovery) == 0 {
		// Record discovery parameters
		instance.CalculatedDiscovery = discoveryParams
	} else if !instance.CalculatedDiscovery.DeepEqual(discoveryParams) {
		// Same component instance, different discovery parameters
		return errors.NewErrorWithDetails(
			fmt.Sprintf("Invalid policy. Conflicting discovery parameters for component instance: %s", instance.GetKey()),
			errors.Details{
				"instance":                  instance.Metadata.Key,
				"discovery_params_existing": instance.CalculatedDiscovery,
				"discovery_params_new":      discoveryParams,
				"diff":                      instance.CalculatedDiscovery.Diff(discoveryParams),
			},
		)
	}
	return nil
}

func (instance *ComponentInstance) addLabels(labels *lang.LabelSet) {
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
	for dependencyKey := range ops.DependencyKeys {
		instance.addDependency(dependencyKey)
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

	// Data for plugins
	for k, v := range ops.DataForPlugins {
		instance.DataForPlugins[k] = v
	}

	return nil
}
