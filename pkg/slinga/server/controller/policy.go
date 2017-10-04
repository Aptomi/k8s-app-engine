package controller

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/slinga/lang"
	"github.com/Aptomi/aptomi/pkg/slinga/object"
	"github.com/Aptomi/aptomi/pkg/slinga/object/store"
	"sync"
)

// PolicyName is an object name under which aptomi policy will be stored in the object store
const PolicyName = "policy"

// PolicyController defines methods for getting policy objects from the underlying data store
type PolicyController interface {
	GetPolicy(object.Generation) (*lang.Policy, error)
	GetPolicyData(object.Generation) (*PolicyData, error)
	GetPolicyFromData(policyData *PolicyData) (*lang.Policy, error)
	UpdatePolicy([]object.Base) (bool, *PolicyData, error)
}

// NewPolicyController creates a new PolicyController
func NewPolicyController(store store.ObjectStore) PolicyController {
	return &PolicyControllerImpl{sync.Mutex{}, store}
}

// PolicyDataObject is an informational data structure with Kind and Constructor for PolicyData
var PolicyDataObject = &object.Info{
	Kind:        "policy",
	Versioned:   true,
	Constructor: func() object.Base { return &PolicyData{} },
}

// PolicyData is a struct which represents policy in the data store. Containing references to a generation for each object included into the policy
type PolicyData struct {
	lang.Metadata

	Objects map[string]map[string]map[string]object.Generation // ns -> kind -> name -> generation
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

// PolicyControllerImpl implements PolicyController to retrieve objects from the underlying data store
type PolicyControllerImpl struct {
	update sync.Mutex
	store  store.ObjectStore
}

// GetPolicyData retrieves PolicyData given its generation
func (ctl *PolicyControllerImpl) GetPolicyData(gen object.Generation) (*PolicyData, error) {
	dataObj, err := ctl.store.GetByName(object.SystemNS, PolicyDataObject.Kind, PolicyName, gen)
	if err != nil {
		return nil, err
	}
	if dataObj == nil {
		return nil, nil
	}
	data, ok := dataObj.(*PolicyData)
	if !ok {
		return nil, fmt.Errorf("Unexpected type while getting PolicyData from DB")
	}
	return data, nil
}

// GetPolicyFromData converts PolicyData with store objects into a Policy for the engine
func (ctl *PolicyControllerImpl) GetPolicyFromData(policyData *PolicyData) (*lang.Policy, error) {
	policy := lang.NewPolicy()

	// in case of first version of policy, we just need to have empty policy
	if policyData != nil && policyData.Objects != nil {
		for ns, kindNameGen := range policyData.Objects {
			for kind, nameGen := range kindNameGen {
				for name, gen := range nameGen {
					obj, err := ctl.store.GetByName(ns, kind, name, gen)
					if err != nil {
						return nil, err
					}
					policy.AddObject(obj)
				}
			}
		}
	}
	return policy, nil
}

// GetPolicy retrieves PolicyData based on its generation and then converts it to Policy
func (ctl *PolicyControllerImpl) GetPolicy(policyGen object.Generation) (*lang.Policy, error) {
	// todo should we use RWMutex for get/update policy?
	policyData, err := ctl.GetPolicyData(policyGen)
	if err != nil {
		return nil, err
	}
	return ctl.GetPolicyFromData(policyData)
}

// UpdatePolicy updates a list of changed objects in the underlying data store
func (ctl *PolicyControllerImpl) UpdatePolicy(updatedObjects []object.Base) (bool, *PolicyData, error) {
	// we should process only a single policy update request at once
	ctl.update.Lock()
	defer ctl.update.Unlock()

	policyData, err := ctl.GetPolicyData(object.LastGen)
	if err != nil {
		return false, nil, err
	}

	changed := false

	// it could happen only for the fist time
	if policyData == nil {
		policyData = &PolicyData{
			Metadata: lang.Metadata{
				Namespace: object.SystemNS,
				Kind:      PolicyDataObject.Kind,
				Name:      "policy",
			},
			Objects: make(map[string]map[string]map[string]object.Generation),
		}
		changed = true
	}

	for _, updatedObj := range updatedObjects {
		var updated bool
		updated, err = ctl.store.Save(updatedObj)
		if err != nil {
			return false, nil, err
		}
		if updated {
			policyData.Add(updatedObj)
			changed = true
		}
	}

	if changed {
		_, err = ctl.store.Save(policyData)
		if err != nil {
			return false, nil, err
		}
	}

	// policy, err := ctl.getPolicyFromData(policyData)

	// todo: add policy changed / not changed response, if changed - show expected policy resolution? + attach resolution event log to the new version of policy

	// [3] Show user what changes will be triggered by his changes to the policy
	//   1. load previous desired state (from last revision)
	//   1. calculate new desired state (run resolver // resolver.ResolveAllDependencies())
	//   1. compare and return changes to the user [without saving to db]

	return changed, policyData, err
}
