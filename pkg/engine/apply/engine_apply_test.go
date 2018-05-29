package apply

import (
	"github.com/Aptomi/aptomi/pkg/config"
	"github.com/Aptomi/aptomi/pkg/engine/actual"
	"github.com/Aptomi/aptomi/pkg/engine/apply/action"
	"github.com/Aptomi/aptomi/pkg/engine/diff"
	"github.com/Aptomi/aptomi/pkg/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/event"
	"github.com/Aptomi/aptomi/pkg/external"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/lang/builder"
	"github.com/Aptomi/aptomi/pkg/plugin"
	"github.com/Aptomi/aptomi/pkg/plugin/fake"
	"github.com/Aptomi/aptomi/pkg/runtime"
	"github.com/Aptomi/aptomi/pkg/util"
	"github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestApplyComponentCreateSuccess(t *testing.T) {
	// resolve empty policy
	empty := newTestData(t, builder.NewPolicyBuilder())
	actualState := empty.resolution()

	// resolve full policy
	desired := newTestData(t, makePolicyBuilder())

	// apply changes
	applier := NewEngineApply(
		desired.policy(),
		desired.resolution(),
		actualState,
		actual.NewNoOpActionStateUpdater(),
		desired.external(),
		mockRegistry(true, false),
		diff.NewPolicyResolutionDiff(desired.resolution(), actualState).ActionPlan,
		event.NewLog(logrus.DebugLevel, "test-apply"),
		action.NewApplyResultUpdaterImpl(),
	)

	// check actual state
	assert.Equal(t, 0, len(actualState.ComponentInstanceMap), "Actual state should be empty")

	// check that policy apply finished with expected results
	actualState = applyAndCheck(t, applier, action.ApplyResult{Success: 5, Failed: 0, Skipped: 0})

	// check that actual state got updated
	assert.Equal(t, 2, len(actualState.ComponentInstanceMap), "Actual state should not be empty after apply()")
}

func TestApplyComponentCreateFailure(t *testing.T) {
	checkApplyComponentCreateFail(t, false)
}

func TestApplyComponentCreatePanic(t *testing.T) {
	checkApplyComponentCreateFail(t, true)
}

func checkApplyComponentCreateFail(t *testing.T, failAsPanic bool) {
	// resolve empty policy
	empty := newTestData(t, builder.NewPolicyBuilder())
	actualState := empty.resolution()

	// resolve full policy
	desired := newTestData(t, makePolicyBuilder())

	// process all actions (and make component fail deployment)
	applier := NewEngineApply(
		desired.policy(),
		desired.resolution(),
		actualState,
		actual.NewNoOpActionStateUpdater(),
		desired.external(),
		mockRegistry(false, failAsPanic),
		diff.NewPolicyResolutionDiff(desired.resolution(), actualState).ActionPlan,
		event.NewLog(logrus.DebugLevel, "test-apply"),
		action.NewApplyResultUpdaterImpl(),
	)
	// check actual state
	assert.Equal(t, 0, len(actualState.ComponentInstanceMap), "Actual state should be empty")

	// check for errors
	actualState = applyAndCheck(t, applier, action.ApplyResult{Success: 0, Failed: 1, Skipped: 4})

	// check that actual state didn't get updated
	assert.Equal(t, 0, len(actualState.ComponentInstanceMap), "Actual state should not be touched by apply()")
}

