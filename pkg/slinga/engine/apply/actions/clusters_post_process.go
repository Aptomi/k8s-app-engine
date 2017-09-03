package actions

import (
	"github.com/Aptomi/aptomi/pkg/slinga/object"
)

type ClustersPostProcess struct {
	object.Metadata
	*BaseAction
}

func NewClustersPostProcessAction() *ClustersPostProcess {
	return &ClustersPostProcess{
		Metadata:   object.Metadata{}, // TODO: initialize
		BaseAction: NewComponentBaseAction(),
	}
}

func (clusterPostProcess *ClustersPostProcess) Apply(context *ActionContext) error {
	return nil
}
