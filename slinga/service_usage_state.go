package slinga

import (
	"os"
	"log"
	"io/ioutil"
	"gopkg.in/yaml.v2"
)

// Service structure - who is currently using what
type ServiceUsageState struct {
	// recorded initial dependencies <service> -> list of users
	Dependencies map[string][]string

	// resolved triples <service, context, allocation, component> -> list of users
	ResolvedLinks map[string][]string

	// the order in which components/services have to be instantiated
	ProcessingOrder []string
}

func NewServiceUsageState() ServiceUsageState {
	return ServiceUsageState{Dependencies: make(map[string][]string), ResolvedLinks: make(map[string][]string)}
}

// Create key for the map
func (state ServiceUsageState) createUsageKey(service *Service, context *Context, allocation *Allocation, component *ServiceComponent) string {
	var componentName string
	if component != nil {
		componentName = component.Name
	} else {
		componentName = "<root>"
	}
	return service.Name + "#" + context.Name + "#" + allocation.NameResolved + "#" + componentName
}

// Create key for the map
func (state ServiceUsageState) createDependencyKey(serviceName string) string {
	return serviceName
}

// Records usage event
func (state *ServiceUsageState) recordUsage(user User, service *Service, context *Context, allocation *Allocation, component *ServiceComponent) {
	key := state.createUsageKey(service, context, allocation, component)
	state.ResolvedLinks[key] = append(state.ResolvedLinks[key], user.Id)
	state.ProcessingOrder = append(state.ProcessingOrder, key)
}

// Records requested dependency
func (state *ServiceUsageState) recordDependency(user User, serviceName string) {
	key := state.createDependencyKey(serviceName)
	state.Dependencies[key] = append(state.Dependencies[key], user.Id)
}

// Stores usage state in a file
func loadServiceUsageState() ServiceUsageState {
	aptomiDB, ok := os.LookupEnv("APTOMI_DB")
	if !ok {
		log.Fatal("Attempting to load state from disk, but APTOMI_DB environment variable is not present. Must point to a directory")
	}
	if stat, err := os.Stat(aptomiDB); err != nil || !stat.IsDir() {
		log.Fatal("Directory APTOMI_DB doesn't exist: " + aptomiDB)
	}
	fileName := aptomiDB + "/" + "db.yaml"
	dat, e := ioutil.ReadFile(fileName)
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
func (state ServiceUsageState) storeServiceUsageState() {
	aptomiDB, ok := os.LookupEnv("APTOMI_DB")
	if !ok {
		log.Fatal("Attempting to write state on disk, but APTOMI_DB environment variable is not present. Must point to a directory")
	}
	if stat, err := os.Stat(aptomiDB); err != nil || !stat.IsDir() {
		log.Fatal("Directory APTOMI_DB doesn't exist: " + aptomiDB)
	}
	fileName := aptomiDB + "/" + "db.yaml"
	err := ioutil.WriteFile(fileName, []byte(serializeObject(state)), 0644);
	if err != nil {
		log.Fatal("Unable to write to a file: " + fileName)
	}
}
