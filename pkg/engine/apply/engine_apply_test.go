package apply

import (
	"testing"
	"time"

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
	"github.com/Aptomi/aptomi/pkg/util"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
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
		actual.NewNoOpActionStateUpdater(actualState),
		desired.external(),
		mockRegistry(true, false),
		diff.NewPolicyResolutionDiff(desired.resolution(), actualState).ActionPlan,
		event.NewLog(logrus.DebugLevel, "test-apply"),
		action.NewApplyResultUpdaterImpl(),
	)

	// check actual state
	assert.Equal(t, 0, len(actualState.ComponentInstanceMap), "Actual state should be empty")

	// check that policy apply finished with expected results
	actualState = applyAndCheck(t, applier, action.ApplyResult{Success: 4, Failed: 0, Skipped: 0})

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
		actual.NewNoOpActionStateUpdater(actualState),
		desired.external(),
		mockRegistry(false, failAsPanic),
		diff.NewPolicyResolutionDiff(desired.resolution(), actualState).ActionPlan,
		event.NewLog(logrus.DebugLevel, "test-apply"),
		action.NewApplyResultUpdaterImpl(),
	)
	// check actual state
	assert.Equal(t, 0, len(actualState.ComponentInstanceMap), "Actual state should be empty")

	// check for errors
	actualState = applyAndCheck(t, applier, action.ApplyResult{Success: 0, Failed: 1, Skipped: 3})

	// check that actual state didn't get updated
	assert.Equal(t, 0, len(actualState.ComponentInstanceMap), "Actual state should not be touched by apply()")
}

