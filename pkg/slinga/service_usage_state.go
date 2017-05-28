package slinga

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"strings"
	"github.com/Sirupsen/logrus"
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
	ProcessingOrder []string

	// map from service instance key to map from component name to component instance key
	ComponentInstanceMap map[string]interface{}
}

// ResolvedLinkUsageStruct is a usage data for a given component instance, containing list of user IDs and calculated labels
type ResolvedLinkUsageStruct struct {
	UserIds              []string
	CalculatedLabels     LabelSet
	CalculatedDiscovery  interface{}
	CalculatedCodeParams interface{}
}

// NewServiceUsageState creates new empty ServiceUsageState
func NewServiceUsageState(policy *Policy, dependencies *GlobalDependencies) ServiceUsageState {
	return ServiceUsageState{
		Policy:               policy,
		Dependencies:         dependencies,
		ResolvedLinks:        make(map[string]*ResolvedLinkUsageStruct),
		ComponentInstanceMap: make(map[string]interface{})}
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

// Parse key
func parseServiceUsageKey(key string) (string, string, string, string) {
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
func (usage *ServiceUsageState) recordUsage(key string, user User, labels LabelSet, codeParams interface{}, discoveryParams interface{}) string {
	if _, ok := usage.ResolvedLinks[key]; !ok {
		usage.ResolvedLinks[key] = &ResolvedLinkUsageStruct{CalculatedLabels: LabelSet{}}
	}

	usage.ResolvedLinks[key].appendToLinkUsageStruct(user.ID, labels, codeParams, discoveryParams)
	usage.ProcessingOrder = append(usage.ProcessingOrder, key)

	return key
}

// Adds user and set of labels to the entry
func (usageStruct *ResolvedLinkUsageStruct) appendToLinkUsageStruct(userID string, labels LabelSet, codeParams interface{}, discoveryParams interface{}) {
	usageStruct.UserIds = append(usageStruct.UserIds, userID)

	// TODO: we can arrive to a service via multiple usages with different labels. what to do?
	usageStruct.CalculatedLabels = labels

	// TODO: what to do with different code contents?
	usageStruct.CalculatedCodeParams = codeParams

	// TODO: what to do with different discovery contents?
	usageStruct.CalculatedDiscovery = discoveryParams
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
			"file": fileName,
			"error": e,
		}).Fatal("Unable to read file", e)
	}

	t := ServiceUsageState{}
	e = yaml.Unmarshal([]byte(dat), &t)
	if e != nil {
		debug.WithFields(log.Fields{
			"file": fileName,
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
			"file": fileName,
			"error": e,
		}).Fatal("Unable to save service usage state")
	}
}
