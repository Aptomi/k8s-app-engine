package apply

import (
	"github.com/Aptomi/aptomi/pkg/slinga/engine/diff"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/plugin"
	"github.com/Aptomi/aptomi/pkg/slinga/language"
	"github.com/Aptomi/aptomi/pkg/slinga/object"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

const (
	ResSuccess = iota
	ResError   = iota
)

func TestApplyCreateSuccess(t *testing.T) {
	externalData := getExternalData()

	// resolve empty policy
	actualPolicy := language.NewPolicyNamespace()
	actualState := resolvePolicy(t, actualPolicy, externalData)

	// resolve full policy
	desiredPolicy := getPolicy()
	desiredState := resolvePolicy(t, desiredPolicy, externalData)

	// make plugin to successfully process all components
	pluginApply := NewEnginePluginImpl([]string{})

	// process all actions
	actions := diff.NewPolicyResolutionDiff(desiredState, actualState).Actions
	plugins := []plugin.EnginePlugin{pluginApply}
	apply := NewEngineApply(
		desiredPolicy,
		desiredState,
		actualPolicy,
		actualState,
		externalData,
		actions,
		plugins,
	)

	// check actual state
	assert.Equal(t, 0, len(actualState.ComponentInstanceMap), "Actual state should be empty")

	// check that policy apply finished with expected results
	actualState = applyAndCheck(t, apply, ResSuccess, 0, "")

	// check that actual state got updated
	assert.Equal(t, 16, len(actualState.ComponentInstanceMap), "Actual state should be empty")
}

func TestApplyCreateFailure(t *testing.T) {
	externalData := getExternalData()

	// resolve empty policy
	actualPolicy := language.NewPolicyNamespace()
	actualState := resolvePolicy(t, actualPolicy, externalData)

	// resolve full policy
	desiredPolicy := getPolicy()
	desiredState := resolvePolicy(t, desiredPolicy, externalData)

	// make plugin to successfully process all components, while failing all instances of component2
	pluginApplyFailComponent2 := NewEnginePluginImpl([]string{"component2"})

	// process all actions
	actions := diff.NewPolicyResolutionDiff(desiredState, actualState).Actions
	plugins := []plugin.EnginePlugin{pluginApplyFailComponent2}
	apply := NewEngineApply(
		desiredPolicy,
		desiredState,
		actualPolicy,
		actualState,
		externalData,
		actions,
		plugins,
	)

	// check actual state
	assert.Equal(t, 0, len(actualState.ComponentInstanceMap), "Actual state should be empty")

	// check that policy apply finished with expected results
	actualState = applyAndCheck(t, apply, ResError, 4, "Apply failed for component")

	// check that actual state got updated
	assert.Equal(t, 12, len(actualState.ComponentInstanceMap), "Actual state should be empty")
}

func TestDiffHasUpdatedComponentsAndCheckTimes(t *testing.T) {
	var key string
	externalData := getExternalData()

	/*
		Step 1: actual = empty, desired = unit test policy, check = kafka update/create times
	*/

	// Create initial empty resolution data (do not resolve any dependencies)
	actualPolicy := language.NewPolicyNamespace()
	actualState := resolvePolicy(t, actualPolicy, externalData)

	// Resolve all dependencies in policy
	desiredPolicy := getPolicy()
	desiredState := resolvePolicy(t, desiredPolicy, externalData)

	// Apply to update component times in actual state
	apply := NewEngineApply(
		desiredPolicy,
		desiredState,
		actualPolicy,
		actualState,
		externalData,
		diff.NewPolicyResolutionDiff(desiredState, actualState).Actions,
		[]plugin.EnginePlugin{},
	)

	// Check that policy apply finished with expected results
	updatedActualState := applyAndCheck(t, apply, ResSuccess, 0, "")

	// Check creation/update times
	key = getInstanceKey("kafka", "test", []string{"platform_services"}, "component2", desiredPolicy)
	kafkaTimes1 := getTimes(t, key, updatedActualState)
	assert.WithinDuration(t, time.Now(), kafkaTimes1.created, time.Second, "Creation time should be initialized correctly for kafka")
	assert.Equal(t, kafkaTimes1.updated, kafkaTimes1.updated, "Update time should be equal to creation time")

	actualState = updatedActualState
	actualPolicy = desiredPolicy

	/*
		Step 2: desired = add a dependency, check = component update/create times remained the same in actual state
	*/

	// Sleep a little bit to introduce time delay
	time.Sleep(25 * time.Millisecond)

	// Add another dependency, resolve, calculate difference against prev resolution data, emulate save/load
	desiredPolicyNext := getPolicy()
	dependencyNew := &language.Dependency{
		Metadata: object.Metadata{
			Namespace: "main",
			Name:      "dep_id_5",
		},
		UserID:  "5",
		Service: "kafka",
	}
	desiredPolicyNext.Dependencies.AddDependency(dependencyNew)
	desiredStateNext := resolvePolicy(t, desiredPolicyNext, externalData)
	assert.True(t, dependencyNew.Resolved, "New dependency should be resolved")

	// Apply to update component times in actual state
	apply = NewEngineApply(
		desiredPolicyNext,
		desiredStateNext,
		actualPolicy,
		actualState,
		externalData,
		diff.NewPolicyResolutionDiff(desiredStateNext, actualState).Actions,
		[]plugin.EnginePlugin{},
	)

	// Check that policy apply finished with expected results
	updatedActualState = applyAndCheck(t, apply, ResSuccess, 0, "")

	// Check creation/update times
	kafkaTimes2 := getTimes(t, key, updatedActualState)
	assert.Equal(t, kafkaTimes1.created, kafkaTimes2.created, "Creation time should be carried over to remain the same")
	assert.Equal(t, kafkaTimes1.updated, kafkaTimes2.updated, "Update time should be carried over to remain the same")

	actualState = updatedActualState
	actualPolicy = desiredPolicy

	/*
		Step 3: desired = update user label, check = component update time changed
	*/

	keyComponent := getInstanceKey("kafka", "prod-high", []string{"Elena"}, "component2", desiredPolicyNext)
	componentTimes := getTimes(t, keyComponent, actualState)
	keyService := getInstanceKey("kafka", "prod-high", []string{"Elena"}, "root", desiredPolicyNext)
	serviceTimes := getTimes(t, keyService, actualState)

	// Sleep a little bit to introduce time delay
	time.Sleep(25 * time.Millisecond)

	// Update user label, re-evaluate and see that component instance has changed
	externalData.UserLoader.LoadUserByID("5").Labels["changinglabel"] = "newvalue"
	desiredStateAfterUpdate := resolvePolicy(t, desiredPolicyNext, externalData)

	// Apply to update component times in actual state
	apply = NewEngineApply(
		desiredPolicyNext,
		desiredStateAfterUpdate,
		actualPolicy,
		actualState,
		externalData,
		diff.NewPolicyResolutionDiff(desiredStateAfterUpdate, actualState).Actions,
		[]plugin.EnginePlugin{},
	)

	// Check that policy apply finished with expected results
	updatedActualState = applyAndCheck(t, apply, ResSuccess, 0, "")

	// Check creation/update times for component
	componentTimesUpdated := getTimes(t, keyComponent, updatedActualState)
	assert.Equal(t, componentTimes.created, componentTimesUpdated.created, "Creation time for component should be carried over to remain the same")
	assert.True(t, componentTimesUpdated.updated.After(componentTimes.updated), "Update time for component should be changed")

	// Check creation/update times for service
	serviceTimesUpdated := getTimes(t, keyService, updatedActualState)
	assert.Equal(t, serviceTimes.created, serviceTimesUpdated.created, "Creation time for parent service should be carried over to remain the same")
	assert.True(t, serviceTimesUpdated.updated.After(serviceTimes.updated), "Update time for parent service should be changed")
}
