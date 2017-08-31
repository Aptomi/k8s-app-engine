package graphviz

import (
	"bytes"
	"fmt"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/diff"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/resolve"
	. "github.com/Aptomi/aptomi/pkg/slinga/util"
	"github.com/awalterschulze/gographviz"
	"image"
	"image/png"
	"os"
	"os/exec"
	"strconv"
)

// See http://www.graphviz.org/doc/info/colors.html
const noEntriesNodeName = "No entries"
const colorScheme = "set19"
const colorCount = 9

// PolicyVisualization accepts diff and defines additional methods for visualizing the policy
type PolicyVisualization struct {
	diff *diff.RevisionDiff
}

// NewPolicyVisualization creates new policy visualization object, given a revision difference
func NewPolicyVisualization(diff *diff.RevisionDiff) PolicyVisualization {
	return PolicyVisualization{diff: diff}
}

func (vis PolicyVisualization) GetImageForRevisionNext() (image.Image, error) {
	nextGraph := makeGraph(vis.diff.Next)
	return vis.getGraphImage(nextGraph)
}

func (vis PolicyVisualization) GetImageForRevisionPrev() (image.Image, error) {
	prevGraph := makeGraph(vis.diff.Prev)
	return vis.getGraphImage(prevGraph)
}

func (vis PolicyVisualization) GetImageForRevisionDiff() (image.Image, error) {
	nextGraph := makeGraph(vis.diff.Next)
	prevGraph := makeGraph(vis.diff.Prev)
	deltaGraph := Delta(prevGraph, nextGraph)
	return vis.getGraphImage(deltaGraph)
}

// Returns a short version of the string
func shorten(s string) string {
	const maxLen = 20
	if len(s) >= maxLen {
		return s[0:maxLen] + "..."
	}
	return s
}

