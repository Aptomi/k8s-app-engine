package slinga

import (
	"log"
	"io/ioutil"
	"gopkg.in/yaml.v2"
	"strings"
	"os"
)

const componentRootName = "root"

// Service structure - who is currently using what
type ServiceUsageState struct {
	// reference to a policy
	Policy *Policy

	// reference to dependencies
	Dependencies *GlobalDependencies

	// resolved triples <service, context, allocation, component> -> list of users
	ResolvedLinks map[string][]string

	// the order in which components/services have to be instantiated
	ProcessingOrder []string
}

func NewServiceUsageState(policy *Policy, dependencies *GlobalDependencies) ServiceUsageState {
	return ServiceUsageState{
		Policy: policy,
		Dependencies: dependencies,
		ResolvedLinks: make(map[string][]string)}
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
func (usage *ServiceUsageState) recordUsage(user User, service *Service, context *Context, allocation *Allocation, component *ServiceComponent) {
	key := usage.createServiceUsageKey(service, context, allocation, component)
	usage.ResolvedLinks[key] = append(usage.ResolvedLinks[key], user.Id)
	usage.ProcessingOrder = append(usage.ProcessingOrder, key)
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
		log.Fatalf("Unable to read file: %v", e)
	}

	if e != nil {
		log.Fatalf("Unable to read file: %v", e)
	}
	t := ServiceUsageState{}
	e = yaml.Unmarshal([]byte(dat), &t)
	if e != nil {
		log.Fatalf("Unable to unmarshal service usage state: %v", e)
	}
	return t
}

// Stores usage state in a file
func (usage ServiceUsageState) SaveServiceUsageState() {
	fileName := GetAptomiDBDir() + "/" + "db.yaml"
	err := ioutil.WriteFile(fileName, []byte(serializeObject(usage)), 0644);
	if err != nil {
		log.Fatal("Unable to write to a file: " + fileName)
	}
}
