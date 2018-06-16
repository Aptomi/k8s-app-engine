package engine

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/runtime"
)

var DesiredStateObject = &runtime.Info{
	Kind:        "desired-state",
	Storable:    true,
	Versioned:   false,
	Constructor: func() runtime.Object { return &DesiredState{} },
}

type DesiredState struct {
	runtime.TypeKind `yaml:",inline"`

	RevisionGen runtime.Generation
	Resolution  resolve.PolicyResolution
}

func NewDesiredState(revision *Revision, resolution *resolve.PolicyResolution) *DesiredState {
	return &DesiredState{
		TypeKind:    DesiredStateObject.GetTypeKind(),
		RevisionGen: revision.GetGeneration(),
		Resolution:  *resolution,
	}
}

func (ds *DesiredState) GetName() string {
	return GetDesiredStateName(ds.RevisionGen)
}

func (ds *DesiredState) GetNamespace() string {
	return runtime.SystemNS
}

func GetDesiredStateName(revisionGen runtime.Generation) string {
	return fmt.Sprintf("revision-%s-desired-state", revisionGen)
}
