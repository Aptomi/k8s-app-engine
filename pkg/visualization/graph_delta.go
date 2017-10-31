package visualization

// CalcDelta takes the current graph and modifies it to contain the difference between this graph and the baseline graph
func (g *Graph) CalcDelta(baseline *Graph) {
	// Added nodes & edges must be highlighted
	for key, entry := range g.hasObject {
		_, inBaseline := baseline.hasObject[key]
		if !inBaseline {
			// object got added
			entry["shadow"] = graphEntry{
				"enabled": true,
				"color":   "rgb(0,255,0)",
				"size":    50,
			}
		}
	}

	// Removed nodes & edges must added & highlighted
	for key, entry := range baseline.hasObject {
		_, inCurrent := g.hasObject[key]
		if !inCurrent {
			// object got removed
			entry["shadow"] = graphEntry{
				"enabled": true,
				"color":   "rgb(255,0,0)",
				"size":    50,
			}
			g.addEntry(entry)
		}
	}
}
