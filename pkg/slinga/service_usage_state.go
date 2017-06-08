package slinga

import (
	log "github.com/Sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

const componentRootName = "root"

// ServiceUsageState contains resolution data for services - who is using what, as well as contains processing order and additional data
type ServiceUsageState struct {
	// reference to a policy
	Policy *Policy

	// reference to dependencies
	Dependencies *GlobalDependencies

	// reference to users
	users *GlobalUsers

	// resolved usage - gets calculated by the main engine
	ResolvedUsage ResolvedServiceUsageData
}

// ResolvedServiceUsageData contains all the data that gets resolved for one or more dependencies
type ResolvedServiceUsageData struct {
	// resolved component instances: componentKey -> componentInstance
	ComponentInstanceMap map[string]*ComponentInstance

	// resolved component processing order in which components/services have to be processed
	componentProcessingOrderHas map[string]bool
	ComponentProcessingOrder    []string

	// resolved component discovery tree (component1.component2...component3 -> component instance key)
	DiscoveryTree NestedParameterMap
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
func NewResolvedServiceUsageData() ResolvedServiceUsageData {
	return ResolvedServiceUsageData{
		ComponentInstanceMap:        make(map[string]*ComponentInstance),
		DiscoveryTree:               NestedParameterMap{},
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
	// TODO: what to do if we came here multiple times with different code params?
	resolvedUsage.getComponentInstanceEntry(key).CalculatedCodeParams = codeParams
}

// Stores calculated discovery params for component instance
func (resolvedUsage *ResolvedServiceUsageData) storeDiscoveryParams(key string, discoveryParams NestedParameterMap) {
	// TODO: what to do if we came here multiple times with different discovery params?
	resolvedUsage.getComponentInstanceEntry(key).CalculatedDiscovery = discoveryParams
}

// Stores calculated labels for component instance
func (resolvedUsage *ResolvedServiceUsageData) storeLabels(key string, labels LabelSet) {
	// TODO: we can arrive to a service via multiple usages with different labels. what to do?
	resolvedUsage.getComponentInstanceEntry(key).CalculatedLabels = labels
}

// Gets a component instance entry or creates an new entry if it doesn't exist
func (resolvedUsage *ResolvedServiceUsageData) getComponentInstanceEntry(key string) *ComponentInstance {
	if _, ok := resolvedUsage.ComponentInstanceMap[key]; !ok {
		resolvedUsage.ComponentInstanceMap[key] = &ComponentInstance{CalculatedLabels: LabelSet{}}
	}
	return resolvedUsage.ComponentInstanceMap[key]
}

// LoadServiceUsageState loads usage state from a file under Aptomi DB
func LoadServiceUsageState() ServiceUsageState {
	fileName := GetAptomiDBDir() + "/" + "db.yaml"

	debug.WithFields(log.Fields{
		"file": fileName,
	}).Info("Loading service usage state")

	dat, e := ioutil.ReadFile(fileName)

	// If the file doesn't exist, it means that DB is empty and we are starting from scratch
	if os.IsNotExist(e) {
		return ServiceUsageState{}
	}

	if e != nil {
		debug.WithFields(log.Fields{
			"file":  fileName,
			"error": e,
		}).Fatal("Unable to read file", e)
	}

	t := ServiceUsageState{}
	e = yaml.Unmarshal([]byte(dat), &t)
	if e != nil {
		debug.WithFields(log.Fields{
			"file":  fileName,
			"error": e,
		}).Fatal("Unable to unmarshal service usage state")
	}
	return t
}

// SaveServiceUsageState stores usage state in a file under Aptomi DB
func (usage ServiceUsageState) SaveServiceUsageState(noop bool) {
	fileName := GetAptomiDBDir() + "/"
	if noop {
		fileName += "db_noop.yaml"
	} else {
		fileName += "db.yaml"
	}

	debug.WithFields(log.Fields{
		"file": fileName,
	}).Info("Saving service usage state")

	e := ioutil.WriteFile(fileName, []byte(serializeObject(usage)), 0644)
	if e != nil {
		debug.WithFields(log.Fields{
			"file":  fileName,
			"error": e,
		}).Fatal("Unable to save service usage state")
	}
}
