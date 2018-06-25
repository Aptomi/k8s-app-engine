package lang

import (
	"testing"

	"github.com/Aptomi/aptomi/pkg/runtime"
	"github.com/stretchr/testify/assert"
)

func TestPolicyViewCommonObjects(t *testing.T) {
	// make policy with objects
	_, policyWithObjects := makePolicyWithObjects()

	// get the list of objects
	objList := []Base{}
	for _, obj := range PolicyObjects {
		objList = append(objList, policyWithObjects.GetObjectsByKind(obj.Kind)...)
	}
	assert.NotEmpty(t, objList, "Object list should not be empty")

	// users which will be used for viewing policy
	users := []*User{
		{Name: "1", Labels: map[string]string{"is_domain_admin": "true"}},
		{Name: "2", Labels: map[string]string{"is_namespace_admin": "true"}},
		{Name: "3", Labels: map[string]string{"is_consumer": "true"}},
	}

	// check AddObject()
	errCnt := []int{0, 0, 0}
	for i := 0; i < len(users); i++ {
		// make empty policy with custom ACL and add objects to it
		policyView := makeEmptyPolicyWithACL().View(users[i])
		for _, obj := range objList {
			if policyView.AddObject(obj) != nil {
				errCnt[i]++
			}
		}
	}
	assert.Equal(t, []int{0, 10, 40}, errCnt, "PolicyView.AddObject() should work correctly")

	// construct policy with ACL and add all objects into it
	policy := makeEmptyPolicyWithACL()
	for _, obj := range objList {
		err := policy.AddObject(obj)
		assert.NoError(t, err, "Policy.AddObject() should work correctly")
	}

	// check ViewObject() and ManageObject() on policy with objects
	errCntView := []int{0, 0, 0}
	errCntManage := []int{0, 0, 0}
	for i := 0; i < len(users); i++ {
		policyView := policy.View(users[i])
		for _, obj := range objList {
			if err := policyView.ViewObject(obj); err != nil {
				errCntView[i]++
			}
			if err := policyView.ManageObject(obj); err != nil {
				errCntManage[i]++
			}
		}
	}
	assert.Equal(t, []int{0, 0, 0}, errCntView, "PolicyView.ViewObject() should work correctly")
	assert.Equal(t, []int{0, 10, 40}, errCntManage, "PolicyView.ManageObject() should work correctly")

	// check CanConsume()
	errCntConsume := []int{0, 0, 0}
	for i := 0; i < len(users); i++ {
		policyView := policy.View(users[i])
		for _, obj := range objList {
			if obj.GetKind() == BundleObject.Kind {
				bundle := obj.(*Bundle) // nolint: errcheck
				if _, err := policyView.CanConsume(bundle); err != nil {
					errCntConsume[i]++
				}
			}
		}
	}
	assert.Equal(t, []int{0, 0, 0}, errCntConsume, "PolicyView.CanConsume() should work correctly")
}

func TestPolicyViewManageACLRules(t *testing.T) {
	// users which will be used for viewing policy
	users := []*User{
		{Name: "1", Labels: map[string]string{"is_domain_admin": "true"}},
		{Name: "2", Labels: map[string]string{"is_namespace_admin": "true"}},
		{Name: "3", Labels: map[string]string{"is_consumer": "true"}},
	}

	// check AddObject()
	errCnt := []int{0, 0, 0}
	customRules := []*ACLRule{
		{
			TypeKind: ACLRuleObject.GetTypeKind(),
			Metadata: Metadata{
				Namespace: runtime.SystemNS,
				Name:      "custom_" + NamespaceAdmin.ID,
			},
			Weight:   1000,
			Criteria: &Criteria{RequireAll: []string{"role == 'custom'"}},
			Actions: &ACLRuleActions{
				AddRole: map[string]string{NamespaceAdmin.ID: "test"},
			},
		},
	}
	// check AddObject()
	for i := 0; i < len(users); i++ {
		// make empty policy with custom ACL and add objects to it
		policyView := makeEmptyPolicyWithACL().View(users[i])
		for _, obj := range customRules {
			if policyView.AddObject(obj) != nil {
				errCnt[i]++
			}
		}
	}
	assert.Equal(t, []int{0, 1, 1}, errCnt, "PolicyView.AddObject() should work correctly for ACL rules")
}

func makeEmptyPolicyWithACL() *Policy {
	var aclRules = []*ACLRule{
		// domain admins
		{
			TypeKind: ACLRuleObject.GetTypeKind(),
			Metadata: Metadata{
				Namespace: runtime.SystemNS,
				Name:      "is_domain_admin",
			},
			Weight:   100,
			Criteria: &Criteria{RequireAll: []string{"is_domain_admin"}},
			Actions: &ACLRuleActions{
				AddRole: map[string]string{DomainAdmin.ID: namespaceAll},
			},
		},
		// namespace admins for 'main' namespace
		{
			TypeKind: ACLRuleObject.GetTypeKind(),
			Metadata: Metadata{
				Namespace: runtime.SystemNS,
				Name:      "is_namespace_admin",
			},
			Weight:   200,
			Criteria: &Criteria{RequireAll: []string{"is_namespace_admin"}},
			Actions: &ACLRuleActions{
				AddRole: map[string]string{NamespaceAdmin.ID: "main"},
			},
		},
		// service consumers for 'main' namespace
		{
			TypeKind: ACLRuleObject.GetTypeKind(),
			Metadata: Metadata{
				Namespace: runtime.SystemNS,
				Name:      "is_consumer",
			},
			Weight:   300,
			Criteria: &Criteria{RequireAll: []string{"is_consumer"}},
			Actions: &ACLRuleActions{
				AddRole: map[string]string{ServiceConsumer.ID: "main"},
			},
		},
	}
	policy := NewPolicy()
	for _, rule := range aclRules {
		err := policy.AddObject(rule)
		if err != nil {
			panic(err)
		}
	}
	return policy
}
