package slinga

import (
	"os"
	"log"
	"io/ioutil"
	"gopkg.in/yaml.v2"
	"github.com/awalterschulze/gographviz"
	"os/exec"
	"strings"
	"bytes"
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

// Return aptomi DB directory
func getAptomiDB() string {
	aptomiDB, ok := os.LookupEnv("APTOMI_DB")
	if !ok {
		log.Fatal("Attempting to load/save state from disk, but APTOMI_DB environment variable is not present. Must point to a directory")
	}
	if stat, err := os.Stat(aptomiDB); err != nil || !stat.IsDir() {
		log.Fatal("Directory APTOMI_DB doesn't exist: " + aptomiDB)
	}
	return aptomiDB
}

// Stores usage state in a file
func loadServiceUsageState() ServiceUsageState {
	fileName := getAptomiDB() + "/" + "db.yaml"
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
func (state ServiceUsageState) saveServiceUsageState() {
	fileName := getAptomiDB() + "/" + "db.yaml"
	err := ioutil.WriteFile(fileName, []byte(serializeObject(state)), 0644);
	if err != nil {
		log.Fatal("Unable to write to a file: " + fileName)
	}

	state.storeServiceUsageStateVisual()
}

// Stores usage state visual into a file
func (state ServiceUsageState) storeServiceUsageStateVisual() {

	// Write graph into a file
	graph := gographviz.NewEscape()
	graph.SetName("Main")
	graph.SetDir(true)
	graph.AddSubGraph("Main", "cluster_Services", map[string]string{"label": "Services"})
	graph.AddSubGraph("Main", "cluster_ContextsAndAllocations", map[string]string{"label": "Contexts_Allocations"})
	graph.AddSubGraph("Main", "cluster_Components", map[string]string{"label": "Components"})
	graph.AddSubGraph("Main", "cluster_Users", map[string]string{"label": "Users"})

	was := make(map[string]bool)
	for key, userIds := range state.ResolvedLinks {
		keyArray := strings.Split(key, "#")
		service := keyArray[0]
		contextAndAllocation := keyArray[1] + "-" + keyArray[2]
		component := keyArray[3]

		graph.AddNode("cluster_Services", service, nil)
		graph.AddNode("cluster_ContextsAndAllocations", contextAndAllocation, nil)
		graph.AddNode("cluster_Components", component, nil)

		addEdgeOnce(graph, service, contextAndAllocation, true, nil, "SC", was)
		addEdgeOnce(graph, contextAndAllocation, component, true, nil, "CA", was)

		for _, userId := range userIds {
			graph.AddNode("cluster_Users", userId, nil)
			addEdgeOnce(graph, userId, service, true, nil, "US", was)
			addEdgeOnce(graph, userId, contextAndAllocation, true, nil, "UCA", was)
			addEdgeOnce(graph, userId, component, true, nil, "UC", was)
		}
	}

	/*
	graph.AddNode("cluster_Services", "a", nil)
	graph.AddNode("cluster_Services", "b", nil)
	graph.AddEdge("a", "b", true, map[string]string{"color": "red"})

	graph.AddNode("cluster_Components", "c", nil)
	graph.AddNode("cluster_Components", "d", nil)
	graph.AddEdge("c", "d", true, map[string]string{"color": "red"})

	graph.AddEdge("a", "c", true, map[string]string{"color": "green"})
	graph.AddEdge("b", "d", true, map[string]string{"color": "green"})
	*/

	fileNameDot := getAptomiDB() + "/" + "db.dot"
	err := ioutil.WriteFile(fileNameDot, []byte(graph.String()), 0644);
	if err != nil {
		log.Fatal("Unable to write to a file: " + fileNameDot)
	}

	// Call graphviz to generate an image
	cmd := "dot"
	fileNamePng := getAptomiDB() + "/" + "db.png"
	args := []string{"-Tpng", "-o" + fileNamePng, fileNameDot}
	command := exec.Command(cmd, args...)
	var outb, errb bytes.Buffer
	command.Stdout = &outb
	command.Stderr = &errb
	if err := command.Run(); err != nil {
		log.Fatal("Unable to execute graphviz: "+outb.String()+" "+errb.String(), err)
	}
}

// AddEdge adds an edge and escapes the src, dst and attrs, if needed.
func addEdgeOnce(g *gographviz.Escape, src string, dst string, directed bool, attrs map[string]string, keyPrefix string, was map[string]bool) {
	wasKey := keyPrefix + "#" + src + "#" + dst;
	if !was[wasKey] {
		g.AddEdge(src, dst, directed, attrs)
		was[wasKey] = true
	}
}

