package diff

import (
	"github.com/Aptomi/aptomi/pkg/slinga/engine/plugin"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/slinga/language"
	"github.com/Aptomi/aptomi/pkg/slinga/object"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func BenchmarkEngine(b *testing.B) {
	t := &testing.T{}
	for i := 0; i < b.N; i++ {
		TestEngineComponentUpdateAndTimes(t)
	}
}

func TestPolicyResolveEmptyDiff(t *testing.T) {
	resolvedPrev := loadPolicyAndResolve(t)
	resolvedPrev = emulateSaveAndLoadState(resolvedPrev)

	// Get usage state next
	resolvedNext := loadPolicyAndResolve(t)

	// Calculate and verify difference
	diff := NewServiceUsageStateDiff(resolvedNext, resolvedPrev, plugin.AllPlugins())
	verifyDiff(t, diff, false, 0, 0, 0, 0, 0)
}

func TestEngineNonEmptyDiffAndApplyNoop(t *testing.T) {
	resolvedPrev := loadPolicyAndResolve(t)
	resolvedPrev = emulateSaveAndLoadState(resolvedPrev)

	// Add another dependency and resolve usage state next
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
	resolvedNext := resolvePolicy(t, nextPolicy)

	// Calculate difference
	diff := NewServiceUsageStateDiff(resolvedNext, resolvedPrev, plugin.AllPlugins())
	verifyDiff(t, diff, true, 8, 0, 0, 8, 0)
}

func TestEngineComponentUpdateAndTimes(t *testing.T) {
	var key string

	// Create initial empty state (do not resolve any dependencies)
	uEmpty := resolve.NewResolvedState(
		language.NewPolicyNamespace(),
		resolve.NewServiceUsageState(),
		getUserLoader(),
	)

	// Resolve all dependencies in policy
	resolvedInitial := loadPolicyAndResolve(t)

	// Calculate difference against empty state to update times, emulate save/load
	diff := NewServiceUsageStateDiff(resolvedInitial, uEmpty, plugin.AllPlugins())
	resolvedInitial = emulateSaveAndLoadState(resolvedInitial)

	// Check creation/update times
	key = getInstanceByParams(t, "kafka", "test", []string{"platform_services"}, "component2", resolvedInitial.Policy, resolvedInitial.State).Key.GetKey()
	p := getTimesNext(t, key, resolvedInitial.State)
	assert.WithinDuration(t, time.Now(), p.timeNextCreated, time.Second, "Creation time should be initialized correctly for kafka")
	assert.Equal(t, p.timeNextUpdated, p.timeNextUpdated, "Update time should be equal to creation time")

	// Sleep a little bit to introduce time delay
	time.Sleep(25 * time.Millisecond)

	// Add another dependency, resolve, calculate difference against prev state, emulate save/load
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
	resolvedNew := resolvePolicy(t, policyNext)
	_ = NewServiceUsageStateDiff(resolvedNew, resolvedInitial, plugin.AllPlugins())

	// Check creation/update times
	pInit := getTimes(t, key, resolvedInitial.State, resolvedNew.State)
	assert.Equal(t, pInit.timePrevCreated, pInit.timeNextCreated, "Creation time should be carried over to remain the same")
	assert.Equal(t, pInit.timePrevUpdated, pInit.timeNextUpdated, "Update time should be carried over to remain the same")

	// Sleep a little bit to introduce time delay
	time.Sleep(25 * time.Millisecond)

	// Update user label, re-evaluate and see that component instance has changed
	userLoader := language.NewUserLoaderFromDir("../../testdata/unittests")
	userLoader.LoadUserByID("5").Labels["changinglabel"] = "newvalue"
	resolvedDependencyUpdate := resolvePolicyInternal(t, resolvedNew.Policy, userLoader)

	// Get the diff
	diff = NewServiceUsageStateDiff(resolvedDependencyUpdate, resolvedNew, plugin.AllPlugins())

	// Check that update has been performed
	verifyDiff(t, diff, true, 0, 0, 1, 0, 0)

	// Check creation/update times for component
	key = getInstanceByParams(t, "kafka", "prod-high", []string{"Elena"}, "component2", resolvedNew.Policy, resolvedNew.State).Key.GetKey()
	pUpdate := getTimes(t, key, resolvedNew.State, resolvedDependencyUpdate.State)
	assert.Equal(t, pUpdate.timePrevCreated, pUpdate.timeNextCreated, "Creation time should be carried over to remain the same")
	assert.True(t, pUpdate.timeNextUpdated.After(pUpdate.timePrevUpdated), "Update time should be changed")

	// Check creation/update times for service
	key = getInstanceByParams(t, "kafka", "prod-high", []string{"Elena"}, "root", resolvedNew.Policy, resolvedNew.State).Key.GetKey()
	pUpdateSvc := getTimes(t, key, resolvedNew.State, resolvedDependencyUpdate.State)
	assert.Equal(t, pUpdateSvc.timePrevCreated, pUpdateSvc.timeNextCreated, "Creation time should be carried over to remain the same")
	assert.True(t, pUpdateSvc.timeNextUpdated.After(pUpdateSvc.timePrevUpdated), "Update time should be changed for service")
}
