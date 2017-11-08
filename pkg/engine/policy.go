package engine

import (
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/runtime"
)

// PolicyDataObject is an informational data structure with Kind and Constructor for PolicyData
var PolicyDataObject = &runtime.Info{
	Kind:        "policy",
	Storable:    true,
	Versioned:   true,
	Constructor: func() runtime.Object { return &PolicyData{} },
}

var PolicyDataKey = runtime.KeyFromParts(runtime.SystemNS, PolicyDataObject.Kind, runtime.EmptyName)

// PolicyData is a struct which contains references to a generation for each object included into the policy
type PolicyData struct {
	runtime.TypeKind `yaml:",inline"`
	Metadata         PolicyDataMetadata

	// Objects stores all policy objects in map: namespace -> kind -> name -> generation
	Objects map[string]map[string]map[string]runtime.Generation
}

type PolicyDataMetadata struct {
	Generation runtime.Generation
}

func (policyData *PolicyData) GetName() string {
	return runtime.EmptyName
}

func (policyData *PolicyData) GetNamespace() string {
	return runtime.SystemNS
}

func (policyData *PolicyData) GetGeneration() runtime.Generation {
	return policyData.Metadata.Generation
}

func (policyData *PolicyData) SetGeneration(gen runtime.Generation) {
	policyData.Metadata.Generation = gen
}

// Add adds an object to PolicyData
func (policyData *PolicyData) Add(obj lang.Base) {
	byNs, exist := policyData.Objects[obj.GetNamespace()]
	if !exist {
		byNs = make(map[string]map[string]runtime.Generation)
		policyData.Objects[obj.GetNamespace()] = byNs
	}
	byKind, exist := byNs[obj.GetKind()]
	if !exist {
		byKind = make(map[string]runtime.Generation)
		byNs[obj.GetKind()] = byKind
	}
	byKind[obj.GetName()] = obj.GetGeneration()
}
