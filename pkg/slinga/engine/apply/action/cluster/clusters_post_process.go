package cluster

import (
	"github.com/Aptomi/aptomi/pkg/slinga/engine/apply/action"
	"github.com/Aptomi/aptomi/pkg/slinga/object"
)

type ClustersPostProcessAction struct {
	object.Metadata
	*action.Base
}

func NewClustersPostProcessAction() *ClustersPostProcessAction {
	return &ClustersPostProcessAction{
		Metadata: object.Metadata{}, // TODO: initialize
		Base:     action.NewBase(),
	}
}

func (clusterPostProcess *ClustersPostProcessAction) Apply(context *action.Context) error {
	return nil
}
