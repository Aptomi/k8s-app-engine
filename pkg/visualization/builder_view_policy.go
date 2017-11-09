package visualization

import (
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/runtime"
)

// PolicyCfg defines graph generation parameters for Policy
type PolicyCfg struct {
	optimizeServicesWithSingleComponent bool
}

// PolicyCfgDefault is default graph generation parameters for Policy
var PolicyCfgDefault = &PolicyCfg{
	optimizeServicesWithSingleComponent: true,
}

// Policy produces just a policy graph without showing any resolution data
func (b *GraphBuilder) Policy(cfg *PolicyCfg) *Graph {
	// we need to find all top-level contracts
	edgesIn := make(map[string]int)
	for _, contractObj := range b.policy.GetObjectsByKind(lang.ContractObject.Kind) {
		contract := contractObj.(*lang.Contract)
		b.findEdgesIn(contract, edgesIn)
	}

	// trace all top-level contracts
	for _, contractObj := range b.policy.GetObjectsByKind(lang.ContractObject.Kind) {
		contract := contractObj.(*lang.Contract)
		if edgesIn[runtime.KeyForStorable(contract)] <= 0 {
			b.traceContract(contract, nil, 0, cfg)
		}
	}
	return b.graph
}

func (b *GraphBuilder) findEdgesIn(contract *lang.Contract, edgesIn map[string]int) {
	for _, context := range contract.Contexts {
		serviceObj, errService := b.policy.GetObject(lang.ServiceObject.Kind, context.Allocation.Service, contract.Namespace)
		if errService != nil {
			continue
		}
		service := serviceObj.(*lang.Service)

		for _, component := range service.Components {
			if len(component.Contract) > 0 {
				contractObjNew, errContract := b.policy.GetObject(lang.ContractObject.Kind, component.Contract, service.Namespace)
				if errContract != nil {
					continue
				}
				contract := contractObjNew.(*lang.Contract)
				edgesIn[runtime.KeyForStorable(contract)]++
			}
		}
	}
}

func (b *GraphBuilder) traceContract(contract *lang.Contract, last graphNode, level int, cfg *PolicyCfg) {
	// [last] -> contract
	ctrNode := contractNode{contract: contract}
	b.graph.addNode(ctrNode, level)
	if last != nil {
		b.graph.addEdge(newEdge(last, ctrNode, ""))
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
		svcNode := serviceNode{service: service}
		b.graph.addNode(svcNode, level+1)
		contextName := context.Name
		if len(context.Allocation.Keys) > 0 {
			contextName += " (+)"
		}
		b.graph.addEdge(newEdge(ctrNode, svcNode, contextName))

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
				b.graph.addNode(cmpNode, level+2)
				b.graph.addEdge(newEdge(svcNode, cmpNode, ""))
			}
			if len(component.Contract) > 0 {
				contractObjNew, errContract := b.policy.GetObject(lang.ContractObject.Kind, component.Contract, service.Namespace)
				if errContract != nil {
					b.graph.addNode(errorNode{err: errContract}, level+2)
					continue
				}
				b.traceContract(contractObjNew.(*lang.Contract), svcNode, level+2, cfg)
			}
		}
	}
}
