package visualization

import (
	"github.com/Aptomi/aptomi/pkg/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/runtime"
)

// Object produces a graph which represents an object
func (b *GraphBuilder) Object(obj runtime.Object) *Graph {
	if bundle, ok := obj.(*lang.Bundle); ok {
		b.traceBundle(bundle, nil, "", 0, PolicyCfgDefault)
	}
	if contract, ok := obj.(*lang.Contract); ok {
		b.traceContract(contract, nil, "", 0, PolicyCfgDefault)
	}
	if claim, ok := obj.(*lang.Claim); ok {
		b.traceClaimResolution("", claim, nil, 0, ClaimResolutionCfgDefault, func(*resolve.ComponentInstance) bool { return true })
	}
	return b.graph
}

func (b *GraphBuilder) traceContract(contract *lang.Contract, last graphNode, lastLabel string, level int, cfg *PolicyCfg) {
	// [last] -> contract
	ctrNode := contractNode{contract: contract}
	b.graph.addNode(ctrNode, level)
	if last != nil {
		b.graph.addEdge(newEdge(last, ctrNode, lastLabel))
	}

	// show all contexts within a given contract
	for _, context := range contract.Contexts {
		// contract -> [context] as edge label -> bundle
		// lookup the corresponding bundle
		bundleObj, errBundle := b.policy.GetObject(lang.BundleObject.Kind, context.Allocation.Bundle, contract.Namespace)
		if errBundle != nil {
			b.graph.addNode(errorNode{err: errBundle}, level)
			continue
		}
		bundle := bundleObj.(*lang.Bundle) // nolint: errcheck

		// context -> bundle
		contextName := context.Name
		if len(context.Allocation.Keys) > 0 {
			contextName += " (+)"
		}
		b.traceBundle(bundle, ctrNode, contextName, level+1, cfg)
	}
}

func (b *GraphBuilder) traceBundle(bundle *lang.Bundle, last graphNode, lastLabel string, level int, cfg *PolicyCfg) {
	svcNode := bundleNode{bundle: bundle}
	b.graph.addNode(svcNode, level)
	if last != nil {
		b.graph.addEdge(newEdge(last, svcNode, lastLabel))
	}

	// process components first
	showedComponents := false
	for _, component := range bundle.Components {
		if component.Code != nil && cfg.showBundleComponents {
			// bundle -> component
			cmpNode := componentNode{bundle: bundle, component: component}
			b.graph.addNode(cmpNode, level+1)
			b.graph.addEdge(newEdge(svcNode, cmpNode, ""))
			showedComponents = true
		}
	}

	// do not show any more bundle components down the tree if we already showed them at top level
	cfgNext := &PolicyCfg{}
	*cfgNext = *cfg
	if cfg.showBundleComponentsOnlyForTopLevel && showedComponents {
		cfgNext.showBundleComponents = false
	}

	// process contracts after that
	for _, component := range bundle.Components {
		if len(component.Contract) > 0 {
			contractObjNew, errContract := b.policy.GetObject(lang.ContractObject.Kind, component.Contract, bundle.Namespace)
			if errContract != nil {
				b.graph.addNode(errorNode{err: errContract}, level+1)
				continue
			}
			b.traceContract(contractObjNew.(*lang.Contract), svcNode, "", level+1, cfgNext)
		}
	}

}