func TestDiffHasUpdatedComponentsAndCheckTimes(t *testing.T) {
	/*
		Step 1: actual = empty, desired = test policy, check = claim update/create times
	*/

	// Create initial empty policy & resolution data
	empty := newTestData(t, builder.NewPolicyBuilder())
	actualState := empty.resolution()

	// Generate policy and resolve all claims in policy
	desired := newTestData(t, makePolicyBuilder())

	// Apply to update component times in actual state
	applier := NewEngineApply(
		desired.policy(),
		desired.resolution(),
		actual.NewNoOpActionStateUpdater(actualState),
		desired.external(),
		mockRegistry(true, false),
		diff.NewPolicyResolutionDiff(desired.resolution(), actualState).ActionPlan,
		event.NewLog(logrus.DebugLevel, "test-apply"),
		action.NewApplyResultUpdaterImpl(),
	)

	// Check that policy apply finished with expected results
	actualState = applyAndCheck(t, applier, action.ApplyResult{Success: 4, Failed: 0, Skipped: 0})

	// Get key to a component
	cluster := desired.policy().GetObjectsByKind(lang.ClusterObject.Kind)[0].(*lang.Cluster)    // nolint: errcheck
	contract := desired.policy().GetObjectsByKind(lang.ContractObject.Kind)[0].(*lang.Contract) // nolint: errcheck
	service := desired.policy().GetObjectsByKind(lang.ServiceObject.Kind)[0].(*lang.Service)    // nolint: errcheck
	key := resolve.NewComponentInstanceKey(cluster, "k8ns", contract, contract.Contexts[0], nil, service, service.Components[0])
	keyService := key.GetParentServiceKey()

	// Check that original claim was resolved successfully
	claim := desired.policy().GetObjectsByKind(lang.ClaimObject.Kind)[0].(*lang.Claim) // nolint: errcheck
	assert.True(t, desired.resolution().GetClaimResolution(claim).Resolved, "Original claim should be resolved successfully")

	// Check creation/update times
	times1 := getTimes(t, key.GetKey(), actualState)
	assert.WithinDuration(t, time.Now(), times1.created, time.Second, "Creation time should be initialized correctly")
	assert.Equal(t, times1.updated, times1.updated, "Update time should be equal to creation time")

	/*
		Step 2: desired = add a claim, check = component update/create times remained the same in actual state
	*/

	// Sleep a little bit to introduce time delay
	time.Sleep(25 * time.Millisecond)

	// Add another claim (with the same label, so code parameters won't change), resolve, calculate difference against prev resolution data
	desiredNext := newTestData(t, makePolicyBuilder())
	claimNew := desiredNext.pBuilder.AddClaim(desiredNext.pBuilder.AddUser(), contract)
	claimNew.Labels["param"] = "value1"

	// Check that both claims were resolved successfully
	assert.True(t, desiredNext.resolution().GetClaimResolution(claim).Resolved, "Original claim should be resolved successfully")
	assert.True(t, desiredNext.resolution().GetClaimResolution(claimNew).Resolved, "Additional claim should be resolved successfully")

	// Apply to update component times in actual state
	applier = NewEngineApply(
		desiredNext.policy(),
		desiredNext.resolution(),
		actual.NewNoOpActionStateUpdater(actualState),
		desiredNext.external(),
		mockRegistry(true, false),
		diff.NewPolicyResolutionDiff(desiredNext.resolution(), actualState).ActionPlan,
		event.NewLog(logrus.DebugLevel, "test-apply"),
		action.NewApplyResultUpdaterImpl(),
	)

	// Check that policy apply finished with expected results
	actualState = applyAndCheck(t, applier, action.ApplyResult{Success: 2, Failed: 0, Skipped: 0})

	// Check creation/update times for the original claim
	times2 := getTimes(t, key.GetKey(), actualState)
	assert.Equal(t, times1.created, times2.created, "Creation time should be preserved (i.e. remain the same)")
	assert.True(t, times2.updated.After(times1.updated), "Update time should be changed (because new claim is attached to a component)")

	/*
		Step 3: desired = update user label, check = component update time changed
	*/
	componentTimes := getTimes(t, key.GetKey(), actualState)
	serviceTimes := getTimes(t, keyService.GetKey(), actualState)

	// Sleep a little bit to introduce time delay
	time.Sleep(25 * time.Millisecond)

	// Update labels, re-evaluate and see that component instance has changed
	desiredNextAfterUpdate := newTestData(t, desiredNext.pBuilder)
	for _, claim := range desiredNextAfterUpdate.policy().GetObjectsByKind(lang.ClaimObject.Kind) {
		claim.(*lang.Claim).Labels["param"] = "value2"
	}

	// Apply to update component times in actual state
	applier = NewEngineApply(
		desiredNextAfterUpdate.policy(),
		desiredNextAfterUpdate.resolution(),
		actual.NewNoOpActionStateUpdater(actualState),
		desiredNextAfterUpdate.external(),
		mockRegistry(true, false),
		diff.NewPolicyResolutionDiff(desiredNextAfterUpdate.resolution(), actualState).ActionPlan,
		event.NewLog(logrus.DebugLevel, "test-apply"),
		action.NewApplyResultUpdaterImpl(),
	)

	// Check that policy apply finished with expected results
	actualState = applyAndCheck(t, applier, action.ApplyResult{Success: 2, Failed: 0, Skipped: 0})

	// Check that both claims were resolved successfully
	assert.True(t, desiredNextAfterUpdate.resolution().GetClaimResolution(claim).Resolved, "Original claim should be resolved successfully")
	assert.True(t, desiredNextAfterUpdate.resolution().GetClaimResolution(claimNew).Resolved, "Additional claim should be resolved successfully")

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
		actual.NewNoOpActionStateUpdater(actualState),
		generated.external(),
		mockRegistry(true, false),
		diff.NewPolicyResolutionDiff(generated.resolution(), actualState).ActionPlan,
		event.NewLog(logrus.DebugLevel, "test-apply"),
		action.NewApplyResultUpdaterImpl(),
	)

	// Check that policy apply finished with expected results
	actualState = applyAndCheck(t, applier, action.ApplyResult{Success: 4, Failed: 0, Skipped: 0})

	assert.Equal(t, 2, len(actualState.ComponentInstanceMap), "Actual state have components instances transferred to it")

	// Reset policy back to empty
	reset := newTestData(t, builder.NewPolicyBuilder())

	// Run apply to update actual state
	applierNext := NewEngineApply(
		reset.policy(),
		reset.resolution(),
		actual.NewNoOpActionStateUpdater(actualState),
		generated.external(),
		mockRegistry(true, false),
		diff.NewPolicyResolutionDiff(reset.resolution(), actualState).ActionPlan,
		event.NewLog(logrus.DebugLevel, "test-apply"),
		action.NewApplyResultUpdaterImpl(),
	)

	// detach successful, deletion fails
	actualState = applyAndCheck(t, applierNext, action.ApplyResult{Success: 2, Failed: 2, Skipped: 0})
	assert.Equal(t, 2, len(actualState.ComponentInstanceMap), "Actual state should still have component instances after actions failing")
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
				"param": "{{ .Labels.param }}",
				"debug": "{{ .Labels.target }}",
			},
			nil,
		),
	)
	contract := b.AddContract(service, b.CriteriaTrue())

	// add rule to set cluster
	clusterObj := b.AddCluster()
	b.AddRule(b.CriteriaTrue(), b.RuleActions(lang.NewLabelOperationsSetSingleLabel(lang.LabelTarget, clusterObj.Name)))

	// add claim
	claim := b.AddClaim(b.AddUser(), contract)
	claim.Labels["param"] = "value1"

	return b
}

func resolvePolicy(t *testing.T, b *builder.PolicyBuilder) *resolve.PolicyResolution {
	t.Helper()
	eventLog := event.NewLog(logrus.DebugLevel, "test-resolve")
	resolver := resolve.NewPolicyResolver(b.Policy(), b.External(), eventLog)
	result := resolver.ResolveAllClaims()

	claims := b.Policy().GetObjectsByKind(lang.ClaimObject.Kind)
	for _, claim := range claims {
		if !assert.True(t, result.GetClaimResolution(claim.(*lang.Claim)).Resolved, "Claim resolution status should be correct for %v", claim) {
			hook := event.NewHookConsole(logrus.DebugLevel)
			eventLog.Save(hook)
			t.FailNow()
		}
	}

	return result
}

func applyAndCheck(t *testing.T, apply *EngineApply, expectedResult action.ApplyResult) *resolve.PolicyResolution {
	t.Helper()
	actualState, result := apply.Apply(50)

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
