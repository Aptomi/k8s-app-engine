package visibility

import "fmt"

type graphNode interface {
	getID() string
	getLabel() string
	getGroup() string
	getEdgeLabel(graphNode) string
}

type graphEntry map[string]interface{}

type graph struct {
	hasObject map[string]bool
	nodes   []graphEntry
	edges   []graphEntry
}

func NewGraph() *graph {
	return &graph{
		hasObject: make(map[string]bool),
		nodes:   []graphEntry{},
		edges:   []graphEntry{},
	}
}

func (g *graph) GetData() graphEntry {
	return graphEntry{
		"nodes": g.nodes,
		"edges": g.edges,
	}
}

func (g *graph) addNode(n graphNode) {
	key := fmt.Sprintf("node-%s", n.getID())
	if !g.hasObject[key] {
		g.nodes = append(g.nodes, graphEntry{
			"id":    n.getID(),
			"label": n.getLabel(),
			"group": n.getGroup(),
		})
		g.hasObject[key] = true
	}
}

func (g *graph) addEdge(src graphNode, dst graphNode) {
	key := fmt.Sprintf("edge-%s-%s", src.getID(), dst.getID())
	if !g.hasObject[key] {
		g.edges = append(g.edges, graphEntry{
			"from":  src.getID(),
			"to":    dst.getID(),
			"label": src.getEdgeLabel(dst),
		})
		g.hasObject[key] = true
	}
}
