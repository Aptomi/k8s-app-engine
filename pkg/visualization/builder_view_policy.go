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
	// we need to find all top-level services
	serviceDegIn := make(map[string]int)
	for _, serviceObj := range b.policy.GetObjectsByKind(lang.ServiceObject.Kind) {
		service := serviceObj.(*lang.Service) // nolint: errcheck
		b.calcServiceDegIn(service, serviceDegIn)
	}

	// trace all top-level services
	for _, serviceObj := range b.policy.GetObjectsByKind(lang.ServiceObject.Kind) {
		service := serviceObj.(*lang.Service) // nolint: errcheck
		if serviceDegIn[runtime.KeyForStorable(service)] <= 0 {
			b.traceService(service, nil, "", 0, cfg)
		}
	}
	return b.graph
}

func (b *GraphBuilder) calcServiceDegIn(serviceFrom *lang.Service, serviceDegIn map[string]int) {
	for _, context := range serviceFrom.Contexts {
		bundleObj, errBundle := b.policy.GetObject(lang.BundleType.Kind, context.Allocation.Bundle, serviceFrom.Namespace)
		if errBundle != nil {
			continue
		}
		bundle := bundleObj.(*lang.Bundle) // nolint: errcheck

		for _, component := range bundle.Components {
			if len(component.Service) > 0 {
				serviceObjNew, errService := b.policy.GetObject(lang.ServiceObject.Kind, component.Service, bundle.Namespace)
				if errService != nil {
					continue
				}
				serviceTo := serviceObjNew.(*lang.Service) // nolint: errcheck
				serviceDegIn[runtime.KeyForStorable(serviceTo)]++
			}
		}
	}
}
