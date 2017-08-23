package engine

import (
	. "github.com/Aptomi/aptomi/pkg/slinga/db"
	. "github.com/Aptomi/aptomi/pkg/slinga/language"
	"github.com/Aptomi/aptomi/pkg/slinga/language/yaml"
	. "github.com/Aptomi/aptomi/pkg/slinga/util"
	"time"
)

// ServiceUsageState contains resolution data for services - who is using what, as well as contains processing order and additional data
type ServiceUsageState struct {
	// reference to a policy
	Policy *PolicyNamespace

	// user loader
	userLoader UserLoader

	// Date when it was created
	CreatedOn time.Time

	// Diff stored as text
	DiffAsText string

	// resolved usage - stores full information about dependencies which have been successfully resolved. should ideally be accessed by a getter
	ResolvedData *ServiceUsageData

	// unresolved usage - stores full information about dependencies which were not resolved. including rule logs with reasons, etc
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

// NewResolvedServiceUsageData creates new empty ServiceUsageData
func newServiceUsageData() *ServiceUsageData {
	return &ServiceUsageData{
		ComponentInstanceMap:        make(map[string]*ComponentInstance),
		componentProcessingOrderHas: make(map[string]bool),
		ComponentProcessingOrder:    []string{},
	}
}

// NewServiceUsageState creates new empty ServiceUsageState
func NewServiceUsageState(policy *PolicyNamespace, userLoader UserLoader) ServiceUsageState {
	return ServiceUsageState{
		Policy:         policy,
		userLoader:     userLoader,
		CreatedOn:      time.Now(),
		ResolvedData:   newServiceUsageData(),
		UnresolvedData: newServiceUsageData(),
	}
}

// GetUserLoader is a getter for a private field userLoader
// The field userLoader has to stay private, so it won't get serialized into YAML
func (state *ServiceUsageState) GetUserLoader() UserLoader {
	return state.userLoader
}

// GetResolvedData returns usage.ResolvedData
// TODO: we can get likely rid of this method (but need to check serialization, etc)
func (state *ServiceUsageState) GetResolvedData() *ServiceUsageData {
	if state.ResolvedData == nil {
		state.ResolvedData = newServiceUsageData()
	}
	return state.ResolvedData
}

// Gets a component instance entry or creates an new entry if it doesn't exist
func (data *ServiceUsageData) getComponentInstanceEntry(cik *ComponentInstanceKey) *ComponentInstance {
	key := cik.GetKey()
	if _, ok := data.ComponentInstanceMap[key]; !ok {
		data.ComponentInstanceMap[key] = newComponentInstance(cik)
	}
	return data.ComponentInstanceMap[key]
}

// Record dependency for component instance
func (data *ServiceUsageData) recordResolved(cik *ComponentInstanceKey, dependency *Dependency) {
	instance := data.getComponentInstanceEntry(cik)
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
func (data *ServiceUsageData) recordCodeParams(cik *ComponentInstanceKey, codeParams NestedParameterMap) {
	data.getComponentInstanceEntry(cik).addCodeParams(codeParams)
}

// Stores calculated discovery params for component instance
func (data *ServiceUsageData) recordDiscoveryParams(cik *ComponentInstanceKey, discoveryParams NestedParameterMap) {
	data.getComponentInstanceEntry(cik).addDiscoveryParams(discoveryParams)
}

// Stores calculated labels for component instance
func (data *ServiceUsageData) recordLabels(cik *ComponentInstanceKey, labels LabelSet) {
	// TODO: write into event log
	data.getComponentInstanceEntry(cik).addLabels(labels)
}

// Stores an outgoing edge for component instance as we are traversing the graph
func (data *ServiceUsageData) storeEdge(src *ComponentInstanceKey, dst *ComponentInstanceKey) {
	// Arrival key can be empty at the very top of the recursive function in engine, so let's check for that
	if src != nil && dst != nil {
		data.getComponentInstanceEntry(src).addEdgeOut(dst.GetKey())
		data.getComponentInstanceEntry(dst).addEdgeIn(src.GetKey())
	}
}

// Stores rule log entry, attaching it to component instance by dependency
func (data *ServiceUsageData) storeRuleLogEntry(cik *ComponentInstanceKey, dependency *Dependency, entry *RuleLogEntry) {
	data.getComponentInstanceEntry(cik).addRuleLogEntries(dependency.GetID(), entry)
}

// Appends data to the current ServiceUsageData
func (data *ServiceUsageData) appendData(ops *ServiceUsageData) {
	for _, instance := range ops.ComponentInstanceMap {
		data.getComponentInstanceEntry(instance.Key).appendData(instance)
	}
	for _, key := range ops.ComponentProcessingOrder {
		data.recordProcessingOrder(ops.ComponentInstanceMap[key].Key)
	}
}

// LoadServiceUsageState loads usage state from a file under Aptomi DB
func LoadServiceUsageState(userLoader UserLoader) ServiceUsageState {
	lastRevision := GetLastRevision(GetAptomiBaseDir())
	fileName := GetAptomiObjectFileFromRun(GetAptomiBaseDir(), lastRevision.GetRunDirectory(), TypePolicyResolution, "db.yaml")
	result := loadServiceUsageStateFromFile(fileName)
	result.userLoader = userLoader
	return result
}

// LoadServiceUsageStatesAll loads all usage states from files under Aptomi DB
func LoadServiceUsageStatesAll(userLoader UserLoader) map[int]ServiceUsageState {
	result := make(map[int]ServiceUsageState)
	lastRevision := GetLastRevision(GetAptomiBaseDir())
	for rev := lastRevision; rev > LastRevisionAbsentValue; rev-- {
		fileName := GetAptomiObjectFileFromRun(GetAptomiBaseDir(), rev.GetRunDirectory(), TypePolicyResolution, "db.yaml")
		state := loadServiceUsageStateFromFile(fileName)
		state.userLoader = userLoader
		if state.Policy != nil {
			// add only non-empty revisions. don't add revision which got deleted
			result[int(rev)] = state
		}
	}
	return result
}

// SaveServiceUsageState saves usage state in a file under Aptomi DB
func (state ServiceUsageState) SaveServiceUsageState() {
	fileName := GetAptomiObjectWriteFileCurrentRun(GetAptomiBaseDir(), TypePolicyResolution, "db.yaml")
	yaml.SaveObjectToFile(fileName, state)
}

// Loads usage state from file
func loadServiceUsageStateFromFile(fileName string) ServiceUsageState {
	result := *yaml.LoadObjectFromFileDefaultEmpty(fileName, new(ServiceUsageState)).(*ServiceUsageState)
	if result.Policy == nil {
		result.Policy = NewPolicyNamespace()
	}
	return result
}
