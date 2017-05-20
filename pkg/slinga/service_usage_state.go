package slinga

import (
	"github.com/golang/glog"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"strings"
)

const componentRootName = "root"

// Service structure - who is currently using what
type ServiceUsageState struct {
	// reference to a policy
	Policy *Policy

	// reference to dependencies
	Dependencies *GlobalDependencies

	// resolved triples <service, context, allocation, component> -> list of users
	ResolvedLinks map[string]*ResolvedLinkUsageStruct

	// the order in which components/services have to be processed
	ProcessingOrder []string

	// map from service instance key to map from component name to component instance key
	ComponentInstanceMap map[string]map[string]string
}

type ResolvedLinkUsageStruct struct {
	UserIds          []string
	CalculatedLabels LabelSet
}

func NewServiceUsageState(policy *Policy, dependencies *GlobalDependencies) ServiceUsageState {
	return ServiceUsageState{
		Policy:        policy,
		Dependencies:  dependencies,
		ResolvedLinks: make(map[string]*ResolvedLinkUsageStruct),
		ComponentInstanceMap: make(map[string]map[string]string)}
}

// Create key for the map
func (usage ServiceUsageState) createServiceUsageKey(service *Service, context *Context, allocation *Allocation, component *ServiceComponent) string {
	var componentName string
	if component != nil {
		componentName = component.Name
	} else {
		componentName = componentRootName
	}
	return service.Name + "#" + context.Name + "#" + allocation.NameResolved + "#" + componentName
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
func (usage *ServiceUsageState) recordUsage(user User, service *Service, context *Context, allocation *Allocation, component *ServiceComponent, labels LabelSet) string {
	key := usage.createServiceUsageKey(service, context, allocation, component)

	if _, ok := usage.ResolvedLinks[key]; !ok {
		usage.ResolvedLinks[key] = &ResolvedLinkUsageStruct{CalculatedLabels: LabelSet{}}
	}
	usage.ResolvedLinks[key].append(user.Id, labels)
	usage.ProcessingOrder = append(usage.ProcessingOrder, key)

	return key
}

// Adds user and set of labels to the entry
func (usageStruct *ResolvedLinkUsageStruct) append(userId string, labelSet LabelSet) {
	usageStruct.UserIds = append(usageStruct.UserIds, userId)

	// TODO: we can arrive to a service via multiple usages with different labels. what to do?
	usageStruct.CalculatedLabels = labelSet
}

// Records requested dependency
func (usage *ServiceUsageState) addDependency(user User, serviceName string) {
	key := usage.createDependencyKey(serviceName)
	usage.Dependencies.Dependencies[key] = append(usage.Dependencies.Dependencies[key], user.Id)
}

// Stores usage state in a file
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

// Stores usage state in a file
func (usage ServiceUsageState) SaveServiceUsageState() {
	fileName := GetAptomiDBDir() + "/" + "db.yaml"
	err := ioutil.WriteFile(fileName, []byte(serializeObject(usage)), 0644)
	if err != nil {
		glog.Fatalf("Unable to write to a file: %s", fileName)
	}
}
