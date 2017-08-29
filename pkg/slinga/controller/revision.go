package controller

//import (
//	. "github.com/Aptomi/aptomi/pkg/slinga/object"
//)

type RevisionController interface {
	/*
		router.GET("/api/v1/revision/:rev/policy", h.handleGetPolicy)               // get full policy from specific revision
		router.GET("/api/v1/revision/:rev/policy/key/:key", h.handleGetPolicy)      // get by key from specific revision
		router.GET("/api/v1/revision/:rev/policy/namespace/:ns", h.handleGetPolicy) // get policy for namespace from specific revision

		router.POST("/api/v1/revision", h.handleNewRevision)
	*/

	//GetRevision(generation Generation)
}

type RevisionControllerImpl struct {
}

func NewRevisionController() RevisionController {
	return &RevisionControllerImpl{}
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
