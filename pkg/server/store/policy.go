package store

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/object"
)

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

// GetPolicyData retrieves PolicyData given its generation
func (s *defaultStore) GetPolicyData(gen object.Generation) (*PolicyData, error) {
	dataObj, err := s.store.GetByName(object.SystemNS, PolicyDataObject.Kind, PolicyName, gen)
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

func (s *defaultStore) getPolicyFromData(policyData *PolicyData) (*lang.Policy, object.Generation, error) {
	policy := lang.NewPolicy()

	// in case of first version of policy, we just need to have empty policy
	if policyData != nil {
		if policyData.Objects != nil {
			for ns, kindNameGen := range policyData.Objects {
				for kind, nameGen := range kindNameGen {
					for name, gen := range nameGen {
						obj, err := s.store.GetByName(ns, kind, name, gen)
						if err != nil {
							return nil, 0, err
						}
						policy.AddObject(obj)
					}
				}
			}
		}
		return policy, policyData.Generation, nil
	} else {
		return policy, 0, nil
	}
}

// GetPolicy retrieves PolicyData based on its generation and then converts it to Policy
func (s *defaultStore) GetPolicy(policyGen object.Generation) (*lang.Policy, object.Generation, error) {
	// todo should we use RWMutex for get/update policy?
	policyData, err := s.GetPolicyData(policyGen)
	if err != nil {
		return nil, 0, err
	}
	return s.getPolicyFromData(policyData)
}

// UpdatePolicy updates a list of changed objects in the underlying data store
func (s *defaultStore) UpdatePolicy(updatedObjects []object.Base) (bool, *PolicyData, error) {
	// we should process only a single policy update request at once
	s.policyUpdate.Lock()
	defer s.policyUpdate.Unlock()

	policyData, err := s.GetPolicyData(object.LastGen)
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
				Name:      PolicyName,
			},
			Objects: make(map[string]map[string]map[string]object.Generation),
		}
		changed = true
	}

	for _, updatedObj := range updatedObjects {
		var updated bool
		updated, err = s.store.Save(updatedObj)
		if err != nil {
			return false, nil, err
		}
		if updated {
			policyData.Add(updatedObj)
			changed = true
		}
	}

	if changed {
		_, err = s.store.Save(policyData)
		if err != nil {
			return false, nil, err
		}
	}

	return changed, policyData, err
}
