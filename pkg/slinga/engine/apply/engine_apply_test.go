package apply

import (
	"github.com/Aptomi/aptomi/pkg/slinga/engine/diff"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/plugin"
	"github.com/Aptomi/aptomi/pkg/slinga/language"
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	ResSuccess = iota
	ResError   = iota
)

func TestApplyCreateSuccess(t *testing.T) {
	userLoader := getUserLoader()

	// resolve empty policy
	actualPolicy := language.NewPolicyNamespace()
	actualState := resolvePolicy(t, actualPolicy, userLoader)

	// resolve full policy
	desiredPolicy := getPolicy()
	desiredState := resolvePolicy(t, desiredPolicy, userLoader)

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
		userLoader,
		actions,
		plugins,
	)

	// check actual state
	assert.Equal(t, 0, len(actualState.Resolved.ComponentInstanceMap), "Actual state should be empty")

	// check that policy apply finished with expected results
	applyAndCheck(t, apply, ResSuccess, 0, "")

	// check that actual state got updated
	assert.Equal(t, 16, len(actualState.Resolved.ComponentInstanceMap), "Actual state should be empty")
}

func TestApplyCreateFailure(t *testing.T) {
	userLoader := getUserLoader()

	// resolve empty policy
	actualPolicy := language.NewPolicyNamespace()
	actualState := resolvePolicy(t, actualPolicy, userLoader)

	// resolve full policy
	desiredPolicy := getPolicy()
	desiredState := resolvePolicy(t, desiredPolicy, userLoader)

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
		userLoader,
		actions,
		plugins,
	)

	// check actual state
	assert.Equal(t, 0, len(actualState.Resolved.ComponentInstanceMap), "Actual state should be empty")

	// check that policy apply finished with expected results
	applyAndCheck(t, apply, ResError, 4, "Apply failed for component")

	// check that actual state got updated
	assert.Equal(t, 12, len(actualState.Resolved.ComponentInstanceMap), "Actual state should be empty")
}
