package engine

import (
	"github.com/Aptomi/aptomi/pkg/engine/apply/action"
	"github.com/Aptomi/aptomi/pkg/event"
	"github.com/Aptomi/aptomi/pkg/runtime"
	"time"
)

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
	// RevisionStatusSuccess represents Revision status with apply finished
	RevisionStatusCompleted = "completed"
	// RevisionStatusError represents Revision status when a critical error happened (we should rarely see those)
	RevisionStatusError = "error"
)

// Revision is a "milestone" in applying policy changes
type Revision struct {
	runtime.TypeKind `yaml:",inline"`
	Metadata         runtime.GenerationMetadata

	// Policy represents generation of the corresponding policy
	Policy runtime.Generation

	Status    string
	Progress  *RevisionProgress
	AppliedAt time.Time

	Stats *action.ApplyResult

	ResolveLog []*event.APIEvent
	ApplyLog   []*event.APIEvent
}

// RevisionProgress represents revision applying progress
type RevisionProgress struct {
	Current int
	Total   int
}

// NewRevision creates a new revision
func NewRevision(gen runtime.Generation, policyGen runtime.Generation) *Revision {
	return &Revision{
		TypeKind: RevisionObject.GetTypeKind(),
		Metadata: runtime.GenerationMetadata{
			Generation: gen,
		},
		Policy:   policyGen,
		Status:   RevisionStatusInProgress,
		Progress: &RevisionProgress{},
		Stats:    &action.ApplyResult{},
	}
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
