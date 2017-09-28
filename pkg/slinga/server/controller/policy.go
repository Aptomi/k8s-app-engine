package controller

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/slinga/language"
	"github.com/Aptomi/aptomi/pkg/slinga/object"
	"github.com/Aptomi/aptomi/pkg/slinga/object/store"
	"sync"
)

const PolicyName = "policy"

type PolicyController interface {
	GetPolicy(object.Generation) (*language.Policy, error)
	GetPolicyData(object.Generation) (*PolicyData, error)
	GetPolicyFromData(policyData *PolicyData) (*language.Policy, error)
	UpdatePolicy([]object.Base) (bool, *PolicyData, error)
}

func NewPolicyController(store store.ObjectStore) PolicyController {
	return &PolicyControllerImpl{sync.Mutex{}, store}
}

var PolicyDataObject = &object.Info{
	Kind:        "policy",
	Versioned:   true,
	Constructor: func() object.Base { return &PolicyData{} },
}

type PolicyData struct {
	language.Metadata

	Objects map[string]map[string]object.Generation // kind -> name -> generation
}

func (p *PolicyData) Add(obj object.Base) {
	byKind, exist := p.Objects[obj.GetKind()]
	if !exist {
		byKind = make(map[string]object.Generation)
		p.Objects[obj.GetKind()] = byKind
	}
	byKind[obj.GetName()] = obj.GetGeneration()
}

type PolicyControllerImpl struct {
	update sync.Mutex
	store  store.ObjectStore
}

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

func (ctl *PolicyControllerImpl) GetPolicyFromData(policyData *PolicyData) (*language.Policy, error) {
	policy := language.NewPolicy()

	// in case of first version of policy, we just need to have empty policy
	if policyData != nil && policyData.Objects != nil {
		for kind, nameGen := range policyData.Objects {
			for name, gen := range nameGen {
				obj, err := ctl.store.GetByName(object.SystemNS, kind, name, gen)
				if err != nil {
					return nil, err
				}
				policy.AddObject(obj)
			}
		}
	}
	return policy, nil
}

func (ctl *PolicyControllerImpl) GetPolicy(policyGen object.Generation) (*language.Policy, error) {
	policyData, err := ctl.GetPolicyData(policyGen)
	if err != nil {
		return nil, err
	}
	return ctl.GetPolicyFromData(policyData)
}

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
			Metadata: language.Metadata{
				Namespace: object.SystemNS,
				Kind:      PolicyDataObject.Kind,
				Name:      "policy",
			},
			Objects: make(map[string]map[string]object.Generation),
		}
		changed = true
	}

	for _, updatedObj := range updatedObjects {
		updated, err := ctl.store.Save(updatedObj)
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

	return changed, policyData, err
}