func TestDiffHasUpdatedComponentsAndCheckTimes(t *testing.T) {
	/*
		Step 1: actual = empty, desired = test policy, check = dependency update/create times
	*/

	// Create initial empty policy & resolution data
	empty := newTestData(t, builder.NewPolicyBuilder())
	actualState := empty.resolution()

	// Generate policy and resolve all dependencies in policy
	desired := newTestData(t, makePolicyBuilder())

	// Apply to update component times in actual state
	applier := NewEngineApply(
		desired.policy(),
		desired.resolution(),
		actualState,
		actual.NewNoOpActionStateUpdater(),
		desired.external(),
		mockRegistry(true, false),
		diff.NewPolicyResolutionDiff(desired.resolution(), actualState).ActionPlan,
		event.NewLog(logrus.DebugLevel, "test-apply"),
		action.NewApplyResultUpdaterImpl(),
	)

	// Check that policy apply finished with expected results
	actualState = applyAndCheck(t, applier, action.ApplyResult{Success: 5, Failed: 0, Skipped: 0})

	// Get key to a component
	cluster := desired.policy().GetObjectsByKind(lang.ClusterObject.Kind)[0].(*lang.Cluster)
	contract := desired.policy().GetObjectsByKind(lang.ContractObject.Kind)[0].(*lang.Contract)
	service := desired.policy().GetObjectsByKind(lang.ServiceObject.Kind)[0].(*lang.Service)
	key := resolve.NewComponentInstanceKey(cluster, contract, contract.Contexts[0], nil, service, service.Components[0])
	keyService := key.GetParentServiceKey()

	// Check that original dependency was resolved successfully
	dependency := desired.policy().GetObjectsByKind(lang.DependencyObject.Kind)[0].(*lang.Dependency)
	assert.Contains(t, desired.resolution().GetDependencyInstanceMap(), runtime.KeyForStorable(dependency), "Original dependency should be present in policy resolution")
	assert.True(t, desired.resolution().GetDependencyInstanceMap()[runtime.KeyForStorable(dependency)].Resolved, "Original dependency should be resolved successfully")

	// Check creation/update times
	times1 := getTimes(t, key.GetKey(), actualState)
	assert.WithinDuration(t, time.Now(), times1.created, time.Second, "Creation time should be initialized correctly")
	assert.Equal(t, times1.updated, times1.updated, "Update time should be equal to creation time")

	/*
		Step 2: desired = add a dependency, check = component update/create times remained the same in actual state
	*/

	// Sleep a little bit to introduce time delay
	time.Sleep(25 * time.Millisecond)

	// Add another dependency (with the same label, so code parameters won't change), resolve, calculate difference against prev resolution data
	desiredNext := newTestData(t, makePolicyBuilder())
	dependencyNew := desiredNext.pBuilder.AddDependency(desiredNext.pBuilder.AddUser(), contract)
	dependencyNew.Labels["param"] = "value1"

	// Check that both dependencies were resolved successfully
	assert.Contains(t, desiredNext.resolution().GetDependencyInstanceMap(), runtime.KeyForStorable(dependency), "Original dependency should be present in policy resolution")
	assert.True(t, desiredNext.resolution().GetDependencyInstanceMap()[runtime.KeyForStorable(dependency)].Resolved, "Original dependency should be resolved successfully")
	assert.Contains(t, desiredNext.resolution().GetDependencyInstanceMap(), runtime.KeyForStorable(dependencyNew), "Additional dependency should be present in policy resolution")
	assert.True(t, desiredNext.resolution().GetDependencyInstanceMap()[runtime.KeyForStorable(dependencyNew)].Resolved, "Additional dependency should be resolved successfully")

	// Apply to update component times in actual state
	applier = NewEngineApply(
		desiredNext.policy(),
		desiredNext.resolution(),
		actualState,
		actual.NewNoOpActionStateUpdater(),
		desiredNext.external(),
		mockRegistry(true, false),
		diff.NewPolicyResolutionDiff(desiredNext.resolution(), actualState).ActionPlan,
		event.NewLog(logrus.DebugLevel, "test-apply"),
		action.NewApplyResultUpdaterImpl(),
	)

	// Check that policy apply finished with expected results
	actualState = applyAndCheck(t, applier, action.ApplyResult{Success: 2, Failed: 0, Skipped: 0})

	// Check creation/update times for the original dependency
	times2 := getTimes(t, key.GetKey(), actualState)
	assert.Equal(t, times1.created, times2.created, "Creation time should be preserved (i.e. remain the same)")
	assert.True(t, times2.updated.After(times1.updated), "Update time should be changed (because new dependency is attached to a component)")

	/*
		Step 3: desired = update user label, check = component update time changed
	*/
	componentTimes := getTimes(t, key.GetKey(), actualState)
	serviceTimes := getTimes(t, keyService.GetKey(), actualState)

	// Sleep a little bit to introduce time delay
	time.Sleep(25 * time.Millisecond)

	// Update labels, re-evaluate and see that component instance has changed
	desiredNextAfterUpdate := newTestData(t, desiredNext.pBuilder)
	for _, dependency := range desiredNextAfterUpdate.policy().GetObjectsByKind(lang.DependencyObject.Kind) {
		dependency.(*lang.Dependency).Labels["param"] = "value2"
	}

	// Apply to update component times in actual state
	applier = NewEngineApply(
		desiredNextAfterUpdate.policy(),
		desiredNextAfterUpdate.resolution(),
		actualState,
		actual.NewNoOpActionStateUpdater(),
		desiredNextAfterUpdate.external(),
		mockRegistry(true, false),
		diff.NewPolicyResolutionDiff(desiredNextAfterUpdate.resolution(), actualState).ActionPlan,
		event.NewLog(logrus.DebugLevel, "test-apply"),
		action.NewApplyResultUpdaterImpl(),
	)

	// Check that policy apply finished with expected results
	actualState = applyAndCheck(t, applier, action.ApplyResult{Success: 3, Failed: 0, Skipped: 0})

	// Check that both dependencies were resolved successfully
	assert.Contains(t, desiredNextAfterUpdate.resolution().GetDependencyInstanceMap(), runtime.KeyForStorable(dependency), "Original dependency should be present in policy resolution")
	assert.True(t, desiredNextAfterUpdate.resolution().GetDependencyInstanceMap()[runtime.KeyForStorable(dependency)].Resolved, "Original dependency should be resolved successfully")
	assert.Contains(t, desiredNextAfterUpdate.resolution().GetDependencyInstanceMap(), runtime.KeyForStorable(dependencyNew), "Additional dependency should be present in policy resolution")
	assert.True(t, desiredNextAfterUpdate.resolution().GetDependencyInstanceMap()[runtime.KeyForStorable(dependencyNew)].Resolved, "Additional dependency should be resolved successfully")

	// Check creation/update times for component
	componentTimesUpdated := getTimes(t, key.GetKey(), actualState)
	assert.Equal(t, componentTimes.created, componentTimesUpdated.created, "Creation time for component should be preserved (i.e. remain the same)")
	assert.True(t, componentTimesUpdated.updated.After(componentTimes.updated), "Update time for component should be changed (because component code param is changed)")

	// Check creation/update times for service
	serviceTimesUpdated := getTimes(t, keyService.GetKey(), actualState)
	assert.Equal(t, serviceTimes.created, serviceTimesUpdated.created, "Creation time for parent service should be preserved (i.e. remain the same)")
	assert.True(t, serviceTimesUpdated.updated.After(serviceTimes.updated), "Update time for parent service should be changed (because component code param is changed)")
}

