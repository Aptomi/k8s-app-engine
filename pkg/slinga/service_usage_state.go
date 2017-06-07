package slinga

import (
	log "github.com/Sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"strings"
)

const componentRootName = "root"

// ServiceUsageState contains resolution data for services - who is using what, as well as contains processing order and additional data
type ServiceUsageState struct {
	// reference to a policy
	Policy *Policy

	// reference to dependencies
	Dependencies *GlobalDependencies

	// resolved triples <service, context, allocation, component> -> list of users & labels
	ResolvedLinks map[string]*ResolvedLinkUsageStruct

	// the order in which components/services have to be processed
	processingOrderHas map[string]bool
	ProcessingOrder    []string

	// travel a path by component names as keys and get to a specific component instance key
	DiscoveryTree NestedParameterMap
}

// ResolvedLinkUsageStruct is a usage data for a given component instance, containing list of user IDs and calculated labels
type ResolvedLinkUsageStruct struct {
	UserIds              []string
	CalculatedLabels     LabelSet
	CalculatedDiscovery  NestedParameterMap
	CalculatedCodeParams NestedParameterMap
}

// NewServiceUsageState creates new empty ServiceUsageState
func NewServiceUsageState(policy *Policy, dependencies *GlobalDependencies) ServiceUsageState {
	return ServiceUsageState{
		Policy:             policy,
		Dependencies:       dependencies,
		ResolvedLinks:      make(map[string]*ResolvedLinkUsageStruct),
		DiscoveryTree:      NestedParameterMap{},
		processingOrderHas: make(map[string]bool)}
}

// Create key for the map
func (usage ServiceUsageState) createServiceUsageKey(service *Service, context *Context, allocation *Allocation, component *ServiceComponent) string {
	var componentName string
	if component != nil {
		componentName = component.Name
	} else {
		componentName = componentRootName
	}
	return usage.createServiceUsageKeyFromStr(service.Name, context.Name, allocation.NameResolved, componentName)
}

// Create key for the map
func (usage ServiceUsageState) createServiceUsageKeyFromStr(serviceName string, contextName string, allocationName string, componentName string) string {
	return serviceName + "#" + contextName + "#" + allocationName + "#" + componentName
}

// ParseServiceUsageKey parses key and returns service, component, allocation, component names
func ParseServiceUsageKey(key string) (string, string, string, string) {
	keyArray := strings.Split(key, "#")
	service := keyArray[0]
	context := keyArray[1]
	allocation := keyArray[2]
	component := keyArray[3]
	return service, context, allocation, component
}

// Create key for the map
func (usage ServiceUsageState) createDependencyKey(serviceName string) string {
	return serviceName
}

// Records usage event
func (usage *ServiceUsageState) recordUsage(key string, user User) string {
	// Add user to the entry
	usageStruct := usage.getComponentInstanceEntry(key)
	usageStruct.UserIds = append(usageStruct.UserIds, user.ID)

	// Add to processing order
	if !usage.processingOrderHas[key] {
		usage.processingOrderHas[key] = true
		usage.ProcessingOrder = append(usage.ProcessingOrder, key)
	}

	return key
}

// Stores calculated discovery params for component instance
func (usage *ServiceUsageState) storeCodeParams(key string, codeParams NestedParameterMap) {
	// TODO: what to do if we came here multiple times with different code params?
	usage.getComponentInstanceEntry(key).CalculatedCodeParams = codeParams
}

// Stores calculated discovery params for component instance
func (usage *ServiceUsageState) storeDiscoveryParams(key string, discoveryParams NestedParameterMap) {
	// TODO: what to do if we came here multiple times with different discovery params?
	usage.getComponentInstanceEntry(key).CalculatedDiscovery = discoveryParams
}

// Stores calculated labels for component instance
func (usage *ServiceUsageState) storeLabels(key string, labels LabelSet) {
	// TODO: we can arrive to a service via multiple usages with different labels. what to do?
	usage.getComponentInstanceEntry(key).CalculatedLabels = labels
}

// Gets a component instance entry or creates an new entry if it doesn't exist
func (usage *ServiceUsageState) getComponentInstanceEntry(key string) *ResolvedLinkUsageStruct {
	if _, ok := usage.ResolvedLinks[key]; !ok {
		usage.ResolvedLinks[key] = &ResolvedLinkUsageStruct{CalculatedLabels: LabelSet{}}
	}
	return usage.ResolvedLinks[key]
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
