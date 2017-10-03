package graphviz

import (
	"github.com/awalterschulze/gographviz"
	"strings"
)

// Delta calculates difference between two graphs and returns it as a graph (it also modifies <next> to represent that difference)
func Delta(prev *gographviz.Graph, next *gographviz.Graph) *gographviz.Graph {
	// New nodes, edges, subgraphs must be highlighted
	{
		/*
			for _, s := range next.SubGraphs.SubGraphs {
				if _, inPrev := prev.SubGraphs.SubGraphs[s.Name]; !inPrev {
					// New subgraph -> no special treatment needed
					// _ = s.Attrs.Add("style", "filled")
				}
			}
		*/
		for _, n := range next.Nodes.Nodes {
			if _, inPrev := prev.Nodes.Lookup[n.Name]; !inPrev {
				// New node -> filled green
				_ = n.Attrs.Add("style", "filled")
				_ = n.Attrs.Add("color", "green2")
			}
		}
		for _, e := range next.Edges.Edges {
			if _, inPrev := prev.Edges.SrcToDsts[e.Src][e.Dst]; !inPrev {
				// New edge -> bold, same color
				_ = e.Attrs.Add("penwidth", "4")
			}
		}
	}

	// Removed nodes, edges, subgraphs must be highlighted
	{
		for _, s := range prev.SubGraphs.SubGraphs {
			if _, inNext := next.SubGraphs.SubGraphs[s.Name]; !inNext {
				// Removed subgraph -> add a sugraph filled red
				_ = next.AddSubGraph(next.Name, s.Name, map[string]string{"style": "filled", "fillcolor": "gray18", "fontcolor": "white", "label": s.Attrs["label"]})
			}
		}

		for _, n := range prev.Nodes.Nodes {
			if _, inNext := next.Nodes.Lookup[n.Name]; !inNext {

				// if the previous graph was empty and contained just one "empty" node, don't show it on delta
				if !strings.Contains(n.Name, noEntriesNodeName) {
					// Removed node -> add a node filled red
					_ = n.Attrs.Add("style", "filled")
					_ = n.Attrs.Add("color", "red")

					// Find previous subgraph & put it into the same subgraph
					subgraphName := findSubraphName(prev, n.Name)
					next.Relations.Add(subgraphName, n.Name)

					next.Nodes.Add(n)
				}
			}
		}
		for _, e := range prev.Edges.Edges {
			if _, inNext := next.Edges.SrcToDsts[e.Src][e.Dst]; !inNext {
				// Removed edge -> add an edge, dashed
				_ = e.Attrs.Add("style", "dashed")
				next.Edges.Add(e)
			}
		}
	}

	return next
}
