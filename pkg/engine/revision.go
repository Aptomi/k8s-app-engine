package engine

import (
	"github.com/Aptomi/aptomi/pkg/runtime"
)

// RevisionName is the name of the only revision that exists in DB (but with many generations)
const RevisionName = "revision"

// RevisionObject is Info for Revision
var RevisionObject = &runtime.Info{
	Kind:        "revision",
	Storable:    true,
	Versioned:   true,
	Constructor: func() runtime.Object { return &Revision{} },
}

// RevisionKey is the default key for the Revision object (there is only one Revision exists but with multiple generations)
var RevisionKey = runtime.KeyFromParts(runtime.SystemNS, RevisionObject.Kind, runtime.EmptyName)

const (
	// RevisionStatusInProgress represents Revision status with apply in progress
	RevisionStatusInProgress = "inprogress"
	// RevisionStatusSuccess represents Revision status with apply successfully finished
	RevisionStatusSuccess = "success"
	// RevisionStatusError represents Revision status with apply finished with error
	RevisionStatusError = "error"
)

// Revision is a "milestone" in applying
type Revision struct {
	runtime.TypeKind `yaml:",inline"`
	Metadata         runtime.GenerationMetadata

	// Policy represents generation of the corresponding policy
	Policy runtime.Generation

	Status   string
	Progress RevisionProgress
}

// RevisionProgress represents revision applying progress
type RevisionProgress struct {
	Current int
	Total   int
}

// GetName returns Revision name
func (revision *Revision) GetName() string {
	return runtime.EmptyName
}

// GetNamespace returns Revision namespace
func (revision *Revision) GetNamespace() string {
	return runtime.SystemNS
}

// GetGeneration returns Revision generation
func (revision *Revision) GetGeneration() runtime.Generation {
	return revision.Metadata.Generation
}

// SetGeneration returns Revision generation
func (revision *Revision) SetGeneration(gen runtime.Generation) {
	revision.Metadata.Generation = gen
}
