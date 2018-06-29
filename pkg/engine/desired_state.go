package engine

import (
	"fmt"

	"github.com/Aptomi/aptomi/pkg/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/runtime"
)

// DesiredStateObject is an informational data structure with Kind and Constructor for DesiredState
var DesiredStateObject = &runtime.TypeInfo{
	Kind:        "desired-state",
	Storable:    true,
	Versioned:   false,
	Constructor: func() runtime.Object { return &DesiredState{} },
}

// DesiredState represents snapshot of the state to be achieved by specific revision
type DesiredState struct {
	runtime.TypeKind `yaml:",inline"`

	RevisionGen runtime.Generation
	Resolution  resolve.PolicyResolution
}

// NewDesiredState creates new DesiredState instance from revision and policy resolution
func NewDesiredState(revision *Revision, resolution *resolve.PolicyResolution) *DesiredState {
	return &DesiredState{
		TypeKind:    DesiredStateObject.GetTypeKind(),
		RevisionGen: revision.GetGeneration(),
		Resolution:  *resolution,
	}
}

// GetName returns name of the DesiredState
func (ds *DesiredState) GetName() string {
	return GetDesiredStateName(ds.RevisionGen)
}

// GetNamespace returns namespace of the DesiredState
func (ds *DesiredState) GetNamespace() string {
	return runtime.SystemNS
}

// GetDesiredStateName returns name of the DesiredState for specific Revision generations
func GetDesiredStateName(revisionGen runtime.Generation) string {
	return fmt.Sprintf("revision-%s-desired-state", revisionGen)
}
