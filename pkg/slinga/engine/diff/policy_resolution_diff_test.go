package diff

import (
	"github.com/Aptomi/aptomi/pkg/slinga/language"
	"github.com/Aptomi/aptomi/pkg/slinga/object"
	"testing"
)

func TestEmptyDiff(t *testing.T) {
	userLoader := getUserLoader()
	resolvedPrev := resolvePolicy(t, getPolicy(), userLoader)
	resolvedPrev = emulateSaveAndLoadResolution(resolvedPrev)

	resolvedNext := resolvePolicy(t, getPolicy(), userLoader)

	// Calculate and verify difference
	diff := NewPolicyResolutionDiff(resolvedNext, resolvedPrev)
	verifyDiff(t, diff, 0, 0, 0, 0, 0)
}

func TestDiffHasCreatedComponents(t *testing.T) {
	userLoader := getUserLoader()

	resolvedPrev := resolvePolicy(t, getPolicy(), userLoader)
	resolvedPrev = emulateSaveAndLoadResolution(resolvedPrev)

	// Add another dependency and resolve policy
	nextPolicy := language.LoadUnitTestsPolicy("../../testdata/unittests")
	nextPolicy.Dependencies.AddDependency(
		&language.Dependency{
			Metadata: object.Metadata{
				Namespace: "main",
				Name:      "dep_id_5",
			},
			UserID:  "5",
			Service: "kafka",
		},
	)
	resolvedNext := resolvePolicy(t, nextPolicy, userLoader)

	// Calculate difference
	diff := NewPolicyResolutionDiff(resolvedNext, resolvedPrev)
	verifyDiff(t, diff, 8, 0, 0, 8, 0)
}

func TestDiffHasUpdatedComponents(t *testing.T) {
	userLoader := getUserLoader()

	// Add dependency, resolve policy
	policyNext := language.LoadUnitTestsPolicy("../../testdata/unittests")
	policyNext.Dependencies.AddDependency(
		&language.Dependency{
			Metadata: object.Metadata{
				Namespace: "main",
				Name:      "dep_id_5",
			},
			UserID:  "5",
			Service: "kafka",
		},
	)
	resolvedNew := resolvePolicy(t, policyNext, userLoader)

	// Update user label, re-evaluate and see that component instance has changed
	userLoader = language.NewUserLoaderFromDir("../../testdata/unittests")
	userLoader.LoadUserByID("5").Labels["changinglabel"] = "newvalue"
	resolvedDependencyUpdate := resolvePolicy(t, policyNext, userLoader)

	// Get the diff
	diff := NewPolicyResolutionDiff(resolvedDependencyUpdate, resolvedNew)

	// Check that update has been performed (on component and on parent service)
	verifyDiff(t, diff, 0, 0, 2, 0, 0)
}

func TestDiffHasDestructedComponents(t *testing.T) {
	// Resolve unit test policy
	userLoader := getUserLoader()
	resolvedPrev := resolvePolicy(t, getPolicy(), userLoader)
	resolvedPrev = emulateSaveAndLoadResolution(resolvedPrev)

	// Now resolve empty policy
	nextPolicy := language.NewPolicyNamespace()
	resolvedNext := resolvePolicy(t, nextPolicy, userLoader)

	// Calculate difference
	diff := NewPolicyResolutionDiff(resolvedNext, resolvedPrev)
	verifyDiff(t, diff, 0, 16, 0, 0, 16)
}
