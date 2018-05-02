package visualization

import (
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/runtime"
)

// DependencyResolutionCfg defines graph generation parameters for DependencyResolution
type DependencyResolutionCfg struct {
	showTrivialContracts bool
	showContracts        bool
}

// DependencyResolutionCfgDefault is default graph generation parameters for DependencyResolution
var DependencyResolutionCfgDefault = &DependencyResolutionCfg{
	showTrivialContracts: true,
	showContracts:        false,
}

// DependencyResolution produces policy resolution graph by tracing every dependency and displaying what got allocated
func (b *GraphBuilder) DependencyResolution(cfg *DependencyResolutionCfg) *Graph {
	// trace all dependencies
	for _, dependencyObj := range b.policy.GetObjectsByKind(lang.DependencyObject.Kind) {
		dependency := dependencyObj.(*lang.Dependency)
		b.traceDependencyResolution("", dependency, nil, 0, cfg)
	}
	return b.graph
}

func (b *GraphBuilder) traceDependencyResolution(keySrc string, dependency *lang.Dependency, last graphNode, level int, cfg *DependencyResolutionCfg) {
	var edgesOut map[string]bool
	if len(keySrc) <= 0 {
		// create a dependency node
		depNode := dependencyNode{dependency: dependency, b: b}
		b.graph.addNode(depNode, 0)
		last = depNode
		level++

		// add an outgoing edge to its corresponding service instance
		edgesOut = make(map[string]bool)
		dResolution := b.resolution.GetDependencyInstanceMap()[runtime.KeyForStorable(dependency)]
		if dResolution.Resolved {
			edgesOut[dResolution.ComponentInstanceKey] = true
		}
	} else {
		// if we are processing a component instance, then follow the recorded graph edges
		edgesOut = b.resolution.ComponentInstanceMap[keySrc].EdgesOut
	}

	// recursively walk the graph
	for keyDst := range edgesOut {
		instanceCurrent := b.resolution.ComponentInstanceMap[keyDst]

		// check that instance contains our dependency
		depKey := runtime.KeyForStorable(dependency)
		if !instanceCurrent.DependencyKeys[depKey] {
			continue
		}

		if instanceCurrent.Metadata.Key.IsService() {
			// if it's a service, then create a contract node
			contractObj, errContract := b.policy.GetObject(lang.ContractObject.Kind, instanceCurrent.Metadata.Key.ContractName, instanceCurrent.Metadata.Key.Namespace)
			if errContract != nil {
				b.graph.addNode(errorNode{err: errContract}, level)
				continue
			}
			contract := contractObj.(*lang.Contract)
			ctrNode := contractNode{contract: contract}

			// then create a service instance node
			serviceObj, errService := b.policy.GetObject(lang.ServiceObject.Kind, instanceCurrent.Metadata.Key.ServiceName, instanceCurrent.Metadata.Key.Namespace)
			if errService != nil {
				b.graph.addNode(errorNode{err: errService}, level)
				continue
			}
			service := serviceObj.(*lang.Service)
			svcInstNode := serviceInstanceNode{instance: instanceCurrent, service: service}

			// let's see if we need to show last -> contract -> serviceInstance, or skip contract all together
			trivialContract := len(contract.Contexts) <= 1
			if cfg.showContracts && (!trivialContract || cfg.showTrivialContracts) {
				// show 'last' -> 'contract' -> 'serviceInstance' -> (continue)
				b.graph.addNode(ctrNode, level)
				b.graph.addEdge(newEdge(last, ctrNode, ""))

				b.graph.addNode(svcInstNode, level+1)
				b.graph.addEdge(newEdge(ctrNode, svcInstNode, instanceCurrent.Metadata.Key.ContextNameWithKeys))

				// continue tracing
				b.traceDependencyResolution(keyDst, dependency, svcInstNode, level+2, cfg)
			} else {
				// skip contract, show just 'last' -> 'serviceInstance' -> (continue)
				b.graph.addNode(svcInstNode, level)
				b.graph.addEdge(newEdge(last, svcInstNode, ""))

				// continue tracing
				b.traceDependencyResolution(keyDst, dependency, svcInstNode, level+1, cfg)
			}
		} else {
			// if it's a component, we don't need to show any additional nodes, let's just continue
			// though, we could introduce additional flag which allows to render components, if needed
			b.traceDependencyResolution(keyDst, dependency, last, level, cfg)
		}
	}
}
