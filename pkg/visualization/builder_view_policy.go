package visualization

import (
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/runtime"
)

// PolicyCfg defines graph generation parameters for Policy
type PolicyCfg struct {
	showServiceComponentsOnlyForTopLevel bool
	showServiceComponents                bool
}

// PolicyCfgDefault is default graph generation parameters for Policy
var PolicyCfgDefault = &PolicyCfg{
	showServiceComponentsOnlyForTopLevel: true,
	showServiceComponents:                true,
}

// Policy produces just a policy graph without showing any resolution data
func (b *GraphBuilder) Policy(cfg *PolicyCfg) *Graph {
	// we need to find all top-level contracts
	contractDegIn := make(map[string]int)
	for _, contractObj := range b.policy.GetObjectsByKind(lang.ContractObject.Kind) {
		contract := contractObj.(*lang.Contract)
		b.calcContractDegIn(contract, contractDegIn)
	}

	// trace all top-level contracts
	for _, contractObj := range b.policy.GetObjectsByKind(lang.ContractObject.Kind) {
		contract := contractObj.(*lang.Contract)
		if contractDegIn[runtime.KeyForStorable(contract)] <= 0 {
			b.traceContract(contract, nil, "", 0, cfg)
		}
	}
	return b.graph
}

func (b *GraphBuilder) calcContractDegIn(contractFrom *lang.Contract, contractDegIn map[string]int) {
	for _, context := range contractFrom.Contexts {
		serviceObj, errService := b.policy.GetObject(lang.ServiceObject.Kind, context.Allocation.Service, contractFrom.Namespace)
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
				contractTo := contractObjNew.(*lang.Contract)
				contractDegIn[runtime.KeyForStorable(contractTo)]++
			}
		}
	}
}
