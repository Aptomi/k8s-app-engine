package visualization

import (
	"github.com/Aptomi/aptomi/pkg/slinga/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/slinga/lang"
	"github.com/Aptomi/aptomi/pkg/slinga/lang/builder"
	"github.com/Aptomi/aptomi/pkg/slinga/util"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestVisualizationDiagram(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	b := makePolicyBuilder()

	// empty policy and empty resolution result
	policyEmpty := lang.NewPolicy()
	resolutionEmpty := resolve.NewPolicyResolution()

	// unit test policy resolved revision
	resolver := resolve.NewPolicyResolver(b.Policy(), b.External())
	resolutionNew, _, err := resolver.ResolveAllDependencies()
	if !assert.NoError(t, err, "Policy should be resolved without errors") {
		t.FailNow()
	}

	// generate images
	{
		imagePrev, err := CreateImage(NewDiagram(policyEmpty, resolutionEmpty, b.External()))
		assert.NoError(t, err, "Image should be generated")
		assert.True(t, imagePrev.Bounds().Dx() < 20, "Image for empty policy resolution should be empty")
		assert.True(t, imagePrev.Bounds().Dy() < 20, "Image for empty policy resolution should be empty")

		// OpenImage(imagePrev)
	}

	{
		imageNext, err := CreateImage(NewDiagram(b.Policy(), resolutionNew, b.External()))
		assert.NoError(t, err, "Image should be generated")
		assert.True(t, imageNext.Bounds().Dx() > 500, "Image for unit test resolved policy should be big enough: %s", imageNext.Bounds())
		assert.True(t, imageNext.Bounds().Dy() > 500, "Image for unit test resolved policy should be big enough: %s", imageNext.Bounds())

		// OpenImage(imageNext)
	}

	{
		// delta (empty) -> (non-empty)
		imageDiff, err := CreateImage(NewDiagramDelta(b.Policy(), resolutionNew, policyEmpty, resolutionEmpty, b.External()))
		assert.NoError(t, err, "Image should be generated")
		assert.True(t, imageDiff.Bounds().Dx() > 500, "Image for unit test resolved policy diff against empty (all additions) should be big enough: %s", imageDiff.Bounds())
		assert.True(t, imageDiff.Bounds().Dy() > 500, "Image for unit test resolved policy diff against empty (all additions) should be big enough: %s", imageDiff.Bounds())

		// OpenImage(imageDiff)
	}

	{
		// delta (non-empty) -> (empty)
		imageDiff, err := CreateImage(NewDiagramDelta(policyEmpty, resolutionEmpty, b.Policy(), resolutionNew, b.External()))
		assert.NoError(t, err, "Image should be generated")
		assert.True(t, imageDiff.Bounds().Dx() > 500, "Image for unit test resolved policy diff against empty (all deletions) should be big enough: %s", imageDiff.Bounds())
		assert.True(t, imageDiff.Bounds().Dy() > 500, "Image for unit test resolved policy diff against empty (all deletions) should be big enough: %s", imageDiff.Bounds())

		// OpenImage(imageDiff)
	}

}

/*
	Helpers
*/

func makePolicyBuilder() *builder.PolicyBuilder {
	b := builder.NewPolicyBuilder()

	// three services
	services := []*lang.Service{}
	contracts := []*lang.Contract{}
	user := b.AddUser()
	for i := 0; i < 3; i++ {
		service := b.AddService(user)
		contract := b.AddContract(service, b.CriteriaTrue())

		// three components each
		for j := 0; j < 3; j++ {
			b.AddServiceComponent(service, b.CodeComponent(util.NestedParameterMap{"cluster": "{{ .Labels.cluster }}"}, nil))
		}

		services = append(services, service)
		contracts = append(contracts, contract)
	}

	// add dependencies i -> i+1 (0 -> 1, 1 -> 2)
	for i := 0; i < 2; i++ {
		b.AddServiceComponent(services[i], b.ContractComponent(contracts[i+1]))
	}

	// one cluster
	clusterObj := b.AddCluster()
	b.AddRule(b.CriteriaTrue(), b.RuleActions(lang.Allow, lang.Allow, lang.NewLabelOperationsSetSingleLabel(lang.LabelCluster, clusterObj.Name)))

	// several dependencies
	for i := 0; i < 5; i++ {
		b.AddDependency(b.AddUser(), contracts[i%len(contracts)])
	}

	return b
}