func TestDeletePolicyObjectsWhileComponentInstancesAreStillRunningFails(t *testing.T) {
	// Start with empty actual state & empty policy
	empty := newTestData(t, builder.NewPolicyBuilder())
	actualState := empty.resolution()
	assert.Equal(t, 0, len(empty.resolution().ComponentInstanceMap), "Initial state should not have any components")
	assert.Equal(t, 0, len(actualState.ComponentInstanceMap), "Actual state should not have any components at this point")

	// Generate policy
	generated := newTestData(t, makePolicyBuilder())
	assert.Equal(t, 2, len(generated.resolution().ComponentInstanceMap), "Desired state should not be empty")

	// Run apply to update actual state
	applier := NewEngineApply(
		generated.policy(),
		generated.resolution(),
		actualState,
		actual.NewNoOpActionStateUpdater(),
		generated.external(),
		mockRegistry(true, false),
		diff.NewPolicyResolutionDiff(generated.resolution(), actualState).ActionPlan,
		event.NewLog(logrus.DebugLevel, "test-apply"),
		action.NewApplyResultUpdaterImpl(),
	)

	// Check that policy apply finished with expected results
	actualState = applyAndCheck(t, applier, action.ApplyResult{Success: 5, Failed: 0, Skipped: 0})

	assert.Equal(t, 2, len(actualState.ComponentInstanceMap), "Actual state should have populated with components at this point")

	// Reset policy back to empty
	reset := newTestData(t, builder.NewPolicyBuilder())

	// Run apply to update actual state
	applierNext := NewEngineApply(
		reset.policy(),
		reset.resolution(),
		actualState,
		actual.NewNoOpActionStateUpdater(),
		generated.external(),
		mockRegistry(true, false),
		diff.NewPolicyResolutionDiff(reset.resolution(), actualState).ActionPlan,
		event.NewLog(logrus.DebugLevel, "test-apply"),
		action.NewApplyResultUpdaterImpl(),
	)

	// delete/detach, delete/detach, endpoints/endpoints - 6 actions failed in total
	actualState = applyAndCheck(t, applierNext, action.ApplyResult{Success: 1, Failed: 1, Skipped: 2})
	assert.Equal(t, 2, len(actualState.ComponentInstanceMap), "Actual state should be intact after actions failing")
}

/*
	Helpers
*/

// Utility data structure for creating & resolving policy via builder in unit tests
type testData struct {
	t        *testing.T
	pBuilder *builder.PolicyBuilder
	resolved *resolve.PolicyResolution
}

