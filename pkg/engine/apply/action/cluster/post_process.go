package cluster

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/engine/apply/action"
	"github.com/Aptomi/aptomi/pkg/object"
)

// PostProcessActionObject is an informational data structure with Kind and Constructor for the action
var PostProcessActionObject = &object.Info{
	Kind:        "action-clusters-post-process",
	Constructor: func() object.Base { return &PostProcessAction{} },
}

// PostProcessAction is a global post-processing action which gets called once after all components have been processed by the engine
type PostProcessAction struct {
	*action.Metadata
}

// NewClustersPostProcessAction creates new PostProcessAction
func NewClustersPostProcessAction(revision object.Generation) *PostProcessAction {
	return &PostProcessAction{
		Metadata: action.NewMetadata(revision, PostProcessActionObject.Kind),
	}
}

// Apply applies the action
func (a *PostProcessAction) Apply(context *action.Context) error {
	for _, plugin := range context.Plugins.GetClustersPostProcessingPlugins() {
		err := plugin.Process(context.DesiredPolicy, context.DesiredState, context.ExternalData, context.EventLog)
		if err != nil {
			context.EventLog.LogError(err)
			return fmt.Errorf("Error while post processing clusters: %s", err)
		}
	}

	return nil
}
