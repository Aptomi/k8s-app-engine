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

var RevisionKey = runtime.KeyFromParts(runtime.SystemNS, RevisionObject.Kind, runtime.EmptyName)

// Revision is a "milestone" in applying
type Revision struct {
	runtime.TypeKind `yaml:",inline"`
	Metadata         runtime.GenerationMetadata

	// Policy represents generation of the corresponding policy
	Policy runtime.Generation

	Progress RevisionProgress
}

type RevisionProgress struct {
	Stage    string
	Current  int
	Total    int
	Finished bool
}

func (revision *Revision) GetName() string {
	return runtime.EmptyName
}

func (revision *Revision) GetNamespace() string {
	return runtime.SystemNS
}

func (revision *Revision) GetGeneration() runtime.Generation {
	return revision.Metadata.Generation
}

func (revision *Revision) SetGeneration(gen runtime.Generation) {
	revision.Metadata.Generation = gen
}
