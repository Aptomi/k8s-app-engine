package visualization

import (
	"github.com/Aptomi/aptomi/pkg/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/runtime"
)

// ClaimResolutionCfg defines graph generation parameters for ClaimResolution
type ClaimResolutionCfg struct {
	showTrivialServices bool
	showServices        bool
}

// ClaimResolutionCfgDefault is default graph generation parameters for ClaimResolution
var ClaimResolutionCfgDefault = &ClaimResolutionCfg{
	showTrivialServices: true,
	showServices:        false,
}

// ClaimResolutionWithFunc produces policy resolution graph by tracing every claim and displaying what got allocated,
// which checking that instances exist (e.g. in actual state)
func (b *GraphBuilder) ClaimResolutionWithFunc(cfg *ClaimResolutionCfg, exists func(*resolve.ComponentInstance) bool) *Graph {
	// trace all claims
	for _, claimObj := range b.policy.GetObjectsByKind(lang.ClaimType.Kind) {
		claim := claimObj.(*lang.Claim) // nolint: errcheck
		b.traceClaimResolution("", claim, nil, 0, cfg, exists)
	}
	return b.graph
}

// ClaimResolution produces policy resolution graph by tracing every claim and displaying what got allocated
func (b *GraphBuilder) ClaimResolution(cfg *ClaimResolutionCfg) *Graph {
	return b.ClaimResolutionWithFunc(cfg, func(*resolve.ComponentInstance) bool { return true })
}

func (b *GraphBuilder) traceClaimResolution(keySrc string, claim *lang.Claim, last graphNode, level int, cfg *ClaimResolutionCfg, exists func(*resolve.ComponentInstance) bool) {
	var edgesOut map[string]bool
	if len(keySrc) <= 0 {
		// create a claim node
		cNode := claimNode{claim: claim, b: b}
		b.graph.addNode(cNode, 0)
		last = cNode
		level++

		// add an outgoing edge to its corresponding bundle instance
		edgesOut = make(map[string]bool)
		dResolution := b.resolution.GetClaimResolution(claim)
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

		// check that instance contains our claim
		depKey := runtime.KeyForStorable(claim)
		if _, found := instanceCurrent.ClaimKeys[depKey]; !found {
			continue
		}

		// check that instance exists
		if !exists(instanceCurrent) {
			continue
		}

		if instanceCurrent.Metadata.Key.IsBundle() {
			// if it's a bundle, then create a service node
			serviceObj, errService := b.policy.GetObject(lang.ServiceObject.Kind, instanceCurrent.Metadata.Key.ServiceName, instanceCurrent.Metadata.Key.Namespace)
			if errService != nil {
				b.graph.addNode(errorNode{err: errService}, level)
				continue
			}
			service := serviceObj.(*lang.Service) // nolint: errcheck
			ctrNode := serviceNode{service: service}

			// then create a bundle instance node
			bundleObj, errBundle := b.policy.GetObject(lang.BundleType.Kind, instanceCurrent.Metadata.Key.BundleName, instanceCurrent.Metadata.Key.Namespace)
			if errBundle != nil {
				b.graph.addNode(errorNode{err: errBundle}, level)
				continue
			}
			bundle := bundleObj.(*lang.Bundle) // nolint: errcheck
			svcInstNode := bundleInstanceNode{instance: instanceCurrent, bundle: bundle}

			// let's see if we need to show last -> service -> bundleInstance, or skip service all together
			trivialService := len(service.Contexts) <= 1
			if cfg.showServices && (!trivialService || cfg.showTrivialServices) {
				// show 'last' -> 'service' -> 'bundleInstance' -> (continue)
				b.graph.addNode(ctrNode, level)
				b.graph.addEdge(newEdge(last, ctrNode, ""))

				b.graph.addNode(svcInstNode, level+1)
				b.graph.addEdge(newEdge(ctrNode, svcInstNode, instanceCurrent.Metadata.Key.ContextNameWithKeys))

				// continue tracing
				b.traceClaimResolution(keyDst, claim, svcInstNode, level+2, cfg, exists)
			} else {
				// skip service, show just 'last' -> 'bundleInstance' -> (continue)
				b.graph.addNode(svcInstNode, level)
				b.graph.addEdge(newEdge(last, svcInstNode, ""))

				// continue tracing
				b.traceClaimResolution(keyDst, claim, svcInstNode, level+1, cfg, exists)
			}
		} else {
			// if it's a component, we don't need to show any additional nodes, let's just continue
			// though, we could introduce additional flag which allows to render components, if needed
			b.traceClaimResolution(keyDst, claim, last, level, cfg, exists)
		}
	}
}
