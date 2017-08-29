package resolve

import (
	. "github.com/Aptomi/aptomi/pkg/slinga/language"
	. "github.com/Aptomi/aptomi/pkg/slinga/util"
	"time"
)

// ServiceUsageState contains resolution data for services - who is using what, as well as contains processing order and additional data
type ServiceUsageState struct {
	// Date when it was created
	CreatedOn time.Time

	// Resolved usage - stores full information about dependencies which have been successfully resolved
	ResolvedData *ServiceUsageData

	// Unresolved usage - stores full information about dependencies which were not resolved
	UnresolvedData *ServiceUsageData
}

// ServiceUsageData contains all the data that gets resolved for one or more dependencies
// When adding new fields to this object, it's crucial to modify appendData() method as well (!)
type ServiceUsageData struct {
	// resolved component instances: componentKey -> componentInstance
	ComponentInstanceMap map[string]*ComponentInstance

	// resolved component processing order in which components/services have to be processed
	componentProcessingOrderHas map[string]bool
	ComponentProcessingOrder    []string
}

// NewServiceUsageData creates new empty ServiceUsageData
func NewServiceUsageData() *ServiceUsageData {
	return &ServiceUsageData{
		ComponentInstanceMap:        make(map[string]*ComponentInstance),
		componentProcessingOrderHas: make(map[string]bool),
		ComponentProcessingOrder:    []string{},
	}
}

// NewServiceUsageState creates new empty ServiceUsageState
func NewServiceUsageState() *ServiceUsageState {
	return &ServiceUsageState{
		CreatedOn:      time.Now(),
		ResolvedData:   NewServiceUsageData(),
		UnresolvedData: NewServiceUsageData(),
	}
}

// Gets a component instance entry or creates an new entry if it doesn't exist
func (data *ServiceUsageData) GetComponentInstanceEntry(cik *ComponentInstanceKey) *ComponentInstance {
	key := cik.GetKey()
	if _, ok := data.ComponentInstanceMap[key]; !ok {
		data.ComponentInstanceMap[key] = newComponentInstance(cik)
	}
	return data.ComponentInstanceMap[key]
}

// Record dependency for component instance
func (data *ServiceUsageData) RecordResolved(cik *ComponentInstanceKey, dependency *Dependency) {
	instance := data.GetComponentInstanceEntry(cik)
	instance.setResolved(true)
	instance.addDependency(dependency.GetID())
	data.recordProcessingOrder(cik)
}

// Record processing order for component instance
func (data *ServiceUsageData) recordProcessingOrder(cik *ComponentInstanceKey) {
	key := cik.GetKey()
	if !data.componentProcessingOrderHas[key] {
		data.componentProcessingOrderHas[key] = true
		data.ComponentProcessingOrder = append(data.ComponentProcessingOrder, key)
	}
}

// Stores calculated discovery params for component instance
func (data *ServiceUsageData) RecordCodeParams(cik *ComponentInstanceKey, codeParams NestedParameterMap) error {
	return data.GetComponentInstanceEntry(cik).addCodeParams(codeParams)
}

// Stores calculated discovery params for component instance
func (data *ServiceUsageData) RecordDiscoveryParams(cik *ComponentInstanceKey, discoveryParams NestedParameterMap) error {
	return data.GetComponentInstanceEntry(cik).addDiscoveryParams(discoveryParams)
}

// Stores calculated labels for component instance
func (data *ServiceUsageData) RecordLabels(cik *ComponentInstanceKey, labels LabelSet) {
	data.GetComponentInstanceEntry(cik).addLabels(labels)
}

// Stores an outgoing edge for component instance as we are traversing the graph
func (data *ServiceUsageData) StoreEdge(src *ComponentInstanceKey, dst *ComponentInstanceKey) {
	// Arrival key can be empty at the very top of the recursive function in engine, so let's check for that
	if src != nil && dst != nil {
		data.GetComponentInstanceEntry(src).addEdgeOut(dst.GetKey())
		data.GetComponentInstanceEntry(dst).addEdgeIn(src.GetKey())
	}
}

// Appends data to the current ServiceUsageData
func (data *ServiceUsageData) AppendData(ops *ServiceUsageData) error {
	for _, instance := range ops.ComponentInstanceMap {
		err := data.GetComponentInstanceEntry(instance.Key).appendData(instance)
		if err != nil {
			return err
		}
	}
	for _, key := range ops.ComponentProcessingOrder {
		data.recordProcessingOrder(ops.ComponentInstanceMap[key].Key)
	}
	return nil
}