func newTestData(t *testing.T, pBuilder *builder.PolicyBuilder) *testData {
	return &testData{t: t, pBuilder: pBuilder}
}

func (td *testData) policy() *lang.Policy {
	return td.pBuilder.Policy()
}

func (td *testData) resolution() *resolve.PolicyResolution {
	if td.resolved == nil {
		td.resolved = resolvePolicy(td.t, td.pBuilder)
	}
	return td.resolved
}

func (td *testData) external() *external.Data {
	return td.pBuilder.External()
}

func makePolicyBuilder() *builder.PolicyBuilder {
	b := builder.NewPolicyBuilder()

	// create a service
	service := b.AddService()
	b.AddServiceComponent(service,
		b.CodeComponent(
			util.NestedParameterMap{
				"param":   "{{ .Labels.param }}",
				"cluster": "{{ .Labels.cluster }}",
			},
			nil,
		),
	)
	contract := b.AddContract(service, b.CriteriaTrue())

	// add rule to set cluster
	clusterObj := b.AddCluster()
	b.AddRule(b.CriteriaTrue(), b.RuleActions(lang.NewLabelOperationsSetSingleLabel(lang.LabelCluster, clusterObj.Name)))

	// add dependency
	dependency := b.AddDependency(b.AddUser(), contract)
	dependency.Labels["param"] = "value1"

	return b
}

func resolvePolicy(t *testing.T, b *builder.PolicyBuilder) *resolve.PolicyResolution {
	t.Helper()
	eventLog := event.NewLog(logrus.DebugLevel, "test-resolve")
	resolver := resolve.NewPolicyResolver(b.Policy(), b.External(), eventLog)
	result := resolver.ResolveAllDependencies()
	if !assert.True(t, result.AllDependenciesResolvedSuccessfully(), "All dependencies should be resolved successfully") {
		hook := event.NewHookConsole(logrus.DebugLevel)
		eventLog.Save(hook)
		t.FailNow()
	}

	return result
}

func applyAndCheck(t *testing.T, apply *EngineApply, expectedResult action.ApplyResult) *resolve.PolicyResolution {
	t.Helper()
	actualState, result := apply.Apply()

	ok := assert.Equal(t, expectedResult.Success, result.Success, "Number of successfully executed actions")
	ok = ok && assert.Equal(t, expectedResult.Failed, result.Failed, "Number of failed actions")
	ok = ok && assert.Equal(t, expectedResult.Skipped, result.Skipped, "Number of skipped actions")
	ok = ok && assert.Equal(t, expectedResult.Success+expectedResult.Failed+expectedResult.Skipped, result.Total, "Number of total actions")

	if !ok {
		// print log into stdout and exit
		hook := event.NewHookConsole(logrus.DebugLevel)
		apply.eventLog.Save(hook)
		t.FailNow()
	}

	return actualState
}

type componentTimes struct {
	created time.Time
	updated time.Time
}

func getTimes(t *testing.T, key string, u2 *resolve.PolicyResolution) componentTimes {
	t.Helper()
	return componentTimes{
		created: getInstanceInternal(t, key, u2).CreatedAt,
		updated: getInstanceInternal(t, key, u2).UpdatedAt,
	}
}

func getInstanceInternal(t *testing.T, key string, resolution *resolve.PolicyResolution) *resolve.ComponentInstance {
	t.Helper()
	instance, ok := resolution.ComponentInstanceMap[key]
	if !assert.True(t, ok, "Component instance exists in resolution data: %s", key) {
		t.FailNow()
	}
	return instance
}

func mockRegistry(applySuccess, failAsPanic bool) plugin.Registry {
	clusterTypes := make(map[string]plugin.ClusterPluginConstructor)
	codeTypes := make(map[string]map[string]plugin.CodePluginConstructor)

	clusterTypes["kubernetes"] = func(cluster *lang.Cluster, cfg config.Plugins) (plugin.ClusterPlugin, error) {
		return fake.NewNoOpClusterPlugin(0), nil
	}

	codeTypes["kubernetes"] = make(map[string]plugin.CodePluginConstructor)
	codeTypes["kubernetes"]["helm"] = func(cluster plugin.ClusterPlugin, cfg config.Plugins) (plugin.CodePlugin, error) {
		if applySuccess {
			return fake.NewNoOpCodePlugin(0), nil
		}
		return fake.NewFailCodePlugin(failAsPanic), nil
	}

	return plugin.NewRegistry(config.Plugins{}, clusterTypes, codeTypes)
}
