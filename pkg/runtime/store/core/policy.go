package core

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/engine"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/runtime"
	"time"
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

// getPolicyFromData() returns Policy converted from PolicyData.
// if PolicyData is nil, it will return nil
func (ds *defaultStore) getPolicyFromData(policyData *engine.PolicyData) (*lang.Policy, runtime.Generation, error) {
	if policyData == nil {
		return nil, runtime.LastGen, nil
	}

	policy := lang.NewPolicy()
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

// GetPolicy retrieves PolicyData based on its generation and then converts it to Policy
// if there is no policy yet (Aptomi not initialized), it will return nil
func (ds *defaultStore) GetPolicy(gen runtime.Generation) (*lang.Policy, runtime.Generation, error) {
	// todo should we use RWMutex for get/update policy?
	policyData, err := ds.GetPolicyData(gen)
	if err != nil {
		return nil, runtime.LastGen, err
	}
	return ds.getPolicyFromData(policyData)
}

// UpdatePolicy updates a list of changed objects in the underlying data store
func (ds *defaultStore) UpdatePolicy(updatedObjects []lang.Base, performedBy string) (bool, *engine.PolicyData, error) {
	// we should process only a single policy update request at once
	ds.policyChangeLock.Lock()
	defer ds.policyChangeLock.Unlock()

	policyData, err := ds.GetPolicyData(runtime.LastGen)
	if err != nil {
		return false, nil, err
	}
	if policyData == nil {
		panic(fmt.Sprintf("Cannot retrieve last policy from the store, policyData is nil"))
	}

	changed := false
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
		// update metadata before saving policy data (to capture who and when edited the policy)
		policyData.Metadata.UpdatedAt = time.Now()
		policyData.Metadata.UpdatedBy = performedBy

		// save policy data
		_, err = ds.store.Save(policyData)
		if err != nil {
			return false, nil, err
		}
	}

	return changed, policyData, err
}

// init policy initializes policy (on the first run of Aptomi)
func (ds *defaultStore) InitPolicy() error {
	// create and save
	initialPolicyData := &engine.PolicyData{
		TypeKind: engine.PolicyDataObject.GetTypeKind(),
		Metadata: engine.PolicyDataMetadata{
			Generation: runtime.FirstGen,
			UpdatedAt:  time.Now(),
			UpdatedBy:  "aptomi",
		},
		Objects: make(map[string]map[string]map[string]runtime.Generation),
	}
	// save policy data
	_, err := ds.store.Save(initialPolicyData)
	return err
}

// DeleteFromPolicy deletes provided objects from policy
func (ds *defaultStore) DeleteFromPolicy(deleted []lang.Base, performedBy string) (bool, *engine.PolicyData, error) {
	// we should process only a single policy update request at once
	ds.policyChangeLock.Lock()
	defer ds.policyChangeLock.Unlock()

	policyData, err := ds.GetPolicyData(runtime.LastGen)
	if err != nil {
		return false, nil, err
	}

	changed := false
	for _, obj := range deleted {
		if policyData.Remove(obj) {
			changed = true
		}
	}

	if changed {
		policyData.Metadata.UpdatedAt = time.Now()
		policyData.Metadata.UpdatedBy = performedBy

		// save policy data
		_, err = ds.store.Save(policyData)
		if err != nil {
			return false, nil, err
		}
	}

	return changed, policyData, nil
}
