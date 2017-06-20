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

	// resolved usage - gets calculated by the main engine. should be accessed by a getter
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

// ComponentInstance is a usage data for a given component instance, containing list of user IDs and calculated labels
type ComponentInstance struct {
	// People who are using this component
	UserIds []string

	// Calculated parameters for the component
	CalculatedLabels     LabelSet
	CalculatedDiscovery  NestedParameterMap
	CalculatedCodeParams NestedParameterMap
}

// NewResolvedServiceUsageData creates new empty ResolvedServiceUsageData
func NewResolvedServiceUsageData() *ResolvedServiceUsageData {
	return &ResolvedServiceUsageData{
		ComponentInstanceMap:        make(map[string]*ComponentInstance),
		componentProcessingOrderHas: make(map[string]bool)}
}

// NewServiceUsageState creates new empty ServiceUsageState
func NewServiceUsageState(policy *Policy, dependencies *GlobalDependencies, users *GlobalUsers) ServiceUsageState {
	return ServiceUsageState{
		Policy:        policy,
		Dependencies:  dependencies,
		users:         users,
		ResolvedUsage: NewResolvedServiceUsageData()}
}

// Records usage event
func (usage *ServiceUsageState) getResolvedUsage() *ResolvedServiceUsageData {
	if usage.ResolvedUsage == nil {
		usage.ResolvedUsage = NewResolvedServiceUsageData()
	}
	return usage.ResolvedUsage
}

// Records usage event
func (resolvedUsage *ResolvedServiceUsageData) recordUsage(key string, user *User) string {
	// Add user to the entry
	usageStruct := resolvedUsage.getComponentInstanceEntry(key)
	usageStruct.UserIds = append(usageStruct.UserIds, user.ID)

	// Add to processing order
	if !resolvedUsage.componentProcessingOrderHas[key] {
		resolvedUsage.componentProcessingOrderHas[key] = true
		resolvedUsage.ComponentProcessingOrder = append(resolvedUsage.ComponentProcessingOrder, key)
	}

	return key
}

// Stores calculated discovery params for component instance
func (resolvedUsage *ResolvedServiceUsageData) storeCodeParams(key string, codeParams NestedParameterMap) {
	cInstance := resolvedUsage.getComponentInstanceEntry(key)
	if len(cInstance.CalculatedCodeParams) == 0 {
		// Record code parameters
		cInstance.CalculatedCodeParams = codeParams
	} else if !cInstance.CalculatedCodeParams.deepEqual(codeParams) {
		// Same component instance, different code parameters
		debug.WithFields(log.Fields{
			"componentKey":   key,
			"prevCodeParams": cInstance.CalculatedCodeParams,
			"nextCodeParams": codeParams,
		}).Fatal("Invalid policy. Arrived to the same component with different code parameters")
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
		}).Fatal("Invalid policy. Arrived to the same component with different discovery parameters")
	}
}

// Stores calculated labels for component instance
func (resolvedUsage *ResolvedServiceUsageData) storeLabels(key string, labels LabelSet) {
	cInstance := resolvedUsage.getComponentInstanceEntry(key)

	// Unfortunately it's pretty typical for us to come with different labels to a component instance, let's combine them all
	cInstance.CalculatedLabels = cInstance.CalculatedLabels.addLabels(labels)
}

// Gets a component instance entry or creates an new entry if it doesn't exist
func (resolvedUsage *ResolvedServiceUsageData) getComponentInstanceEntry(key string) *ComponentInstance {
	if _, ok := resolvedUsage.ComponentInstanceMap[key]; !ok {
		resolvedUsage.ComponentInstanceMap[key] = &ComponentInstance{
			CalculatedLabels:     LabelSet{},
			CalculatedDiscovery:  NestedParameterMap{},
			CalculatedCodeParams: NestedParameterMap{}}
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
