package visibility

import (
	"fmt"
	"sort"
)

type graph struct {
	hasObject map[string]bool
	nodes     lineEntryList
	edges     lineEntryList
}

// Creates a new graph
func newGraph() *graph {
	return &graph{
		hasObject: make(map[string]bool),
		nodes:     lineEntryList{},
		edges:     lineEntryList{},
	}
}

func (g *graph) GetData() lineEntry {
	// Sort nodes and edges, so we can get a stable response from API that doesn't change over reloads
	// This will ensure that UI will show the same layout over refreshes
	sort.Sort(g.nodes)
	sort.Sort(g.edges)

	// Wrap it into a graph structure for vis.js
	return lineEntry{
		"nodes": g.nodes,
		"edges": g.edges,
	}
}

func (g *graph) addNode(n graphNode, level int) {
	key := fmt.Sprintf("node-%s", n.getID())
	if !g.hasObject[key] {
		g.nodes = append(g.nodes, lineEntry{
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
		g.edges = append(g.edges, lineEntry{
			"id":    key,
			"from":  src.getID(),
			"to":    dst.getID(),
			"label": src.getEdgeLabel(dst),
		})
		g.hasObject[key] = true
	}
}
