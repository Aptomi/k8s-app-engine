package slinga

import (
	log "github.com/Sirupsen/logrus"
)

// ServiceUsageState contains resolution data for services - who is using what, as well as contains processing order and additional data
type ServiceUsageState struct {
	// reference to a policy
	Policy *Policy

	// reference to dependencies
	Dependencies *GlobalDependencies

	// reference to users
	users *GlobalUsers

	// resolved usage - gets calculated by the main engine. should ideally be accessed by a getter
	ResolvedUsage *ResolvedServiceUsageData
}

// ResolvedServiceUsageData contains all the data that gets resolved for one or more dependencies
type ResolvedServiceUsageData struct {
	// resolved component instances: componentKey -> componentInstance
	ComponentInstanceMap map[string]*ComponentInstance

	// resolved component processing order in which components/services have to be processed
	componentProcessingOrderHas map[string]bool
	ComponentProcessingOrder    []string
}

// NewResolvedServiceUsageData creates new empty ResolvedServiceUsageData
func newResolvedServiceUsageData() *ResolvedServiceUsageData {
	return &ResolvedServiceUsageData{
		ComponentInstanceMap:        make(map[string]*ComponentInstance),
		componentProcessingOrderHas: make(map[string]bool),
	}
}

// NewServiceUsageState creates new empty ServiceUsageState
func NewServiceUsageState(policy *Policy, dependencies *GlobalDependencies, users *GlobalUsers) ServiceUsageState {
	return ServiceUsageState{
		Policy:        policy,
		Dependencies:  dependencies,
		users:         users,
		ResolvedUsage: newResolvedServiceUsageData()}
}

// Records usage event
func (usage *ServiceUsageState) GetResolvedUsage() *ResolvedServiceUsageData {
	if usage.ResolvedUsage == nil {
		usage.ResolvedUsage = newResolvedServiceUsageData()
	}
	return usage.ResolvedUsage
}

// Records usage event
func (resolvedUsage *ResolvedServiceUsageData) recordUsage(key string, dependency *Dependency) string {
	// Add user to the entry
	instance := resolvedUsage.getComponentInstanceEntry(key)
	instance.DependencyIds = append(instance.DependencyIds, dependency.ID)

	// Add to processing order
	if !resolvedUsage.componentProcessingOrderHas[key] {
		resolvedUsage.componentProcessingOrderHas[key] = true
		resolvedUsage.ComponentProcessingOrder = append(resolvedUsage.ComponentProcessingOrder, key)
	}

	return key
}

// Stores calculated discovery params for component instance
func (resolvedUsage *ResolvedServiceUsageData) storeCodeParams(key string, codeParams NestedParameterMap) {
	instance := resolvedUsage.getComponentInstanceEntry(key)
	if len(instance.CalculatedCodeParams) == 0 {
		// Record code parameters
		instance.CalculatedCodeParams = codeParams
	} else if !instance.CalculatedCodeParams.deepEqual(codeParams) {
		// Same component instance, different code parameters
		debug.WithFields(log.Fields{
			"componentKey":   key,
			"prevCodeParams": instance.CalculatedCodeParams,
			"nextCodeParams": codeParams,
		}).Panic("Invalid policy. Arrived to the same component with different code parameters")
	}
}

// Stores calculated discovery params for component instance
func (resolvedUsage *ResolvedServiceUsageData) storeDiscoveryParams(key string, discoveryParams NestedParameterMap) {
	cInstance := resolvedUsage.getComponentInstanceEntry(key)
	if len(cInstance.CalculatedDiscovery) == 0 {
		// Record discovery parameters
		cInstance.CalculatedDiscovery = discoveryParams
	} else if !cInstance.CalculatedDiscovery.deepEqual(discoveryParams) {
		// Same component instance, different discovery parameters
		debug.WithFields(log.Fields{
			"componentKey":        key,
			"prevDiscoveryParams": cInstance.CalculatedDiscovery,
			"nextDiscoveryParams": discoveryParams,
		}).Panic("Invalid policy. Arrived to the same component with different discovery parameters")
	}
}

// Stores calculated labels for component instance
func (resolvedUsage *ResolvedServiceUsageData) storeLabels(key string, labels LabelSet) {
	cInstance := resolvedUsage.getComponentInstanceEntry(key)

	// Unfortunately it's pretty typical for us to come with different labels to a component instance, let's combine them all
	cInstance.CalculatedLabels = cInstance.CalculatedLabels.addLabels(labels)
}

// Stores an outgoing edge for component instance as we are traversing the graph
func (resolvedUsage *ResolvedServiceUsageData) storeEdge(key string, keyDst string) {
	// Arrival key can be empty at the very top of the recursive function in engine, so let's check for that
	if len(key) > 0 && len(keyDst) > 0 {
		resolvedUsage.getComponentInstanceEntry(key).EdgesOut[keyDst] = true
		resolvedUsage.getComponentInstanceEntry(keyDst).EdgesIn[key] = true
	}
}

// Stores rule log entry, attaching it to component instance by dependency
func (resolvedUsage *ResolvedServiceUsageData) storeRuleLogEntry(key string, dependency *Dependency, entry *RuleLogEntry) {
	instance := resolvedUsage.getComponentInstanceEntry(key)
	instance.RuleLog[dependency.ID] = append(instance.RuleLog[dependency.ID], entry)
}

// Gets a component instance entry or creates an new entry if it doesn't exist
func (resolvedUsage *ResolvedServiceUsageData) getComponentInstanceEntry(key string) *ComponentInstance {
	if _, ok := resolvedUsage.ComponentInstanceMap[key]; !ok {
		resolvedUsage.ComponentInstanceMap[key] = newComponentInstance()
	}
	return resolvedUsage.ComponentInstanceMap[key]
}

// LoadServiceUsageState loads usage state from a file under Aptomi DB
func LoadServiceUsageState() ServiceUsageState {
	lastRevision := GetLastRevision(GetAptomiBaseDir())
	fileName := GetAptomiObjectFileFromRun(GetAptomiBaseDir(), lastRevision, TypePolicyResolution, "db.yaml")
	return loadServiceUsageStateFromFile(fileName)
}

// SaveServiceUsageState stores usage state in a file under Aptomi DB
func (usage ServiceUsageState) SaveServiceUsageState() {
	fileName := GetAptomiObjectWriteFileCurrentRun(GetAptomiBaseDir(), TypePolicyResolution, "db.yaml")
	saveObjectToFile(fileName, usage)
}
