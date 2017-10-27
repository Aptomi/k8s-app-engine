package global

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/engine/apply/action"
	"github.com/Aptomi/aptomi/pkg/object"
)

// PostProcessActionObject is an informational data structure with Kind and Constructor for the action
var PostProcessActionObject = &object.Info{
	Kind:        "action-post-process",
	Constructor: func() object.Base { return &PostProcessAction{} },
}

// PostProcessAction is a post-processing action which gets called once after all components have been
// processed by the engine apply
type PostProcessAction struct {
	*action.Metadata
}

// NewPostProcessAction creates new PostProcessAction
func NewPostProcessAction(revision object.Generation) *PostProcessAction {
	return &PostProcessAction{
		Metadata: action.NewMetadata(revision, PostProcessActionObject.Kind),
	}
}

// Apply runs all registered post-processing plugins
func (a *PostProcessAction) Apply(context *action.Context) error {
	for _, plugin := range context.Plugins.GetPostProcessingPlugins() {
		err := plugin.Process(context.DesiredPolicy, context.DesiredState, context.ExternalData, context.EventLog)
		if err != nil {
			context.EventLog.LogError(err)
			return fmt.Errorf("error while running post processing action: %s", err)
		}
	}

	return nil
}
