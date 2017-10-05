package graphviz

import (
	"github.com/Aptomi/aptomi/pkg/slinga/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/slinga/external"
	"github.com/Aptomi/aptomi/pkg/slinga/external/secrets"
	"github.com/Aptomi/aptomi/pkg/slinga/external/users"
	"github.com/Aptomi/aptomi/pkg/slinga/lang"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPolicyVisualization(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	externalData := external.NewData(
		users.NewUserLoaderFromDir("../testdata/unittests"),
		secrets.NewSecretLoaderFromDir("../testdata/unittests"),
	)

	// empty policy and empty resolution result
	policyEmpty := lang.NewPolicy()
	resolutionEmpty := resolve.NewPolicyResolution()

	// unit test policy resolved revision
	policy := lang.LoadUnitTestsPolicy("../testdata/unittests")
	resolver := resolve.NewPolicyResolver(policy, externalData)
	resolutionNew, _, err := resolver.ResolveAllDependencies()
	if !assert.NoError(t, err, "Policy should be resolved without errors") {
		t.FailNow()
	}

	// generate images
	{
		imagePrev, err := NewPolicyVisualizationImage(policyEmpty, resolutionEmpty, externalData)
		assert.NoError(t, err, "Image should be generated")
		assert.True(t, imagePrev.Bounds().Dx() < 20, "Image for empty policy resolution should be empty")
		assert.True(t, imagePrev.Bounds().Dy() < 20, "Image for empty policy resolution should be empty")

		// OpenImage(imagePrev)
	}

	{
		imageNext, err := NewPolicyVisualizationImage(policy, resolutionNew, externalData)
		assert.NoError(t, err, "Image should be generated")
		assert.True(t, imageNext.Bounds().Dx() > 800, "Image for unit test resolved policy should be big enough")
		assert.True(t, imageNext.Bounds().Dy() > 800, "Image for unit test resolved policy should be big enough")

		// OpenImage(imageNext)
	}

	{
		// delta (empty) -> (non-empty)
		imageDiff, err := NewPolicyVisualizationDeltaImage(policy, resolutionNew, policyEmpty, resolutionEmpty, externalData)
		assert.NoError(t, err, "Image should be generated")
		assert.True(t, imageDiff.Bounds().Dx() > 800, "Image for unit test resolved policy diff against empty (all additions) should be big enough")
		assert.True(t, imageDiff.Bounds().Dy() > 800, "Image for unit test resolved policy diff against empty (all additions) should be big enough")

		// OpenImage(imageDiff)
	}

	{
		// delta (non-empty) -> (empty)
		imageDiff, err := NewPolicyVisualizationDeltaImage(policyEmpty, resolutionEmpty, policy, resolutionNew, externalData)
		assert.NoError(t, err, "Image should be generated")
		assert.True(t, imageDiff.Bounds().Dx() > 800, "Image for unit test resolved policy diff against empty (all deletions) should be big enough")
		assert.True(t, imageDiff.Bounds().Dy() > 800, "Image for unit test resolved policy diff against empty (all deletions) should be big enough")

		// OpenImage(imageDiff)
	}

}
