package resolve

import (
	. "github.com/Aptomi/aptomi/pkg/slinga/language"
	. "github.com/Aptomi/aptomi/pkg/slinga/util"
)

// PolicyResolution contains resolution data for the policy
// When adding new fields to this object, it's crucial to modify appendData() method as well (!)
type PolicyResolution struct {
	// resolved component instances: componentKey -> componentInstance
	ComponentInstanceMap map[string]*ComponentInstance

	// resolved dependencies: dependencyID -> serviceKey
	DependencyInstanceMap map[string]string

	// resolved component processing order in which components/services have to be processed
	componentProcessingOrderHas map[string]bool
	ComponentProcessingOrder    []string
}

// NewPolicyResolution creates new empty PolicyResolution
func NewPolicyResolution() *PolicyResolution {
	return &PolicyResolution{
		ComponentInstanceMap:        make(map[string]*ComponentInstance),
		DependencyInstanceMap:       make(map[string]string),
		componentProcessingOrderHas: make(map[string]bool),
		ComponentProcessingOrder:    []string{},
	}
}

// Gets a component instance entry or creates an new entry if it doesn't exist
func (resolution *PolicyResolution) GetComponentInstanceEntry(cik *ComponentInstanceKey) *ComponentInstance {
	key := cik.GetKey()
	if _, ok := resolution.ComponentInstanceMap[key]; !ok {
		resolution.ComponentInstanceMap[key] = newComponentInstance(cik)
	}
	return resolution.ComponentInstanceMap[key]
}

// Record dependency for component instance
func (resolution *PolicyResolution) RecordResolved(cik *ComponentInstanceKey, dependency *Dependency) {
	instance := resolution.GetComponentInstanceEntry(cik)
	instance.addDependency(dependency.GetID())
	resolution.recordProcessingOrder(cik)
}

// Record processing order for component instance
func (resolution *PolicyResolution) recordProcessingOrder(cik *ComponentInstanceKey) {
	key := cik.GetKey()
	if !resolution.componentProcessingOrderHas[key] {
		resolution.componentProcessingOrderHas[key] = true
		resolution.ComponentProcessingOrder = append(resolution.ComponentProcessingOrder, key)
	}
}

// Stores calculated discovery params for component instance
func (resolution *PolicyResolution) RecordCodeParams(cik *ComponentInstanceKey, codeParams NestedParameterMap) error {
	return resolution.GetComponentInstanceEntry(cik).addCodeParams(codeParams)
}

// Stores calculated discovery params for component instance
func (resolution *PolicyResolution) RecordDiscoveryParams(cik *ComponentInstanceKey, discoveryParams NestedParameterMap) error {
	return resolution.GetComponentInstanceEntry(cik).addDiscoveryParams(discoveryParams)
}

// Stores calculated labels for component instance
func (resolution *PolicyResolution) RecordLabels(cik *ComponentInstanceKey, labels *LabelSet) {
	resolution.GetComponentInstanceEntry(cik).addLabels(labels)
}

// Stores an outgoing edge for component instance as we are traversing the graph
func (resolution *PolicyResolution) StoreEdge(src *ComponentInstanceKey, dst *ComponentInstanceKey) {
	// Arrival key can be empty at the very top of the recursive function in engine, so let's check for that
	if src != nil && dst != nil {
		resolution.GetComponentInstanceEntry(src).addEdgeOut(dst.GetKey())
		resolution.GetComponentInstanceEntry(dst).addEdgeIn(src.GetKey())
	}
}

// Appends data to the current PolicyResolution
func (resolution *PolicyResolution) AppendData(ops *PolicyResolution) error {
	for _, instance := range ops.ComponentInstanceMap {
		err := resolution.GetComponentInstanceEntry(instance.Key).appendData(instance)
		if err != nil {
			return err
		}
	}
	for key := range ops.ComponentInstanceMap {
		resolution.recordProcessingOrder(ops.ComponentInstanceMap[key].Key)
	}
	return nil
}
