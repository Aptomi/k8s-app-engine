package visibility

import (
	"fmt"
	"sort"
)

type graph struct {
	hasObject map[string]bool
	nodes     graphEntryList
	edges     graphEntryList
}

type graphEntry map[string]interface{}

type graphEntryList []graphEntry

func (list graphEntryList) Len() int {
	return len(list)
}

func (list graphEntryList) Less(i, j int) bool {
	return list[i]["id"].(string) < list[j]["id"].(string)
}

func (list graphEntryList) Swap(i, j int) {
	list[i], list[j] = list[j], list[i]
}

// Creates a new graph
func newGraph() *graph {
	return &graph{
		hasObject: make(map[string]bool),
		nodes:     graphEntryList{},
		edges:     graphEntryList{},
	}
}

func (g *graph) GetData() graphEntry {
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

func (g *graph) addNode(n graphNode, level int) {
	key := fmt.Sprintf("node-%s", n.getID())
	if !g.hasObject[key] {
		g.nodes = append(g.nodes, graphEntry{
			"id":    n.getID(),
			"label": n.getLabel(),
			"group": n.getGroup(),
			"level": level,
		})
		g.hasObject[key] = true
	}
}

func (g *graph) addEdge(src graphNode, dst graphNode) {
	key := fmt.Sprintf("edge-%s-%s", src.getID(), dst.getID())
	if !g.hasObject[key] {
		g.edges = append(g.edges, graphEntry{
			"id":    key,
			"from":  src.getID(),
			"to":    dst.getID(),
			"label": src.getEdgeLabel(dst),
		})
		g.hasObject[key] = true
	}
}
