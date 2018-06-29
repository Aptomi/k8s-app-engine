package registry

import (
	"fmt"
	"time"

	"github.com/Aptomi/aptomi/pkg/engine"
	"github.com/Aptomi/aptomi/pkg/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/runtime"
)

// GetPolicyData retrieves PolicyData given its generation
func (reg *defaultRegistry) GetPolicyData(gen runtime.Generation) (*engine.PolicyData, error) {
	dataObj, err := reg.store.GetGen(engine.PolicyDataKey, gen)
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
func (reg *defaultRegistry) getPolicyFromData(policyData *engine.PolicyData) (*lang.Policy, runtime.Generation, error) {
	if policyData == nil {
		return nil, runtime.LastGen, nil
	}

	policy := lang.NewPolicy()
	if policyData.Objects != nil {
		for ns, kindNameGen := range policyData.Objects {
			for kind, nameGen := range kindNameGen {
				for name, gen := range nameGen {
					obj, errStore := reg.store.GetGen(runtime.KeyFromParts(ns, kind, name), gen)
					if errStore != nil {
						return nil, 0, errStore
					}
					langObj, ok := obj.(lang.Base)
					if !ok {
						return nil, 0, fmt.Errorf("can't cast obj %s to lang.Base", runtime.KeyForStorable(obj))
					}
					errPolicy := policy.AddObject(langObj)
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
func (reg *defaultRegistry) GetPolicy(gen runtime.Generation) (*lang.Policy, runtime.Generation, error) {
	policyData, err := reg.GetPolicyData(gen)
	if err != nil {
		return nil, runtime.LastGen, err
	}
	return reg.getPolicyFromData(policyData)
}

// UpdatePolicy updates a list of changed objects in the underlying data registry
func (reg *defaultRegistry) UpdatePolicy(updatedObjects []lang.Base, performedBy string) (bool, *engine.PolicyData, error) {
	// we should process only a single policy update request at once
	reg.policyChangeLock.Lock()
	defer reg.policyChangeLock.Unlock()

	policyData, err := reg.GetPolicyData(runtime.LastGen)
	if err != nil {
		return false, nil, err
	}
	if policyData == nil {
		panic(fmt.Sprintf("cannot retrieve last policy from the registry, policyData is nil"))
	}

	changed := false
	for _, updatedObj := range updatedObjects {
		if updatedObj.IsDeleted() {
			return false, nil, fmt.Errorf("objects with deleted=true not supported while updating policy: %s", runtime.KeyForStorable(updatedObj))
		}

		var changedObj bool
		changedObj, err = reg.store.Save(updatedObj)
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
		_, err = reg.store.Save(policyData)
		if err != nil {
			return false, nil, err
		}
	}

	return changed, policyData, err
}

// InitPolicy initializes policy (on the first run of Aptomi)
func (reg *defaultRegistry) InitPolicy() error {
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
	_, err := reg.store.Save(initialPolicyData)
	if err != nil {
		return err
	}

	// create a new revision as well
	_, err = reg.NewRevision(initialPolicyData.GetGeneration(), resolve.NewPolicyResolution(), false)
	return err
}

// DeleteFromPolicy deletes provided objects from policy
func (reg *defaultRegistry) DeleteFromPolicy(deleted []lang.Base, performedBy string) (bool, *engine.PolicyData, error) {
	// we should process only a single policy update request at once
	reg.policyChangeLock.Lock()
	defer reg.policyChangeLock.Unlock()

	policyData, err := reg.GetPolicyData(runtime.LastGen)
	if err != nil {
		return false, nil, err
	}

	policyChanged := false
	for _, obj := range deleted {
		if policyData.Remove(obj) {
			policyChanged = true
		}

		if !obj.IsDeleted() {
			obj.SetDeleted(true)
			_, err = reg.store.Save(obj)
			if err != nil {
				return false, nil, fmt.Errorf("error while setting deleted=true for %s: %s", runtime.KeyForStorable(obj), err)
			}
		}
	}

	if policyChanged {
		policyData.Metadata.UpdatedAt = time.Now()
		policyData.Metadata.UpdatedBy = performedBy

		// save policy data
		_, err = reg.store.Save(policyData)
		if err != nil {
			return false, nil, err
		}
	}

	return policyChanged, policyData, nil
}
