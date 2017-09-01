package diff

import (
	"github.com/Aptomi/aptomi/pkg/slinga/language"
	"github.com/Aptomi/aptomi/pkg/slinga/object"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func BenchmarkEngine(b *testing.B) {
	t := &testing.T{}
	for i := 0; i < b.N; i++ {
		TestDiffHasUpdatedComponentsAndCheckTimes(t)
	}
}

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

func TestDiffHasUpdatedComponentsAndCheckTimes(t *testing.T) {
	var key string
	userLoader := getUserLoader()

	// Create initial empty resolution data (do not resolve any dependencies)
	uEmpty := resolvePolicy(t, language.NewPolicyNamespace(), userLoader)

	// Resolve all dependencies in policy
	policyInitial := getPolicy()
	resolvedInitial := resolvePolicy(t, policyInitial, userLoader)

	// Calculate difference against empty resolution data to update times, emulate save/load
	diff := NewPolicyResolutionDiff(resolvedInitial, uEmpty)
	resolvedInitial = emulateSaveAndLoadResolution(resolvedInitial)

	// Check creation/update times
	key = getInstanceByParams(t, "kafka", "test", []string{"platform_services"}, "component2", policyInitial, resolvedInitial).Key.GetKey()
	p := getTimesNext(t, key, resolvedInitial)
	assert.WithinDuration(t, time.Now(), p.timeNextCreated, time.Second, "Creation time should be initialized correctly for kafka")
	assert.Equal(t, p.timeNextUpdated, p.timeNextUpdated, "Update time should be equal to creation time")

	// Sleep a little bit to introduce time delay
	time.Sleep(25 * time.Millisecond)

	// Add another dependency, resolve, calculate difference against prev resolution data, emulate save/load
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
	_ = NewPolicyResolutionDiff(resolvedNew, resolvedInitial)

	// Check creation/update times
	pInit := getTimes(t, key, resolvedInitial, resolvedNew)
	assert.Equal(t, pInit.timePrevCreated, pInit.timeNextCreated, "Creation time should be carried over to remain the same")
	assert.Equal(t, pInit.timePrevUpdated, pInit.timeNextUpdated, "Update time should be carried over to remain the same")

	// Sleep a little bit to introduce time delay
	time.Sleep(25 * time.Millisecond)

	// Update user label, re-evaluate and see that component instance has changed
	userLoader = language.NewUserLoaderFromDir("../../testdata/unittests")
	userLoader.LoadUserByID("5").Labels["changinglabel"] = "newvalue"
	resolvedDependencyUpdate := resolvePolicy(t, policyNext, userLoader)

	// Get the diff
	diff = NewPolicyResolutionDiff(resolvedDependencyUpdate, resolvedNew)

	// Check that update has been performed
	verifyDiff(t, diff, 0, 0, 1, 0, 0)

	// Check creation/update times for component
	key = getInstanceByParams(t, "kafka", "prod-high", []string{"Elena"}, "component2", policyNext, resolvedNew).Key.GetKey()
	pUpdate := getTimes(t, key, resolvedNew, resolvedDependencyUpdate)
	assert.Equal(t, pUpdate.timePrevCreated, pUpdate.timeNextCreated, "Creation time should be carried over to remain the same")
	assert.True(t, pUpdate.timeNextUpdated.After(pUpdate.timePrevUpdated), "Update time should be changed")

	// Check creation/update times for service
	key = getInstanceByParams(t, "kafka", "prod-high", []string{"Elena"}, "root", policyNext, resolvedNew).Key.GetKey()
	pUpdateSvc := getTimes(t, key, resolvedNew, resolvedDependencyUpdate)
	assert.Equal(t, pUpdateSvc.timePrevCreated, pUpdateSvc.timeNextCreated, "Creation time should be carried over to remain the same")
	assert.True(t, pUpdateSvc.timeNextUpdated.After(pUpdateSvc.timePrevUpdated), "Update time should be changed for service")
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
