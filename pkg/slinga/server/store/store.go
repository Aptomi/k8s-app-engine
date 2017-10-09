package store

import (
	"github.com/Aptomi/aptomi/pkg/slinga/engine/actual"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/progress"
	"github.com/Aptomi/aptomi/pkg/slinga/lang"
	"github.com/Aptomi/aptomi/pkg/slinga/object"
	"github.com/Aptomi/aptomi/pkg/slinga/object/store"
)

type ServerStore interface {
	// Object returns store.ObjectStore
	Object() store.ObjectStore

	PolicyStore
	RevisionStore

	Progress(store store.ObjectStore, revision *RevisionData) progress.Indicator

	ActualStateUpdater() actual.StateUpdater
}

type PolicyStore interface {
	GetPolicy(object.Generation) (*lang.Policy, object.Generation, error)
	GetPolicyData(object.Generation) (*PolicyData, error)
	UpdatePolicy([]object.Base) (bool, *PolicyData, error)
}

type RevisionStore interface {
	GetRevision(object.Generation) (*RevisionData, error)
	NextRevision(policyGen object.Generation) (*RevisionData, error)
	SaveRevision(revision *RevisionData) error
}

// PolicyName is an object name under which aptomi policy will be stored in the object store
const PolicyName = "policy"

// PolicyDataObject is an informational data structure with Kind and Constructor for PolicyData
var PolicyDataObject = &object.Info{
	Kind:        "policy",
	Versioned:   true,
	Constructor: func() object.Base { return &PolicyData{} },
}
