package common

import (
	"github.com/Aptomi/aptomi/pkg/api"
	"github.com/Aptomi/aptomi/pkg/config"
	"github.com/Aptomi/aptomi/pkg/engine/apply/action/component"
	"github.com/Aptomi/aptomi/pkg/engine/apply/action/global"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFormat_Text(t *testing.T) {
	cfg := config.Client{Output: Text}
	result := &api.PolicyUpdateResult{
		PolicyGeneration: 42,
		Actions: []string{
			component.CreateActionObject.Kind,
			component.UpdateActionObject.Kind,
			component.DeleteActionObject.Kind,
			component.DetachDependencyActionObject.Kind,
			component.AttachDependencyActionObject.Kind,
			component.EndpointsActionObject.Kind,
			global.PostProcessActionObject.Kind,
		},
	}
	data, err := Format(cfg, true, result)

	assert.Nil(t, err, "Format should work without error")

	assert.Equal(t, `Policy Generation	Expected Actions
42               	action-component-create
			action-component-update
			action-component-delete`, string(data), "Format should return expected table")
}
