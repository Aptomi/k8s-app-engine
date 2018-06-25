package visualization

import (
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/runtime"
)

// PolicyCfg defines graph generation parameters for Policy
type PolicyCfg struct {
	showBundleComponentsOnlyForTopLevel bool
	showBundleComponents                bool
}

// PolicyCfgDefault is default graph generation parameters for Policy
var PolicyCfgDefault = &PolicyCfg{
	showBundleComponentsOnlyForTopLevel: true,
	showBundleComponents:                true,
}

// Policy produces just a policy graph without showing any resolution data
func (b *GraphBuilder) Policy(cfg *PolicyCfg) *Graph {
	// we need to find all top-level contracts
	contractDegIn := make(map[string]int)
	for _, contractObj := range b.policy.GetObjectsByKind(lang.ContractObject.Kind) {
		contract := contractObj.(*lang.Contract) // nolint: errcheck
		b.calcContractDegIn(contract, contractDegIn)
	}

	// trace all top-level contracts
	for _, contractObj := range b.policy.GetObjectsByKind(lang.ContractObject.Kind) {
		contract := contractObj.(*lang.Contract) // nolint: errcheck
		if contractDegIn[runtime.KeyForStorable(contract)] <= 0 {
			b.traceContract(contract, nil, "", 0, cfg)
		}
	}
	return b.graph
}

func (b *GraphBuilder) calcContractDegIn(contractFrom *lang.Contract, contractDegIn map[string]int) {
	for _, context := range contractFrom.Contexts {
		bundleObj, errBundle := b.policy.GetObject(lang.BundleObject.Kind, context.Allocation.Bundle, contractFrom.Namespace)
		if errBundle != nil {
			continue
		}
		bundle := bundleObj.(*lang.Bundle) // nolint: errcheck

		for _, component := range bundle.Components {
			if len(component.Contract) > 0 {
				contractObjNew, errContract := b.policy.GetObject(lang.ContractObject.Kind, component.Contract, bundle.Namespace)
				if errContract != nil {
					continue
				}
				contractTo := contractObjNew.(*lang.Contract) // nolint: errcheck
				contractDegIn[runtime.KeyForStorable(contractTo)]++
			}
		}
	}
}
