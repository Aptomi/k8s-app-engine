package slinga

import (
	"bytes"
	"github.com/awalterschulze/gographviz"
	"github.com/golang/glog"
	"io/ioutil"
	"os/exec"
	"strconv"
	"strings"
	"sort"
)

// See http://www.graphviz.org/doc/info/colors.html
const colorScheme = "set19"
const colorCount = 9

// Returns name of the file where visual is stored
func (usage ServiceUsageState) GetVisualFileNamePNG() string {
	return GetAptomiDBDir() + "/" + "graph.png"
}

// Stores usage state visual into a file
func (usage ServiceUsageState) DrawVisualAndStore() {
	users := LoadUsersFromDir(GetAptomiPolicyDir())

	// Write graph into a file
	graph := gographviz.NewEscape()
	graph.SetName("Main")
	graph.AddAttr("Main", "compound", "true")
	graph.SetDir(true)

	was := make(map[string]bool)

	// Add box/subgraph for users
	addSubgraphOnce(graph, "Main", "cluster_Users", map[string]string{"label": "Users"}, was)

	// Add box/subgraph for services
	addSubgraphOnce(graph, "Main", "cluster_Services", map[string]string{"label": "Services"}, was)

	// How many colors have been used
	usedColors := 0
	colorForUser := make(map[string]int)

	// First of all, let's show all dependencies (who requested what)
	if usage.Dependencies != nil {
		for service, dependencies := range usage.Dependencies.Dependencies {
			// Add a node with service
			addNodeOnce(graph, "cluster_Services", service, nil, was)

			// For every user who has a dependency on this service
			for _, d := range dependencies {
				color := getUserColor(d.UserId, colorForUser, &usedColors)

				// Add a node with user
				user := users.Users[d.UserId]
				label := "Name: " + user.Name + " (" + user.Id + ")"
				var keys []string
				for k := range user.Labels {
					keys = append(keys, k)
				}
				sort.Strings(keys)
				for _, k := range keys {
					label += "\n" + k + " = " + user.Labels[k]
				}
				addNodeOnce(graph, "cluster_Users", d.UserId, map[string]string{"label": label, "style": "filled", "fillcolor": "/" + colorScheme + "/" + strconv.Itoa(color)}, was)

				// Add an edge from user to a service
				addEdge(graph, d.UserId, service, map[string]string{"color": "/" + colorScheme + "/" + strconv.Itoa(color)})
			}
		}
	}

	// Second, visualize evaluated links
	for key, linkStruct := range usage.ResolvedLinks {
		keyArray := strings.Split(key, "#")
		service := keyArray[0]
		contextAndAllocation := keyArray[1] + "#" + keyArray[2]
		component := keyArray[3]

		// only add edges to "root" components (i.e. services)
		if component != componentRootName {
			continue
		}

		// Key for allocation
		serviceAllocationKey := service + "_" + contextAndAllocation

		// Add a node with service
		addNodeOnce(graph, "cluster_Services", service, nil, was)

		// Add box/subgraph for a given service, containing all its allocations
		addSubgraphOnce(graph, "Main", "cluster_Service_Allocations_" + service, map[string]string{"label": "Allocations for service: " + service}, was)

		// Add a node with allocation
		addNodeOnce(graph, "cluster_Service_Allocations_" + service, serviceAllocationKey, map[string]string{"label": "Context: " + keyArray[1] + "\n" + "Allocation: " + keyArray[2]}, was)

		// Add an edge from service to allocation box
		for _, userId := range linkStruct.UserIds {
			color := getUserColor(userId, colorForUser, &usedColors)
			addEdge(graph, service, serviceAllocationKey, map[string]string{"color": "/" + colorScheme + "/" + strconv.Itoa(color)})
		}
	}

	// Third, show cross-service dependencies
	if usage.Policy != nil {
		for serviceName1, service1 := range usage.Policy.Services {
			// Resolve every component
			for _, component := range service1.Components {
				serviceName2 := component.Service
				if serviceName2 != "" {
					// Add a node with service1
					addNodeOnce(graph, "cluster_Services", serviceName1, nil, was)

					// Add a node with service2
					addNodeOnce(graph, "cluster_Services", serviceName2, nil, was)

					// Show dependency
					addEdge(graph, serviceName1, serviceName2, map[string]string{"color": "gray60"})
				}
			}
		}
	} else {
		addNodeOnce(graph, "", "Empty", nil, was)
	}

	fileNameDot := GetAptomiDBDir() + "/" + "graph_full.dot"
	fileNameDotFlat := GetAptomiDBDir() + "/" + "graph_flat.dot"
	err := ioutil.WriteFile(fileNameDot, []byte(graph.String()), 0644)
	if err != nil {
		glog.Fatalf("Unable to write to a file: %s", fileNameDot)
	}

	// Call graphviz to flatten an image
	{
		cmd := "unflatten"
		args := []string{"-f", "-l", "4", "-o" + fileNameDotFlat, fileNameDot}
		command := exec.Command(cmd, args...)
		var outb, errb bytes.Buffer
		command.Stdout = &outb
		command.Stderr = &errb
		if err := command.Run(); err != nil {
			glog.Fatalf("Unable to execute graphviz (%s): %s %s", cmd, outb.String(), errb.String(), err)
		}
	}

	// Call graphviz to generate an image
	{
		// -Kfdp will call a different engine
		cmd := "dot"
		args := []string{"-Tpng", "-o" + usage.GetVisualFileNamePNG(), fileNameDotFlat}
		command := exec.Command(cmd, args...)
		var outb, errb bytes.Buffer
		command.Stdout = &outb
		command.Stderr = &errb
		if err := command.Run(); err != nil {
			glog.Fatalf("Unable to execute graphviz (%s): %s %s", cmd, outb.String(), errb.String(), err)
		}
	}
}

func getUserColor(userId string, colorForUser map[string]int, usedColors *int) int {
	color, ok := colorForUser[userId]
	if !ok {
		*usedColors++
		if *usedColors > colorCount {
			*usedColors = 1
		}
		color = *usedColors
		colorForUser[userId] = color
	}
	return color
}

// Adds an edge if it doesn't exist already
func addEdge(g *gographviz.Escape, src string, dst string, attrs map[string]string) {
	g.AddEdge(src, dst, true, attrs)
}

// Adds a subgraph if it doesn't exist already
func addSubgraphOnce(g *gographviz.Escape, parentGraph string, name string, attrs map[string]string, was map[string]bool) {
	wasKey := "SUBGRAPH" + "#" + parentGraph + "#" + name
	if !was[wasKey] {
		g.AddSubGraph(parentGraph, name, attrs)
		was[wasKey] = true
	}
}

// Adds a node if it doesn't exist already
func addNodeOnce(g *gographviz.Escape, parentGraph string, name string, attrs map[string]string, was map[string]bool) {
	wasKey := "NODE" + "#" + parentGraph + "#" + name
	if !was[wasKey] {
		g.AddNode(parentGraph, name, attrs)
		was[wasKey] = true
	}
}
