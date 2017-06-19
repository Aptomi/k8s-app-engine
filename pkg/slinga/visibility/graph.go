package visibility

type graphNode interface {
	getID() string
	getLabel() string
	getGroup() string
	getEdgeLabel(graphNode) string
}

type graphEntry map[string]interface{}

type graph struct {
	nodes []graphEntry
	edges []graphEntry
}

func NewGraph() graph {
	return graph{
		nodes: []graphEntry{},
		edges: []graphEntry{},
	}
}

func (g *graph) GetData() graphEntry {
	return graphEntry{
		"nodes": g.nodes,
		"edges": g.edges,
	}
}

func (g *graph) addNode(n graphNode) {
	g.nodes = append(g.nodes, graphEntry{
		"id":    n.getID(),
		"label": n.getLabel(),
		"group": n.getGroup(),
	})
}

func (g *graph) addEdge(src graphNode, dst graphNode) {
	label := src.getEdgeLabel(dst)
	g.edges = append(g.edges, graphEntry{
		"from":  src.getID(),
		"to":    dst.getID(),
		"label": label,
	})
}
