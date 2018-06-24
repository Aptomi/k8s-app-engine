package visualization

import (
	"testing"

	"github.com/Aptomi/aptomi/pkg/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/event"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/lang/builder"
	"github.com/Aptomi/aptomi/pkg/util"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestVisualizationDiagram(t *testing.T) {
	b := makePolicyBuilder()

	// empty policy and empty resolution result
	policyEmpty := lang.NewPolicy()
	resolutionEmpty := resolve.NewPolicyResolution()

	// unit test policy resolved revision
	eventLog := event.NewLog(logrus.WarnLevel, "test-resolve")
	resolver := resolve.NewPolicyResolver(b.Policy(), b.External(), eventLog)
	resolutionNew := resolver.ResolveAllClaims()
	if !assert.Equal(t, 14, len(resolutionNew.ComponentInstanceMap), "Instances should be resolved") {
		t.FailNow()
	}

	{
		data := NewGraphBuilder(policyEmpty, resolutionEmpty, b.External()).Policy(PolicyCfgDefault).GetDataJSON()
		if !assert.Condition(t, func() bool { return len(data) < 100 }, "Policy visualization: empty policy") {
			debug(t, data)
		}
	}

	{
		data := NewGraphBuilder(b.Policy(), resolutionEmpty, b.External()).Policy(PolicyCfgDefault).GetDataJSON()
		if !assert.Condition(t, func() bool { return len(data) > 2000 }, "Policy visualization: non-empty policy") {
			debug(t, data)
		}
	}

	{
		data := NewGraphBuilder(b.Policy(), resolutionNew, b.External()).ClaimResolution(ClaimResolutionCfgDefault).GetDataJSON()
		if !assert.Condition(t, func() bool { return len(data) > 2000 }, "Claim resolution visualization: non-empty policy") {
			debug(t, data)
		}
	}

	{
		empty := NewGraphBuilder(policyEmpty, resolutionEmpty, b.External()).ClaimResolution(ClaimResolutionCfgDefault)
		full := NewGraphBuilder(b.Policy(), resolutionNew, b.External()).ClaimResolution(ClaimResolutionCfgDefault)
		full.CalcDelta(empty)
		data := full.GetDataJSON()
		if !assert.Condition(t, func() bool { return len(data) > 4500 }, "Claim resolution visualization diff: empty -> non-empty (adding instances)") {
			debug(t, data)
		}
	}

	{
		empty := NewGraphBuilder(policyEmpty, resolutionEmpty, b.External()).ClaimResolution(ClaimResolutionCfgDefault)
		full := NewGraphBuilder(b.Policy(), resolutionNew, b.External()).ClaimResolution(ClaimResolutionCfgDefault)
		empty.CalcDelta(full)
		data := empty.GetDataJSON()
		if !assert.Condition(t, func() bool { return len(data) > 4500 }, "Claim resolution visualization diff: non-empty -> empty (removing instances)") {
			debug(t, data)
		}
	}

}

func debug(t *testing.T, data []byte) {
	t.Logf("JSON size: %d", len(data))
}

/*
	Helpers
*/

func makePolicyBuilder() *builder.PolicyBuilder {
	b := builder.NewPolicyBuilder()

	// three services
	services := []*lang.Service{}
	contracts := []*lang.Contract{}
	for i := 0; i < 3; i++ {
		service := b.AddService()
		contract := b.AddContract(service, b.CriteriaTrue())

		// three components each
		for j := 0; j < 3; j++ {
			b.AddServiceComponent(service, b.CodeComponent(util.NestedParameterMap{"debug": "{{ .Labels.target }}"}, nil))
		}

		services = append(services, service)
		contracts = append(contracts, contract)
	}

	// add component dependencies i -> i+1 (0 -> 1, 1 -> 2)
	for i := 0; i < 2; i++ {
		b.AddServiceComponent(services[i], b.ContractComponent(contracts[i+1]))
	}

	// one cluster
	clusterObj := b.AddCluster()
	b.AddRule(b.CriteriaTrue(), b.RuleActions(lang.NewLabelOperationsSetSingleLabel(lang.LabelTarget, clusterObj.Name)))

	// several claims
	for i := 0; i < 5; i++ {
		b.AddClaim(b.AddUser(), contracts[i%len(contracts)])
	}

	return b
}
