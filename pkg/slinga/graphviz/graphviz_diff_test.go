package graphviz

import (
	"github.com/Aptomi/aptomi/pkg/slinga/engine/diff"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/slinga/language"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPolicyVisualizationViaGraphviz(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	userLoader := language.NewUserLoaderFromDir("../testdata/unittests")

	// empty revision
	revisionEmpty := resolve.NewRevision(language.NewPolicyNamespace(),
		resolve.NewPolicyResolution(),
		userLoader,
	)

	// unit test policy resolved revision
	policy := language.LoadUnitTestsPolicy("../testdata/unittests")
	resolver := resolve.NewPolicyResolver(policy, userLoader)
	revisionNew, err := resolver.ResolveAllDependencies()
	if !assert.Nil(t, err, "Policy should be resolved without errors") {
		t.FailNow()
	}

	// create visualization engine from diff
	vis := NewPolicyVisualization(diff.NewRevisionDiff(revisionNew, revisionEmpty))

	// generate images
	{
		imagePrev, err := vis.GetImageForRevisionPrev()
		assert.Nil(t, err, "Image should be generated")
		assert.True(t, imagePrev.Bounds().Dx() < 20, "Image for empty revision should be empty")
		assert.True(t, imagePrev.Bounds().Dy() < 20, "Image for empty revision should be empty")

		// OpenImage(imagePrev)
	}

	{
		imageNext, err := vis.GetImageForRevisionNext()
		assert.Nil(t, err, "Image should be generated")
		assert.True(t, imageNext.Bounds().Dx() > 800, "Image for unit test resolved policy should be big enough")
		assert.True(t, imageNext.Bounds().Dy() > 800, "Image for unit test resolved policy should be big enough")

		// OpenImage(imageNext)
	}

	{
		imageDiff, err := vis.GetImageForRevisionDiff()
		assert.Nil(t, err, "Image should be generated")
		assert.True(t, imageDiff.Bounds().Dx() > 800, "Image for unit test resolved policy diff with empty should be big enough")
		assert.True(t, imageDiff.Bounds().Dy() > 800, "Image for unit test resolved policy diff with empty should be big enough")

		// OpenImage(imageDiff)
	}
}
