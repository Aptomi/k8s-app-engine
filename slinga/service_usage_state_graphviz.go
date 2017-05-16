package slinga

import (
	"log"
	"io/ioutil"
	"github.com/awalterschulze/gographviz"
	"os/exec"
	"strings"
	"bytes"
	"strconv"
)

// See http://www.graphviz.org/doc/info/colors.html
const colorScheme = "set19"
const colorCount = 9

// Stores usage state visual into a file
func (usage ServiceUsageState) storeServiceUsageStateVisual() {

	// Write graph into a file
	graph := gographviz.NewEscape()
	graph.SetName("Main")
	graph.AddAttr("Main", "compound", "true")
	graph.SetDir(true)

	was := make(map[string]bool)

	// Add box/subgraph for users
	addSubgraphOnce(graph, "Main", "cluster_Users", map[string]string{"label": "Users"}, was)

	// Add box/subgraph for services
	addSubgraphOnce(graph, "Main", "cluster_Services", map[string]string{"label": "Services"}, was);

	// How many colors have been used
	usedColors := 0
	colorForUser := make(map[string]int)

	// First of all, let's show all dependencies (who requested what)
	for service, userIds := range usage.Dependencies.Dependencies {
		// Add a node with service
		addNodeOnce(graph, "cluster_Services", service, nil, was)

		// For every user who has a dependency on this service
		for _, userId := range userIds {
			color := getUserColor(userId, colorForUser, &usedColors)

			// Add a node with user
			addNodeOnce(graph, "cluster_Users", userId, map[string]string{"style": "filled", "fillcolor": "/" + colorScheme + "/" + strconv.Itoa(color)}, was)

			// Add an edge from user to a service
			addEdge(graph, userId, service, map[string]string{"color": "/" + colorScheme + "/" + strconv.Itoa(color)})
		}
	}

	// Second, visualize evaluated links
	for key, userIds := range usage.ResolvedLinks {
		keyArray := strings.Split(key, "#")
		service := keyArray[0]
		contextAndAllocation := keyArray[1] + "#" + keyArray[2]
		component := keyArray[3]

		componentKey := service + "_" + contextAndAllocation + "_" + component
		componentLabel := service + "_" + component

		// Add box/subgraph for a given context/allocation
		addSubgraphOnce(graph, "Main", "cluster_"+contextAndAllocation, map[string]string{"label": "Context/Allocation: " + contextAndAllocation}, was)

		// Add a node with service
		addNodeOnce(graph, "cluster_Services", service, nil, was)

		// Add a node with component
		addNodeOnce(graph, "cluster_"+contextAndAllocation, componentKey, map[string]string{"label": componentLabel}, was)

		// Add an edge from service to allocation box
		for _, userId := range userIds {
			color := getUserColor(userId, colorForUser, &usedColors)
			addEdge(graph, service, componentKey, map[string]string{"color": "/" + colorScheme + "/" + strconv.Itoa(color)})
		}
	}

	// Third, show cross-service dependencies
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

	fileNameDot := GetAptomiDB() + "/" + "graph_full.dot"
	fileNameDotFlat := GetAptomiDB() + "/" + "graph_flat.dot"
	err := ioutil.WriteFile(fileNameDot, []byte(graph.String()), 0644);
	if err != nil {
		log.Fatal("Unable to write to a file: " + fileNameDot)
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
			log.Fatal("Unable to execute graphviz: "+outb.String()+" "+errb.String(), err)
		}
	}

	// Call graphviz to generate an image
	{
		cmd := "dot"
		fileNamePng := GetAptomiDB() + "/" + "graph.png"
		args := []string{"-Tpng", "-o" + fileNamePng, fileNameDotFlat}
		command := exec.Command(cmd, args...)
		var outb, errb bytes.Buffer
		command.Stdout = &outb
		command.Stderr = &errb
		if err := command.Run(); err != nil {
			log.Fatal("Unable to execute graphviz: "+outb.String()+" "+errb.String(), err)
		}
	}
}
func getUserColor(userId string, colorForUser map[string]int, usedColors *int) int {
	color, ok := colorForUser[userId]
	if !ok {
		*usedColors++;
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
	wasKey := "SUBGRAPH" + "#" + parentGraph + "#" + name;
	if !was[wasKey] {
		g.AddSubGraph(parentGraph, name, attrs)
		was[wasKey] = true
	}
}

// Adds a node if it doesn't exist already
func addNodeOnce(g *gographviz.Escape, parentGraph string, name string, attrs map[string]string, was map[string]bool) {
	wasKey := "NODE" + "#" + parentGraph + "#" + name;
	if !was[wasKey] {
		g.AddNode(parentGraph, name, attrs)
		was[wasKey] = true
	}
}
