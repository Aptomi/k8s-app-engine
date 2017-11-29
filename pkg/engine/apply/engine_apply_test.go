package apply

import (
	"github.com/Aptomi/aptomi/pkg/engine/actual"
	"github.com/Aptomi/aptomi/pkg/engine/diff"
	"github.com/Aptomi/aptomi/pkg/engine/progress"
	"github.com/Aptomi/aptomi/pkg/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/event"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/lang/builder"
	"github.com/Aptomi/aptomi/pkg/plugin"
	"github.com/Aptomi/aptomi/pkg/runtime"
	"github.com/Aptomi/aptomi/pkg/util"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

const (
	ResSuccess = iota
	ResError   = iota
)

func TestApplyComponentCreateSuccess(t *testing.T) {
	// resolve empty policy
	bActual := builder.NewPolicyBuilder()
	actualState := resolvePolicy(t, bActual)

	// resolve full policy
	bDesired := makePolicyBuilder()
	desiredState := resolvePolicy(t, bDesired)

	// process all actions
	actions := diff.NewPolicyResolutionDiff(desiredState, actualState).Actions

	applier := NewEngineApply(
		bDesired.Policy(),
		desiredState,
		bActual.Policy(),
		actualState,
		actual.NewNoOpActionStateUpdater(),
		bDesired.External(),
		mockRegistryFailOnComponent(false),
		actions,
		event.NewLog("test-apply", false),
		progress.NewNoop(),
	)

	// check actual state
	assert.Equal(t, 0, len(actualState.ComponentInstanceMap), "Actual state should be empty")

	// check that policy apply finished with expected results
	actualState = applyAndCheck(t, applier, ResSuccess, 0, "Successfully resolved")

	// check that actual state got updated
	assert.Equal(t, 2, len(actualState.ComponentInstanceMap), "Actual state should not be empty after apply()")
}

func TestApplyComponentCreateFailure(t *testing.T) {
	checkComponentCreateFail(t, false)
}

func TestApplyComponentCreatePanic(t *testing.T) {
	checkComponentCreateFail(t, true)
}

func checkComponentCreateFail(t *testing.T, failAsPanic bool) {
	// resolve empty policy
	bActual := builder.NewPolicyBuilder()
	actualState := resolvePolicy(t, bActual)
	// resolve full policy
	bDesired := makePolicyBuilder()
	desiredState := resolvePolicy(t, bDesired)
	// process all actions (and make component fail deployment)
	actions := diff.NewPolicyResolutionDiff(desiredState, actualState).Actions
	applier := NewEngineApply(
		bDesired.Policy(),
		desiredState,
		bActual.Policy(),
		actualState,
		actual.NewNoOpActionStateUpdater(),
		bDesired.External(),
		mockRegistryFailOnComponent(failAsPanic, bDesired.Policy().GetObjectsByKind(lang.ServiceObject.Kind)[0].(*lang.Service).Components[0].Name),
		actions,
		event.NewLog("test-apply", false),
		progress.NewNoop(),
	)
	// check actual state
	assert.Equal(t, 0, len(actualState.ComponentInstanceMap), "Actual state should be empty")

	// check that policy apply finished with expected results

	// each plugin action fails independently
	errCnt := 2
	if failAsPanic {
		// panic inside a plugin results in a single error
		errCnt = 1
	}
	actualState = applyAndCheck(t, applier, ResError, errCnt, "failed by plugin mock for component")

	// check that actual state got updated (service component exists, but no child components got deployed)
	assert.Equal(t, 1, len(actualState.ComponentInstanceMap), "Actual state should not be empty after apply()")
}

func TestDiffHasUpdatedComponentsAndCheckTimes(t *testing.T) {
	/*
		Step 1: actual = empty, desired = test policy, check = kafka update/create times
	*/

	// Create initial empty resolution data (do not resolve any dependencies)
	bActual := builder.NewPolicyBuilder()
	actualState := resolvePolicy(t, bActual)

	// Resolve all dependencies in policy
	bDesired := makePolicyBuilder()
	desiredState := resolvePolicy(t, bDesired)

	// Apply to update component times in actual state
	applier := NewEngineApply(
		bDesired.Policy(),
		desiredState,
		bActual.Policy(),
		actualState,
		actual.NewNoOpActionStateUpdater(),
		bDesired.External(),
		mockRegistryFailOnComponent(false),
		diff.NewPolicyResolutionDiff(desiredState, actualState).Actions,
		event.NewLog("test-apply", false),
		progress.NewNoop(),
	)

	// Check that policy apply finished with expected results
	updatedActualState := applyAndCheck(t, applier, ResSuccess, 0, "Successfully resolved")

	// Get key to a component
	cluster := bDesired.Policy().GetObjectsByKind(lang.ClusterObject.Kind)[0].(*lang.Cluster)
	contract := bDesired.Policy().GetObjectsByKind(lang.ContractObject.Kind)[0].(*lang.Contract)
	service := bDesired.Policy().GetObjectsByKind(lang.ServiceObject.Kind)[0].(*lang.Service)
	key := resolve.NewComponentInstanceKey(cluster, contract, contract.Contexts[0], nil, service, service.Components[0])
	keyService := key.GetParentServiceKey()

	// Check creation/update times
	times1 := getTimes(t, key.GetKey(), updatedActualState)
	assert.WithinDuration(t, time.Now(), times1.created, time.Second, "Creation time should be initialized correctly")
	assert.Equal(t, times1.updated, times1.updated, "Update time should be equal to creation time")

	actualState = updatedActualState
	bActual = bDesired

	/*
		Step 2: desired = add a dependency, check = component update/create times remained the same in actual state
	*/

	// Sleep a little bit to introduce time delay
	time.Sleep(25 * time.Millisecond)

	// Add another dependency, resolve, calculate difference against prev resolution data
	bDesiredNext := makePolicyBuilder()
	dependencyNew := bDesiredNext.AddDependency(bDesiredNext.AddUser(), contract)
	dependencyNew.Labels["param"] = "value1"

	desiredStateNext := resolvePolicy(t, bDesiredNext)
	assert.Contains(t, desiredStateNext.DependencyInstanceMap, runtime.KeyForStorable(dependencyNew), "New dependency should also be resolved")

	// Apply to update component times in actual state
	applier = NewEngineApply(
		bDesiredNext.Policy(),
		desiredStateNext,
		bActual.Policy(),
		actualState,
		actual.NewNoOpActionStateUpdater(),
		bDesiredNext.External(),
		mockRegistryFailOnComponent(false),
		diff.NewPolicyResolutionDiff(desiredStateNext, actualState).Actions,
		event.NewLog("test-apply", false),
		progress.NewNoop(),
	)

	// Check that policy apply finished with expected results
	updatedActualState = applyAndCheck(t, applier, ResSuccess, 0, "Successfully resolved")

	// Check creation/update times
	times2 := getTimes(t, key.GetKey(), updatedActualState)
	assert.Equal(t, times1.created, times2.created, "Creation time should be carried over to remain the same")
	assert.Equal(t, times1.updated, times2.updated, "Update time should be carried over to remain the same")

	actualState = updatedActualState
	bActual = bDesiredNext

	/*
		Step 3: desired = update user label, check = component update time changed
	*/
	componentTimes := getTimes(t, key.GetKey(), actualState)
	serviceTimes := getTimes(t, keyService.GetKey(), actualState)

	// Sleep a little bit to introduce time delay
	time.Sleep(25 * time.Millisecond)

	// Update labels, re-evaluate and see that component instance has changed
	for _, dependency := range bDesiredNext.Policy().GetObjectsByKind(lang.DependencyObject.Kind) {
		dependency.(*lang.Dependency).Labels["param"] = "value2"
	}
	desiredStateAfterUpdate := resolvePolicy(t, bDesiredNext)

	// Apply to update component times in actual state
	applier = NewEngineApply(
		bDesiredNext.Policy(),
		desiredStateAfterUpdate,
		bActual.Policy(),
		actualState,
		actual.NewNoOpActionStateUpdater(),
		bDesiredNext.External(),
		mockRegistryFailOnComponent(false),
		diff.NewPolicyResolutionDiff(desiredStateAfterUpdate, actualState).Actions,
		event.NewLog("test-apply", false),
		progress.NewNoop(),
	)

	// Check that policy apply finished with expected results
	updatedActualState = applyAndCheck(t, applier, ResSuccess, 0, "Successfully resolved")

	// Check creation/update times for component
	componentTimesUpdated := getTimes(t, key.GetKey(), updatedActualState)
	assert.Equal(t, componentTimes.created, componentTimesUpdated.created, "Creation time for component should be carried over to remain the same")
	assert.True(t, componentTimesUpdated.updated.After(componentTimes.updated), "Update time for component should be changed")

	// Check creation/update times for service
	serviceTimesUpdated := getTimes(t, keyService.GetKey(), updatedActualState)
	assert.Equal(t, serviceTimes.created, serviceTimesUpdated.created, "Creation time for parent service should be carried over to remain the same")
	assert.True(t, serviceTimesUpdated.updated.After(serviceTimes.updated), "Update time for parent service should be changed")
}

/*
	Helpers
*/

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
	eventLog := event.NewLog("test-resolve", false)
	resolver := resolve.NewPolicyResolver(b.Policy(), b.External(), eventLog)
	result, err := resolver.ResolveAllDependencies()
	if !assert.NoError(t, err, "Policy should be resolved without errors") {
		hook := &event.HookConsole{}
		eventLog.Save(hook)
		t.FailNow()
	}

	return result
}

