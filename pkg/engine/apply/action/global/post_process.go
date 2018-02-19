package global

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/engine/apply/action"
	"github.com/Aptomi/aptomi/pkg/runtime"
)

// PostProcessActionObject is an informational data structure with Kind and Constructor for the action
var PostProcessActionObject = &runtime.Info{
	Kind:        "action-post-process",
	Constructor: func() runtime.Object { return &PostProcessAction{} },
}

// PostProcessAction is a post-processing action which gets called once after all components have been
// processed by the engine apply
type PostProcessAction struct {
	runtime.TypeKind `yaml:",inline"`
	*action.Metadata
}

// NewPostProcessAction creates new PostProcessAction
func NewPostProcessAction() *PostProcessAction {
	return &PostProcessAction{
		TypeKind: PostProcessActionObject.GetTypeKind(),
		Metadata: action.NewMetadata(PostProcessActionObject.Kind),
	}
}

// Apply runs all registered post-processing plugins
func (a *PostProcessAction) Apply(context *action.Context) error {
	for _, plugin := range context.Plugins.PostProcess() {
		err := plugin.Process(context.DesiredPolicy, context.DesiredState, context.ExternalData, context.EventLog)
		if err != nil {
			return fmt.Errorf("error while running post processing action: %s", err)
		}
	}

	return nil
}
