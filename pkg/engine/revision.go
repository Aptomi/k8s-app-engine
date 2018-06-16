package engine

import (
	"time"

	"github.com/Aptomi/aptomi/pkg/engine/apply/action"
	"github.com/Aptomi/aptomi/pkg/event"
	"github.com/Aptomi/aptomi/pkg/runtime"
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
	// RevisionStatusWaiting represents Revision status when it has been created, but apply haven't started yet
	RevisionStatusWaiting = "waiting"
	// RevisionStatusInProgress represents Revision status with apply in progress
	RevisionStatusInProgress = "inprogress"
	// RevisionStatusCompleted represents Revision status with apply finished
	RevisionStatusCompleted = "completed"
	// RevisionStatusError represents Revision status when a critical error happened (we should rarely see those)
	RevisionStatusError = "error"
)

// Revision is a "milestone" in applying policy changes
type Revision struct {
	runtime.TypeKind `yaml:",inline"`
	Metadata         runtime.GenerationMetadata

	// Policy to which this revision is attached to
	PolicyGen runtime.Generation

	Status         string
	CreatedAt      time.Time
	RecalculateAll bool

	Result    *action.ApplyResult
	AppliedAt time.Time

	// TODO: do not store apply log in revision
	ApplyLog []*event.APIEvent
}

// NewRevision creates a new revision
func NewRevision(gen runtime.Generation, policyGen runtime.Generation, recalculateAll bool) *Revision {
	return &Revision{
		TypeKind: RevisionObject.GetTypeKind(),
		Metadata: runtime.GenerationMetadata{
			Generation: gen,
		},
		PolicyGen:      policyGen,
		Status:         RevisionStatusWaiting,
		CreatedAt:      time.Now(),
		RecalculateAll: recalculateAll,
		Result:         &action.ApplyResult{},
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
