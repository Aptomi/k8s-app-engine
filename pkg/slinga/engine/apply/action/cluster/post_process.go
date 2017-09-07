package cluster

import (
	"github.com/Aptomi/aptomi/pkg/slinga/engine/apply/action"
	"github.com/Aptomi/aptomi/pkg/slinga/object"
)

var PostProcessActionObject = &object.Info{
	Kind:        "action-clusters-post-process",
	Constructor: func() object.Base { return &PostProcessAction{} },
}

type PostProcessAction struct {
	*action.Metadata
}

func NewClustersPostProcessAction(revision object.Generation) *PostProcessAction {
	return &PostProcessAction{
		Metadata: action.NewMetadata(revision, PostProcessActionObject.Kind),
	}
}

func (a *PostProcessAction) GetName() string {
	return "Clusters post process"
}

func (a *PostProcessAction) Apply(context *action.Context) error {
	return nil
}
