package lang

import (
	"github.com/Aptomi/aptomi/pkg/object"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPolicyViewCommonObjects(t *testing.T) {
	// make policy with objects
	_, policyWithObjects := makePolicyWithObjects()

	// get the list of objects
	objList := []object.Base{}
	for _, obj := range Objects {
		objList = append(objList, policyWithObjects.GetObjectsByKind(obj.Kind)...)
	}
	assert.NotEmpty(t, objList, "Object list should not be empty")

	// users which will be used for viewing policy
	users := []*User{
		{ID: "1", Name: "1", Labels: map[string]string{"is_domain_admin": "true"}},
		{ID: "2", Name: "2", Labels: map[string]string{"is_namespace_admin": "true"}},
		{ID: "3", Name: "3", Labels: map[string]string{"is_consumer": "true"}},
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
			if _, err := policyView.ViewObject(obj.GetKind(), obj.GetName(), obj.GetNamespace()); err != nil {
				errCntView[i]++
			}
			if _, err := policyView.ManageObject(obj.GetKind(), obj.GetName(), obj.GetNamespace()); err != nil {
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
			if obj.GetKind() == ServiceObject.Kind {
				service := obj.(*Service)
				if _, err := policyView.CanConsume(service); err != nil {
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
		{ID: "1", Name: "1", Labels: map[string]string{"is_domain_admin": "true"}},
		{ID: "2", Name: "2", Labels: map[string]string{"is_namespace_admin": "true"}},
		{ID: "3", Name: "3", Labels: map[string]string{"is_consumer": "true"}},
	}

	// check AddObject()
	errCnt := []int{0, 0, 0}
	customRules := []*ACLRule{
		{
			Metadata: Metadata{
				Kind:      ACLRuleObject.Kind,
				Namespace: object.SystemNS,
				Name:      "custom_" + namespaceAdmin.ID,
			},
			Weight:   1000,
			Criteria: &Criteria{RequireAll: []string{"role == 'custom'"}},
			Actions: &RuleActions{
				AddRole: map[string]string{namespaceAdmin.ID: "test"},
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
			Metadata: Metadata{
				Kind:      ACLRuleObject.Kind,
				Namespace: object.SystemNS,
				Name:      "is_domain_admin",
			},
			Weight:   100,
			Criteria: &Criteria{RequireAll: []string{"is_domain_admin"}},
			Actions: &RuleActions{
				AddRole: map[string]string{domainAdmin.ID: namespaceAll},
			},
		},
		// namespace admins for 'main' namespace
		{
			Metadata: Metadata{
				Kind:      ACLRuleObject.Kind,
				Namespace: object.SystemNS,
				Name:      "is_namespace_admin",
			},
			Weight:   200,
			Criteria: &Criteria{RequireAll: []string{"is_namespace_admin"}},
			Actions: &RuleActions{
				AddRole: map[string]string{namespaceAdmin.ID: "main"},
			},
		},
		// service consumers for 'main' namespace
		{
			Metadata: Metadata{
				Kind:      ACLRuleObject.Kind,
				Namespace: object.SystemNS,
				Name:      "is_consumer",
			},
			Weight:   300,
			Criteria: &Criteria{RequireAll: []string{"is_consumer"}},
			Actions: &RuleActions{
				AddRole: map[string]string{serviceConsumer.ID: "main"},
			},
		},
	}
	policy := NewPolicy()
	for _, rule := range aclRules {
		policy.AddObject(rule)
	}
	return policy
}
