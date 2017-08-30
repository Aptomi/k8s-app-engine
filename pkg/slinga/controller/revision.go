package controller

import (
	"github.com/Aptomi/aptomi/pkg/slinga/language"
	"github.com/Aptomi/aptomi/pkg/slinga/object"
	"github.com/Aptomi/aptomi/pkg/slinga/object/store"
)

type RevisionController interface {
	GetRevision(object.Generation) (*language.PolicyNamespace, error)
	NewRevision([]object.Base) error
}

func NewRevisionController(store store.ObjectStore) RevisionController {
	return &RevisionControllerImpl{store}
}

type RevisionControllerImpl struct {
	store store.ObjectStore
}

func (c *RevisionControllerImpl) GetRevision(gen object.Generation) (*language.PolicyNamespace, error) {
	return nil, nil
}

func (c *RevisionControllerImpl) NewRevision(update []object.Base) error {
	return nil
}

/*

func (reg *Registry) LoadPolicy(gen Generation) (*PolicyNamespace, error) {
	policyObj, err := reg.store.GetNewestOne("system", PolicyNamespaceDataObject.Kind, "main")
	if err != nil {
		return nil, err
	}
	policyData, ok := policyObj.(*PolicyNamespaceData)
	if !ok {
		return nil, fmt.Errorf("Can't cast object from store to PolicyData: %v", policyObj)
	}

	policy := NewPolicyNamespace()

	keys := make([]Key, 0, len(policyData.Objects))
	for _, key := range policyData.Objects {
		keys = append(keys, key)
	}

	objects, err := reg.store.GetManyByKeys(keys)
	if err != nil {
		return nil, fmt.Errorf("Can't load objects for policy data %s: %s", policyData.GetKey(), err)
	}

	for _, obj := range objects {
		fmt.Println("Loaded object")
		fmt.Println(obj)
		policy.AddObject(obj)
	}

	return policy, nil
}

*/
