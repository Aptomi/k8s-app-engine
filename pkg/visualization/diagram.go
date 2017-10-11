package visualization

import (
	"github.com/Aptomi/aptomi/pkg/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/external"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/object"
	"github.com/awalterschulze/gographviz"
	"strconv"
)

// See http://www.graphviz.org/doc/info/colors.html
const noEntriesNodeName = "No entries"
const colorScheme = "set19"
const colorCount = 9

// Diagram is a visualization diagram for policy and resolution data
type Diagram struct {
	// Input data
	policy        *lang.Policy
	resolution    *resolve.PolicyResolution
	externalData  *external.Data
	showContracts bool

	// Internal map to keep track of used colors
	usedColors int
	colorMap   map[string]int

	// Keeps track of already created objects, so we don't create them twice
	created map[string]bool

	// Resulting graph
	graph *gographviz.Graph
}

// NewDiagram returns a Diagram visualizing policy and resolution data
func NewDiagram(policy *lang.Policy, resolution *resolve.PolicyResolution, externalData *external.Data) *gographviz.Graph {
	diagram := &Diagram{
		policy:        policy,
		resolution:    resolution,
		externalData:  externalData,
		showContracts: false,
		usedColors:    0,
		colorMap:      make(map[string]int),
		created:       make(map[string]bool),
		graph:         gographviz.NewGraph(),
	}
	return diagram.toGraphviz()
}

// NewDiagramDelta returns a Diagram visualizing delta between two "policy/resolution" states
func NewDiagramDelta(nextPolicy *lang.Policy, nextResolution *resolve.PolicyResolution, prevPolicy *lang.Policy, prevResolution *resolve.PolicyResolution, externalData *external.Data) *gographviz.Graph {
	nextGraph := NewDiagram(nextPolicy, nextResolution, externalData)
	prevGraph := NewDiagram(prevPolicy, prevResolution, externalData)
	return Delta(prevGraph, nextGraph)
}

// Produces diagram as graphviz graph
func (d *Diagram) toGraphviz() *gographviz.Graph {
	// Initialize graph properties
	_ = d.graph.SetName("Main")
	_ = d.graph.AddAttr("Main", "compound", "true")
	_ = d.graph.SetDir(true)

	if d.showContracts {
		// Render all contexts
		// [namespace] box
		//     -> [contract_1] box
		//         -> [context_1] node
		//         ...
		//         -> [context_M] node
		//     ...
		//     -> [contract_N] box
		//         -> [context_1] node
		//         ...
		//         -> [context_M] node
		for _, contractObj := range d.policy.GetObjectsByKind(lang.ContractObject.Kind) {
			contract := contractObj.(*lang.Contract)
			for _, context := range contract.Contexts {
				contextNode{contract: contract, contextName: context.Name}.render(d)
			}
		}
	}

	// Render all dependencies. As edges into service boxes and instances
	// [namespace] box
	//     -> [service_1] box
	//         -> [instance_1] node
	//         ...
	//         -> [instance_M] node
	//     ...
	//     -> [service_N] box
	//         -> [instance_1] node
	//         ...
	//         -> [instance_M] node
	for _, dependencyObj := range d.policy.GetObjectsByKind(lang.DependencyObject.Kind) {
		dependency := dependencyObj.(*lang.Dependency)
		dependencyNode := dependencyNode{dependency: dependency}
		dependencyNode.render(d)
		d.traceKey("", dependency, dependencyNode)
	}

	return d.graph
}

func (d *Diagram) traceKey(keySrc string, dependency *lang.Dependency, last node) {
	var edgesOut map[string]bool
	if len(keySrc) <= 0 {
		edgesOut = make(map[string]bool)
		resolvedKey := d.resolution.DependencyInstanceMap[object.GetKey(dependency)]
		if len(resolvedKey) > 0 {
			edgesOut[resolvedKey] = true
		}
	} else {
		instance := d.resolution.ComponentInstanceMap[keySrc]
		edgesOut = instance.EdgesOut
	}

	// recursively walk the graph
	for keyDst := range edgesOut {
		instanceCurrent := d.resolution.ComponentInstanceMap[keyDst]
		if instanceCurrent.Metadata.Key.IsService() {
			if d.showContracts {
				// render contract node
				contractObj, errContract := d.policy.GetObject(lang.ContractObject.Kind, instanceCurrent.Metadata.Key.ContractName, instanceCurrent.Metadata.Key.Namespace)
				if errContract == nil {
					ctxNode := contextNode{contract: contractObj.(*lang.Contract), contextName: instanceCurrent.Metadata.Key.ContextName}
					ctxNode.render(d)

					// add an edge, last -> context node
					edge{src: last, dst: ctxNode, color: d.getColor(object.GetKey(dependency))}.render(d)

					// render service instance node
					serviceObj, errService := d.policy.GetObject(lang.ServiceObject.Kind, instanceCurrent.Metadata.Key.ServiceName, instanceCurrent.Metadata.Key.Namespace)
					if errService == nil {
						svcInstNode := serviceInstanceNode{instance: instanceCurrent, service: serviceObj.(*lang.Service)}
						svcInstNode.render(d)

						// add an edge, last -> service instance node
						edge{src: ctxNode, dst: svcInstNode, color: d.getColor(object.GetKey(dependency))}.render(d)

						// trace current key with updated last
						d.traceKey(keyDst, dependency, svcInstNode)
					} else {
						errorNode{err: errService}.render(d)
					}
				} else {
					errorNode{err: errContract}.render(d)
				}
			} else {
				// render service instance node
				serviceObj, errService := d.policy.GetObject(lang.ServiceObject.Kind, instanceCurrent.Metadata.Key.ServiceName, instanceCurrent.Metadata.Key.Namespace)
				if errService == nil {
					svcInstNode := serviceInstanceNode{instance: instanceCurrent, service: serviceObj.(*lang.Service)}
					svcInstNode.render(d)

					// add an edge, last -> service instance node
					edge{src: last, dst: svcInstNode, color: d.getColor(object.GetKey(dependency))}.render(d)

					// trace current key with updated last
					d.traceKey(keyDst, dependency, svcInstNode)
				} else {
					errorNode{err: errService}.render(d)
				}
			}
		} else {
			// trace current key and keep last as is
			d.traceKey(keyDst, dependency, last)
		}
	}
}

// Returns a color for the given key
func (d *Diagram) getColor(key string) string {
	color, ok := d.colorMap[key]
	if !ok {
		d.usedColors++
		if d.usedColors > colorCount {
			d.usedColors = 1
		}
		d.colorMap[key] = d.usedColors
		color = d.usedColors
	}
	return "/" + colorScheme + "/" + strconv.Itoa(color)
}
