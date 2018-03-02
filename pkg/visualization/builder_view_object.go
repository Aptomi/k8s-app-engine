package visualization

import (
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/runtime"
)

// Object produces a graph which represents an object
func (b *GraphBuilder) Object(cfg *PolicyCfg, obj runtime.Object) *Graph {
	if service, ok := obj.(*lang.Service); ok {
		b.traceService(service, nil, "", 0, cfg)
	}
	if contract, ok := obj.(*lang.Contract); ok {
		b.traceContract(contract, nil, "", 0, cfg)
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
		// contract -> [context] as edge label -> service
		// lookup the corresponding service
		serviceObj, errService := b.policy.GetObject(lang.ServiceObject.Kind, context.Allocation.Service, contract.Namespace)
		if errService != nil {
			b.graph.addNode(errorNode{err: errService}, level)
			continue
		}
		service := serviceObj.(*lang.Service)

		// context -> service
		contextName := context.Name
		if len(context.Allocation.Keys) > 0 {
			contextName += " (+)"
		}
		b.traceService(service, ctrNode, contextName, level+1, cfg)
	}
}

func (b *GraphBuilder) traceService(service *lang.Service, last graphNode, lastLabel string, level int, cfg *PolicyCfg) {
	svcNode := serviceNode{service: service}
	b.graph.addNode(svcNode, level)
	if last != nil {
		b.graph.addEdge(newEdge(last, svcNode, lastLabel))
	}

	// for every contract service relies on
	codeComponents := 0
	for _, component := range service.Components {
		if component.Code != nil {
			codeComponents++
		}
	}
	for _, component := range service.Components {
		if component.Code != nil && (codeComponents > 1 || !cfg.optimizeServicesWithSingleComponent) {
			// service -> component
			cmpNode := componentNode{service: service, component: component}
			b.graph.addNode(cmpNode, level+1)
			b.graph.addEdge(newEdge(svcNode, cmpNode, ""))
		}
		if len(component.Contract) > 0 {
			contractObjNew, errContract := b.policy.GetObject(lang.ContractObject.Kind, component.Contract, service.Namespace)
			if errContract != nil {
				b.graph.addNode(errorNode{err: errContract}, level+1)
				continue
			}
			b.traceContract(contractObjNew.(*lang.Contract), svcNode, "", level+1, cfg)
		}
	}

}
