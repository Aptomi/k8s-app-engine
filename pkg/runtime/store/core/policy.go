package core

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/engine"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/runtime"
)

// GetPolicyData retrieves PolicyData given its generation
func (ds *defaultStore) GetPolicyData(gen runtime.Generation) (*engine.PolicyData, error) {
	dataObj, err := ds.store.GetGen(engine.PolicyDataKey, gen)
	if err != nil {
		return nil, err
	}
	if dataObj == nil {
		return nil, nil
	}
	data, ok := dataObj.(*engine.PolicyData)
	if !ok {
		return nil, fmt.Errorf("unexpected type while getting PolicyData from DB")
	}
	return data, nil
}

func (ds *defaultStore) getPolicyFromData(policyData *engine.PolicyData) (*lang.Policy, runtime.Generation, error) {
	policy := lang.NewPolicy()

	// in case of first version of policy, we just need to have empty policy
	if policyData != nil {
		if policyData.Objects != nil {
			for ns, kindNameGen := range policyData.Objects {
				for kind, nameGen := range kindNameGen {
					for name, gen := range nameGen {
						obj, errStore := ds.store.GetGen(runtime.KeyFromParts(ns, kind, name), gen)
						if errStore != nil {
							return nil, 0, errStore
						}
						errPolicy := policy.AddObject(obj)
						if errPolicy != nil {
							return nil, runtime.LastGen, errPolicy
						}
					}
				}
			}
		}
		return policy, policyData.GetGeneration(), nil
	}

	return policy, 0, nil
}

// GetPolicy retrieves PolicyData based on its generation and then converts it to Policy
func (ds *defaultStore) GetPolicy(gen runtime.Generation) (*lang.Policy, runtime.Generation, error) {
	// todo should we use RWMutex for get/update policy?
	policyData, err := ds.GetPolicyData(gen)
	if err != nil {
		return nil, runtime.LastGen, err
	}
	return ds.getPolicyFromData(policyData)
}

// UpdatePolicy updates a list of changed objects in the underlying data store
func (ds *defaultStore) UpdatePolicy(updatedObjects []lang.Base, deleted []runtime.Key) (bool, *engine.PolicyData, error) {
	// todo(slukjanov): handle deleted

	// we should process only a single policy update request at once
	ds.policyChangeLock.Lock()
	defer ds.policyChangeLock.Unlock()

	policyData, err := ds.GetPolicyData(runtime.LastGen)
	if err != nil {
		return false, nil, err
	}

	changed := false

	// it could happen only for the fist time
	if policyData == nil {
		policyData = &engine.PolicyData{
			TypeKind: engine.PolicyDataObject.GetTypeKind(),
			Metadata: engine.PolicyDataMetadata{Generation: runtime.FirstGen},
			Objects:  make(map[string]map[string]map[string]runtime.Generation),
		}
		changed = true
	}

	for _, updatedObj := range updatedObjects {
		var changedObj bool
		changedObj, err = ds.store.Save(updatedObj)
		if err != nil {
			return false, nil, err
		}
		if changedObj {
			policyData.Add(updatedObj)
			changed = true
		}
	}

	if changed {
		_, err = ds.store.Save(policyData)
		if err != nil {
			return false, nil, err
		}
	}

	return changed, policyData, err
}
