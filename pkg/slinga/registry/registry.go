package registry

import (
	"fmt"
	. "github.com/Aptomi/aptomi/pkg/slinga/language"
	. "github.com/Aptomi/aptomi/pkg/slinga/object"
	. "github.com/Aptomi/aptomi/pkg/slinga/object/codec"
	. "github.com/Aptomi/aptomi/pkg/slinga/object/store"
)

type Registry struct {
	// TODO: looks like codec field is not used?
	codec   MarshalUnmarshaler
	store   ObjectStore
	catalog *ObjectCatalog
}

func (reg *Registry) AddKind(infos ...*ObjectInfo) {
	reg.catalog.Add(infos...)
}

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

var PolicyObjects = []*ObjectInfo{ServiceObject, ContextObject, ClusterObject, RuleObject, DependencyObject}

func NewDefaultRegistry() *Registry {
	reg := &Registry{
		catalog: NewObjectCatalog(),
	}
	reg.AddKind(PolicyObjects...)

	//reg.store = &file.FileStore{}
	//reg.store.SetObjectCatalog()
	//reg.store.Open(path)

	return reg
}
