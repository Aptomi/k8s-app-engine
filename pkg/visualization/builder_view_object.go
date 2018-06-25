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
	if service, ok := obj.(*lang.Service); ok {
		b.traceService(service, nil, "", 0, PolicyCfgDefault)
	}
	if claim, ok := obj.(*lang.Claim); ok {
		b.traceClaimResolution("", claim, nil, 0, ClaimResolutionCfgDefault, func(*resolve.ComponentInstance) bool { return true })
	}
	return b.graph
}

func (b *GraphBuilder) traceService(service *lang.Service, last graphNode, lastLabel string, level int, cfg *PolicyCfg) {
	// [last] -> service
	ctrNode := serviceNode{service: service}
	b.graph.addNode(ctrNode, level)
	if last != nil {
		b.graph.addEdge(newEdge(last, ctrNode, lastLabel))
	}

	// show all contexts within a given service
	for _, context := range service.Contexts {
		// service -> [context] as edge label -> bundle
		// lookup the corresponding bundle
		bundleObj, errBundle := b.policy.GetObject(lang.BundleObject.Kind, context.Allocation.Bundle, service.Namespace)
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

	// process services after that
	for _, component := range bundle.Components {
		if len(component.Service) > 0 {
			serviceObjNew, errService := b.policy.GetObject(lang.ServiceObject.Kind, component.Service, bundle.Namespace)
			if errService != nil {
				b.graph.addNode(errorNode{err: errService}, level+1)
				continue
			}
			b.traceService(serviceObjNew.(*lang.Service), svcNode, "", level+1, cfgNext)
		}
	}

}