func applyAndCheck(t *testing.T, apply *EngineApply, expectedResult int, errorCnt int, expectedMessage string) *resolve.PolicyResolution {
	t.Helper()
	actualState, err := apply.Apply()

	if !assert.Equal(t, expectedResult != ResError, err == nil, "Apply status (success vs. error)") {
		// print log into stdout and exit
		hook := &event.HookConsole{}
		apply.eventLog.Save(hook)
		t.FailNow()
	}

	if expectedResult == ResError {
		// check for error messages
		verifier := event.NewLogVerifier(expectedMessage, expectedResult == ResError)
		apply.eventLog.Save(verifier)
		if !assert.Equal(t, errorCnt, verifier.MatchedErrorsCount(), "Apply event log should have correct number of messages containing words: "+expectedMessage) {
			hook := &event.HookConsole{}
			apply.eventLog.Save(hook)
			t.FailNow()
		}
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
	if !assert.True(t, ok, "Component instance exists in resolution data: "+key) {
		t.FailNow()
	}
	return instance
}

func mockRegistryFailOnComponent(failAsPanic bool, failComponents ...string) plugin.Registry {
	return &plugin.MockRegistry{
		DeployPlugin: &plugin.MockDeployPluginFailComponents{
			FailComponents: failComponents,
			FailAsPanic:    failAsPanic,
		},
		PostProcessPlugin: &plugin.MockPostProcessPlugin{},
	}
}
