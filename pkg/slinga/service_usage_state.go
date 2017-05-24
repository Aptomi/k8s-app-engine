package slinga

import (
	"github.com/golang/glog"
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
	ProcessingOrder []string

	// map from service instance key to map from component name to component instance key
	ComponentInstanceMap map[string]interface{}

	// tracing - gets populated with detailed debug information if tracing is requested
	tracing *ServiceUsageTracing
}

// ResolvedLinkUsageStruct is a usage data for a given component instance, containing list of user IDs and calculated labels
type ResolvedLinkUsageStruct struct {
	UserIds               []string
	CalculatedLabels      LabelSet
	CalculatedCodeContent map[string]map[string]string
}

// NewServiceUsageState creates new empty ServiceUsageState
func NewServiceUsageState(policy *Policy, dependencies *GlobalDependencies) ServiceUsageState {
	return ServiceUsageState{
		Policy:               policy,
		Dependencies:         dependencies,
		ResolvedLinks:        make(map[string]*ResolvedLinkUsageStruct),
		ComponentInstanceMap: make(map[string]interface{}),
		tracing:              NewServiceUsageTracing()}
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
func (usage *ServiceUsageState) recordUsage(user User, service *Service, context *Context, allocation *Allocation, component *ServiceComponent, labels LabelSet, codeContent map[string]map[string]string) string {
	key := usage.createServiceUsageKey(service, context, allocation, component)

	if _, ok := usage.ResolvedLinks[key]; !ok {
		usage.ResolvedLinks[key] = &ResolvedLinkUsageStruct{CalculatedLabels: LabelSet{}, CalculatedCodeContent: make(map[string]map[string]string)}
	}

	usage.ResolvedLinks[key].appendToLinkUsageStruct(user.ID, labels, codeContent)
	usage.ProcessingOrder = append(usage.ProcessingOrder, key)

	return key
}

// Adds user and set of labels to the entry
func (usageStruct *ResolvedLinkUsageStruct) appendToLinkUsageStruct(userID string, labels LabelSet, codeContent map[string]map[string]string) {
	usageStruct.UserIds = append(usageStruct.UserIds, userID)

	// TODO: we can arrive to a service via multiple usages with different labels. what to do?
	usageStruct.CalculatedLabels = labels

	// TODO: what to do with different code contents? they should be the same
	usageStruct.CalculatedCodeContent = codeContent
}

// LoadServiceUsageState loads usage state from a file under Aptomi DB
func LoadServiceUsageState() ServiceUsageState {
	fileName := GetAptomiDBDir() + "/" + "db.yaml"

	dat, e := ioutil.ReadFile(fileName)

	// If the file doesn't exist, it means that DB is empty and we are starting from scratch
	if os.IsNotExist(e) {
		return ServiceUsageState{}
	} else if e != nil {
		glog.Fatalf("Unable to read file: %v", e)
	}

	if e != nil {
		glog.Fatalf("Unable to read file: %v", e)
	}
	t := ServiceUsageState{}
	e = yaml.Unmarshal([]byte(dat), &t)
	if e != nil {
		glog.Fatalf("Unable to unmarshal service usage state: %v", e)
	}
	return t
}

// SaveServiceUsageState stores usage state in a file under Aptomi DB
func (usage ServiceUsageState) SaveServiceUsageState() {
	fileName := GetAptomiDBDir() + "/" + "db.yaml"
	err := ioutil.WriteFile(fileName, []byte(serializeObject(usage)), 0644)
	if err != nil {
		glog.Fatalf("Unable to write to a file: %s", fileName)
	}
}
