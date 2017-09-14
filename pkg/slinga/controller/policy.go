package controller

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/slinga/language"
	"github.com/Aptomi/aptomi/pkg/slinga/object"
	"github.com/Aptomi/aptomi/pkg/slinga/object/store"
)

const PolicyKey = "policy"

type PolicyController interface {
	GetPolicy(object.Generation) (*language.Policy, error)
	UpdatePolicy([]object.Base) (*language.Policy, error)
}

func NewPolicyController(store store.ObjectStore) PolicyController {
	return &PolicyControllerImpl{store}
}

var PolicyDataObject = &object.Info{
	Kind:        "policy",
	Constructor: func() object.Base { return &PolicyData{} },
}

type PolicyData struct {
	language.Metadata

	Objects map[string]map[string]object.Generation // kind -> key -> generation
}

func (p *PolicyData) Add(obj object.Base) {
	byKind, exist := p.Objects[obj.GetKind()]
	if !exist {
		byKind = make(map[string]object.Generation)
		p.Objects[obj.GetKind()] = byKind
	}
	byKind[obj.GetKey()] = obj.GetGeneration()
}

type PolicyControllerImpl struct {
	store store.ObjectStore
}

func (c *PolicyControllerImpl) getPolicyData(gen object.Generation) (*PolicyData, error) {
	dataObj, err := c.store.GetByKey(object.SystemNS, PolicyDataObject.Kind, PolicyKey, gen)
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

func (c *PolicyControllerImpl) getPolicyFromData(policyData *PolicyData) (*language.Policy, error) {
	policy := language.NewPolicy()

	// in case of first version of policy, we just need to have empty policy
	if policyData != nil && policyData.Objects != nil {
		for kind, keyAndGen := range policyData.Objects {
			for key, gen := range keyAndGen {
				obj, err := c.store.GetByKey(object.DefaultNS, kind, key, gen)
				if err != nil {
					return nil, err
				}
				policy.AddObject(obj)
			}
		}
	}
	return policy, nil
}

func (c *PolicyControllerImpl) GetPolicy(policyGen object.Generation) (*language.Policy, error) {
	policyData, err := c.getPolicyData(policyGen)
	if err != nil {
		return nil, err
	}
	return c.getPolicyFromData(policyData)
}

func (c *PolicyControllerImpl) UpdatePolicy(updatedObjects []object.Base) (*language.Policy, error) {
	policyData, err := c.getPolicyData(object.LastGen)
	if err != nil {
		return nil, err
	}

	changed := false
	for _, updatedObj := range updatedObjects {
		updated, err := c.store.Save(updatedObj)
		if err != nil {
			return nil, err
		}
		if updated {
			policyData.Add(updatedObj)
			changed = true
		}
	}

	if changed {
		updated, err := c.store.Save(policyData)
		if err != nil {
			return nil, err
		}
		if !updated {
			return nil, fmt.Errorf("Policy was changed but save returned 'not updated'")
		}
	}

	return c.getPolicyFromData(policyData)
}
