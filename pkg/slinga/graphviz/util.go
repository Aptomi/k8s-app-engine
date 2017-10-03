package graphviz

import (
	"bytes"
	"fmt"
	"github.com/awalterschulze/gographviz"
	"image"
	"image/png"
	"io/ioutil"
	"os/exec"
)

// Returns a short version of the string
func shorten(s string) string {
	const maxLen = 20
	const suffix = "..."
	if len(s) > maxLen-len(suffix) {
		return s[0:maxLen-len(suffix)] + suffix
	}
	return s
}

// Finds subgraph name from relations
func findSubraphName(prev *gographviz.Graph, nName string) string {
	for gName := range prev.Relations.ParentToChildren {
		if prev.Relations.ParentToChildren[gName][nName] {
			return gName
		}
	}
	return prev.Name
}

// Adds an edge if it doesn't exist already
func addEdge(g *gographviz.Graph, src string, dst string, attrs map[string]string) {
	_ = g.AddEdge(esc(src), esc(dst), true, escAttrs(attrs))
}

// Adds a subgraph if it doesn't exist already
func addSubgraphOnce(g *gographviz.Graph, parentGraph string, name string, attrs map[string]string, was map[string]bool) {
	wasKey := "SUBGRAPH" + "#" + parentGraph + "#" + name
	if !was[wasKey] {
		_ = g.AddSubGraph(esc(parentGraph), esc(name), escAttrs(attrs))
		was[wasKey] = true
	}
}

// Adds a node if it doesn't exist already
func addNodeOnce(g *gographviz.Graph, parentGraph string, name string, attrs map[string]string, was map[string]bool) {
	wasKey := "NODE" + "#" + parentGraph + "#" + name
	if !was[wasKey] {
		_ = g.AddNode(esc(parentGraph), esc(name), escAttrs(attrs))
		was[wasKey] = true
	}
}

// OpenImage an image via "open" (preview in Mac OS) for debugging purposes
func OpenImage(image image.Image) {
	tmpFile, err := ioutil.TempFile("", "aptomi-graphviz-debug")
	if err != nil {
		panic("Failed to create temp file")
	}
	defer tmpFile.Close() // nolint: errcheck

	err = png.Encode(tmpFile, image)
	if err != nil {
		panic("Failed to write to temp file")
	}

	// Call open on the image
	{
		cmd := "open"
		args := []string{tmpFile.Name()}
		command := exec.Command(cmd, args...)
		var outb, errb bytes.Buffer
		command.Stdout = &outb
		command.Stderr = &errb
		if err := command.Run(); err != nil || len(errb.String()) > 0 {
			panic(fmt.Sprintf("Unable to open '%s': %s", tmpFile.Name(), err.Error()))
		}
	}
}