func makeGraph(revision *resolve.Revision) *gographviz.Graph {
	// Write graph into a file
	graph := gographviz.NewGraph()
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
	if revision.Policy.Dependencies != nil {
		for service, dependencies := range revision.Policy.Dependencies.DependenciesByService {
			// Add a node with service
			addNodeOnce(graph, "cluster_Services", service, nil, was)

			// For every user who has a dependency on this service
			for _, d := range dependencies {
				color := getUserColor(d.UserID, colorForUser, &usedColors)

				// Add a node with user
				user := revision.UserLoader.LoadUserByID(d.UserID)
				label := "Name: " + user.Name + " (" + user.ID + ")"
				keys := GetSortedStringKeys(user.Labels)
				for _, k := range keys {
					label += "\n" + k + " = " + shorten(user.Labels[k])
				}
				addNodeOnce(graph, "cluster_Users", d.UserID, map[string]string{"label": label, "style": "filled", "fillcolor": "/" + colorScheme + "/" + strconv.Itoa(color)}, was)

				// Add an edge from user to a service
				addEdge(graph, d.UserID, service, map[string]string{"color": "/" + colorScheme + "/" + strconv.Itoa(color)})
			}
		}
	}

	// Second, visualize evaluated links
	for _, instance := range revision.Resolution.Resolved.ComponentInstanceMap {
		key := instance.Key

		// only add edges to "root" components (i.e. services)
		if !key.IsService() {
			continue
		}

		// Key for service
		serviceAllocationKey := key.ServiceName + "_" + key.ContextNameWithKeys

		// Add a node with service
		addNodeOnce(graph, "cluster_Services", key.ServiceName, nil, was)

		// Add box/subgraph for a given service, containing all its instances
		addSubgraphOnce(graph, "Main", "cluster_Service_Allocations_"+key.ServiceName, map[string]string{"label": "Instances for service: " + key.ServiceName}, was)

		// Add a node with context
		addNodeOnce(graph, "cluster_Service_Allocations_"+key.ServiceName, serviceAllocationKey, map[string]string{"label": "Context: " + key.ContextNameWithKeys}, was)

		// Add an edge from service to service instances box
		for dependencyID := range instance.DependencyIds {
			userID := revision.Policy.Dependencies.DependenciesByID[dependencyID].UserID
			color := getUserColor(userID, colorForUser, &usedColors)
			addEdge(graph, key.ServiceName, serviceAllocationKey, map[string]string{"color": "/" + colorScheme + "/" + strconv.Itoa(color)})
		}
	}

	// Third, show cross-service dependencies
	if revision.Policy != nil {
		for serviceName1, service1 := range revision.Policy.Services {
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
		addNodeOnce(graph, "", noEntriesNodeName, nil, was)
	}

	return graph
}

// Saves graph into a file
func (vis PolicyVisualization) getGraphImage(graph *gographviz.Graph) (image.Image, error) {
	// Original graph in .dot
	fileNameDot := WriteTempFile("graphviz", graph.String())
	defer os.Remove(fileNameDot)

	// Graph with improved layout in .dot
	fileNameDotFlat := fileNameDot + ".flat.dot"
	defer os.Remove(fileNameDotFlat)

	// Graph in .png
	fileNamePng := fileNameDot + ".png"
	defer os.Remove(fileNamePng)

	// Call graphviz to unflatten an image and get a better layout
	{
		cmd := "unflatten"
		args := []string{"-f", "-l", "4", "-o" + fileNameDotFlat, fileNameDot}
		command := exec.Command(cmd, args...)
		var outb, errb bytes.Buffer
		command.Stdout = &outb
		command.Stderr = &errb
		if err := command.Run(); err != nil || len(errb.String()) > 0 {
			panic(fmt.Sprintf("Unable to execute graphviz '%s' with '%s': %s %s %s", cmd, args, outb.String(), errb.String(), err.Error()))
		}
	}

	// Call graphviz to generate an image
	{
		cmd := "dot"
		args := []string{"-Tpng", "-o" + fileNamePng, fileNameDotFlat}
		// args := []string{"-Tpng", "-Kfdp", "-o" + fileNamePng, fileNameDotFlat}
		command := exec.Command(cmd, args...)
		var outb, errb bytes.Buffer
		command.Stdout = &outb
		command.Stderr = &errb
		if err := command.Run(); err != nil || len(errb.String()) > 0 {
			panic(fmt.Sprintf("Unable to execute graphviz '%s' with '%s': %s %s %s", cmd, args, outb.String(), errb.String(), err.Error()))
		}
	}

	// Read image and return
	filePng, err := os.Open(fileNamePng)
	if err != nil {
		panic(fmt.Sprintf("Unable to load PNG generated by graphviz '%s': %s", fileNamePng, err.Error()))
	}
	defer filePng.Close()
	return png.Decode(filePng)
}

// Returns a color for the given user
func getUserColor(userID string, colorForUser map[string]int, usedColors *int) int {
	color, ok := colorForUser[userID]
	if !ok {
		*usedColors++
		if *usedColors > colorCount {
			*usedColors = 1
		}
		color = *usedColors
		colorForUser[userID] = color
	}
	return color
}

// Adds an edge if it doesn't exist already
func addEdge(g *gographviz.Graph, src string, dst string, attrs map[string]string) {
	g.AddEdge(esc(src), esc(dst), true, escAttrs(attrs))
}

// Adds a subgraph if it doesn't exist already
func addSubgraphOnce(g *gographviz.Graph, parentGraph string, name string, attrs map[string]string, was map[string]bool) {
	wasKey := "SUBGRAPH" + "#" + parentGraph + "#" + name
	if !was[wasKey] {
		g.AddSubGraph(esc(parentGraph), esc(name), escAttrs(attrs))
		was[wasKey] = true
	}
}

// Adds a node if it doesn't exist already
func addNodeOnce(g *gographviz.Graph, parentGraph string, name string, attrs map[string]string, was map[string]bool) {
	wasKey := "NODE" + "#" + parentGraph + "#" + name
	if !was[wasKey] {
		g.AddNode(esc(parentGraph), esc(name), escAttrs(attrs))
		was[wasKey] = true
	}
}
