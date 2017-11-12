package visualization

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

type graphNode interface {
	getID() string
	getLabel() string
	getGroup() string
}

type graphEdge interface {
	getSrc() graphNode
	getDst() graphNode
	getLabel() string
}

// Graph is a data structure for visualizing Aptomi objects in a form of a graph with nodes and edges.
// It gets created by GraphBuilder. Once created, use GetDataJSON() to get the result and feed into vis.js.
type Graph struct {
	hasObject map[string]graphEntry
	nodes     graphEntryList
	edges     graphEntryList
}

// Creates a new graph
func newGraph() *Graph {
	return &Graph{
		hasObject: make(map[string]graphEntry),
		nodes:     graphEntryList{},
		edges:     graphEntryList{},
	}
}

func idEscape(id string) string {
	return strings.NewReplacer("#", "_", ":", "_").Replace(id)
}

// GetData returns a struct, which can be fed directly into vis.js as network map
func (g *Graph) GetData() interface{} {
	// Sort nodes and edges, so we can get a stable response from API that doesn't change over reloads
	// This will ensure that UI will show the same layout over refreshes
	sort.Sort(g.nodes)
	sort.Sort(g.edges)

	// Wrap it into a graph structure for visjs
	return graphEntry{
		"nodes": g.nodes,
		"edges": g.edges,
	}
}

// GetDataJSON returns a JSON-formatted byte array, which can be fed directly into vis.js as network map
func (g *Graph) GetDataJSON() []byte {
	result, err := json.Marshal(g.GetData())
	if err != nil {
		panic(fmt.Sprintf("Failed to marshal graph to JSON: %s", err))
	}
	return result
}

func (g *Graph) addNode(n graphNode, level int) {
	id := idEscape(fmt.Sprintf("node-%s", n.getID()))
	if _, ok := g.hasObject[id]; !ok {
		nodeEntry := graphEntry{
			"id":     id,
			"g_type": "node",
			"label":  n.getLabel(),
			"group":  n.getGroup(),
		}
		if level >= 0 {
			nodeEntry["level"] = level
		}
		g.addEntry(nodeEntry)
	}
}

func (g *Graph) addEdge(e graphEdge) {
	idFrom := idEscape(fmt.Sprintf("node-%s", e.getSrc().getID()))
	idTo := idEscape(fmt.Sprintf("node-%s", e.getDst().getID()))
	id := fmt.Sprintf("edge-%s-%s", idFrom, idTo)
	if existing, ok := g.hasObject[id]; !ok {
		edgeEntry := graphEntry{
			"id":       id,
			"g_type":   "edge",
			"from":     idFrom,
			"to":       idTo,
			"labelMap": map[string]bool{e.getLabel(): true},
			"label":    e.getLabel(),
		}
		g.addEntry(edgeEntry)
	} else {
		existing["labelMap"].(map[string]bool)[e.getLabel()] = true
		labels := []string{}
		for label := range existing["labelMap"].(map[string]bool) {
			labels = append(labels, label)
		}
		sort.Strings(labels)
		existing["label"] = strings.Join(labels, ",")
	}
}

func (g *Graph) addEntry(e graphEntry) {
	g.hasObject[e["id"].(string)] = e
	if e["g_type"] == "node" {
		g.nodes = append(g.nodes, e)
	} else if e["g_type"] == "edge" {
		g.edges = append(g.edges, e)
	} else {
		panic(fmt.Sprintf("Can't add entry to the graph. Unknown type: %s", e))
	}
}
