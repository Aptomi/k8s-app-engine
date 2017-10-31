package visualization

import (
	"github.com/Aptomi/aptomi/pkg/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/external"
	"github.com/Aptomi/aptomi/pkg/lang"
)

// GraphBuilder allows to visualize Aptomi policy and resolution data as a graph
// It can also render deltas between two visualizations, showing new and deleted objects between policies and resolution data.
type GraphBuilder struct {
	// Input data
	policy       *lang.Policy
	resolution   *resolve.PolicyResolution
	externalData *external.Data

	// Resulting graph
	graph *Graph
}

// NewGraphBuilder returns a graph builder that visualizes policy and resolution data
func NewGraphBuilder(policy *lang.Policy, resolution *resolve.PolicyResolution, externalData *external.Data) *GraphBuilder {
	return &GraphBuilder{
		policy:       policy,
		resolution:   resolution,
		externalData: externalData,
		graph:        newGraph(),
	}
}
