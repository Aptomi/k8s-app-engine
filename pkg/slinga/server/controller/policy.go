package controller

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/slinga/lang"
	"github.com/Aptomi/aptomi/pkg/slinga/object"
	"github.com/Aptomi/aptomi/pkg/slinga/object/store"
	"sync"
)

const PolicyName = "policy"

type PolicyController interface {
	GetPolicy(object.Generation) (*lang.Policy, error)
	GetPolicyData(object.Generation) (*PolicyData, error)
	GetPolicyFromData(policyData *PolicyData) (*lang.Policy, error)
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
	lang.Metadata

	Objects map[string]map[string]map[string]object.Generation // ns -> kind -> name -> generation
}

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

func (ctl *PolicyControllerImpl) GetPolicy(policyGen object.Generation) (*lang.Policy, error) {
	// todo should we use RWMutex for get/update policy?
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
