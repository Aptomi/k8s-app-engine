package store

import (
	"github.com/Aptomi/aptomi/pkg/slinga/lang"
	"github.com/Aptomi/aptomi/pkg/slinga/object"
	"github.com/Aptomi/aptomi/pkg/slinga/object/store"
)

type ServerStore interface {
	// Object returns store.ObjectStore
	Object() store.ObjectStore

	PolicyStore
	RevisionStore
}

type PolicyStore interface {
	GetPolicy(object.Generation) (*lang.Policy, error)
	GetPolicyData(object.Generation) (*PolicyData, error)
	UpdatePolicy([]object.Base) (bool, *PolicyData, error)
}

type RevisionStore interface {
	GetRevision(object.Generation) (*lang.Policy, error)
	UpdateRevision() error
}

type ActualStateStore interface {
}

// PolicyName is an object name under which aptomi policy will be stored in the object store
const PolicyName = "policy"

// PolicyDataObject is an informational data structure with Kind and Constructor for PolicyData
var PolicyDataObject = &object.Info{
	Kind:        "policy",
	Versioned:   true,
	Constructor: func() object.Base { return &PolicyData{} },
}

// PolicyData is a struct which represents policy in the data store. Containing references to a generation for each object included into the policy
type PolicyData struct {
	lang.Metadata

	// Objects stores all policy objects in map: namespace -> kind -> name -> generation
	Objects map[string]map[string]map[string]object.Generation
}

// Add adds an object to PolicyData
func (p *PolicyData) Add(obj object.Base) {
	byNs, exist := p.Objects[obj.GetNamespace()]
	if !exist {
		byNs = make(map[string]map[string]object.Generation)
		p.Objects[obj.GetNamespace()] = byNs
	}
	byKind, exist := byNs[obj.GetKind()]
	if !exist {
		byKind = make(map[string]object.Generation)
		byNs[obj.GetKind()] = byKind
	}
	byKind[obj.GetName()] = obj.GetGeneration()
}
