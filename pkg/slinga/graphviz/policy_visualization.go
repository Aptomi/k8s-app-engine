package graphviz

import (
	"github.com/Aptomi/aptomi/pkg/slinga/engine/diff"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/slinga/external"
	"github.com/Aptomi/aptomi/pkg/slinga/language"
	. "github.com/Aptomi/aptomi/pkg/slinga/util"
	"github.com/awalterschulze/gographviz"
	"image"
	"strconv"
)

// See http://www.graphviz.org/doc/info/colors.html
const noEntriesNodeName = "No entries"
const colorScheme = "set19"
const colorCount = 9

// PolicyVisualization accepts diff and defines additional methods for visualizing the policy
type PolicyVisualization struct {
	diff *diff.PolicyResolutionDiff
}

// NewPolicyVisualizationImage returns an image with policy/resolution information
func NewPolicyVisualizationImage(policy *language.Policy, resolution *resolve.PolicyResolution, externalData *external.Data) (image.Image, error) {
	graph := makeGraph(policy, resolution, externalData)
	return getGraphImage(graph)
}

// NewPolicyVisualizationDeltaImage returns an image with policy/resolution information
func NewPolicyVisualizationDeltaImage(nextPolicy *language.Policy, nextResolution *resolve.PolicyResolution, prevPolicy *language.Policy, prevResolution *resolve.PolicyResolution, externalData *external.Data) (image.Image, error) {
	nextGraph := makeGraph(nextPolicy, nextResolution, externalData)
	prevGraph := makeGraph(prevPolicy, prevResolution, externalData)
	deltaGraph := Delta(prevGraph, nextGraph)
	return getGraphImage(deltaGraph)
}

func makeGraph(policy *language.Policy, resolution *resolve.PolicyResolution, externalData *external.Data) *gographviz.Graph {
	// Write graph into a file
	graph := gographviz.NewGraph()
	graph.SetName("Main")
	graph.AddAttr("Main", "compound", "true")
	graph.SetDir(true)

	was := make(map[string]bool)

	// Add box/subgraph for users
	addSubgraphOnce(graph, "Main", "cluster_Users", map[string]string{"label": "Users"}, was)

	// Add box/subgraph for contracts
	addSubgraphOnce(graph, "Main", "cluster_Contracts", map[string]string{"label": "Contracts"}, was)

	// Add box/subgraph for services
	addSubgraphOnce(graph, "Main", "cluster_Services", map[string]string{"label": "Services"}, was)

	// How many colors have been used
	usedColors := 0
	colorForUser := make(map[string]int)

	// First of all, let's show all dependencies (who requested what)
	if policy.Dependencies != nil {
		for contractName, dependencies := range policy.Dependencies.DependenciesByContract {
			// Add a node with contract
			addNodeOnce(graph, "cluster_Contracts", contractName, nil, was)

			// For every user who has a dependency on this service
			for _, d := range dependencies {
				color := getUserColor(d.UserID, colorForUser, &usedColors)

				// Add a node with user
				user := externalData.UserLoader.LoadUserByID(d.UserID)
				label := "Name: " + user.Name + " (" + user.ID + ")"
				keys := GetSortedStringKeys(user.Labels)
				for _, k := range keys {
					label += "\n" + k + " = " + shorten(user.Labels[k])
				}
				addNodeOnce(graph, "cluster_Users", d.UserID, map[string]string{"label": label, "style": "filled", "fillcolor": "/" + colorScheme + "/" + strconv.Itoa(color)}, was)

				// Add an edge from user to a contract
				addEdge(graph, d.UserID, contractName, map[string]string{"color": "/" + colorScheme + "/" + strconv.Itoa(color)})
			}
		}
	}

	// Second, visualize evaluated links
	for _, instance := range resolution.ComponentInstanceMap {
		key := instance.Metadata.Key

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
			userID := policy.Dependencies.DependenciesByID[dependencyID].UserID
			color := getUserColor(userID, colorForUser, &usedColors)
			addEdge(graph, key.ServiceName, serviceAllocationKey, map[string]string{"color": "/" + colorScheme + "/" + strconv.Itoa(color)})
		}
	}

	// Third, show service-contract dependencies
	if policy != nil {
		for serviceName1, service1 := range policy.Services {
			// Resolve every component
			for _, component := range service1.Components {
				contractName2 := component.Contract
				if contractName2 != "" {
					// Add a node with service1
					addNodeOnce(graph, "cluster_Services", serviceName1, nil, was)

					// Add a node with service2
					addNodeOnce(graph, "cluster_Contracts", contractName2, nil, was)

					// Show dependency
					addEdge(graph, serviceName1, contractName2, map[string]string{"color": "gray60"})
				}
			}
		}
	} else {
		addNodeOnce(graph, "", noEntriesNodeName, nil, was)
	}

	return graph
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
