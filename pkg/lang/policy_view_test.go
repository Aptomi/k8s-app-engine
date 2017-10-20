package lang

import (
	"github.com/Aptomi/aptomi/pkg/object"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPolicyViewCommonObjects(t *testing.T) {
	// make policy with objects
	_, policyOrig := makePolicyWithObjects(t)

	// get the list of objects
	objList := []object.Base{}
	for _, obj := range Objects {
		objList = append(objList, policyOrig.GetObjectsByKind(obj.Kind)...)
	}
	assert.NotEmpty(t, objList, "Object list should not be empty")

	// users which will be used for viewing policy
	userViews := []*User{
		{ID: "1", Name: "1", Labels: map[string]string{"role": "aptomi_domain_admin"}},
		{ID: "2", Name: "2", Labels: map[string]string{"role": "aptomi_main_ns_admin"}},
		{ID: "3", Name: "3", Labels: map[string]string{"role": "aptomi_main_ns_consumer"}},
	}

	// check AddObject()
	errCnt := []int{0, 0, 0}
	for i := 0; i < len(userViews); i++ {
		// make empty policy with bootstrap ACL and add objects to it
		policyView := makeEmptyPolicy(t).View(userViews[i])
		for _, obj := range objList {
			if policyView.AddObject(obj) != nil {
				errCnt[i]++
			}
		}
	}
	bootstrapRulesCnt := len(ACLRulesBootstrap) // they already exist in the policy and can't be added again
	assert.Equal(t, []int{0 + bootstrapRulesCnt, 10 + bootstrapRulesCnt, 40 + bootstrapRulesCnt}, errCnt, "PolicyView.AddObject() should work correctly")

	// check ViewObject() and ManageObject() on policy with objects
	errCntView := []int{0, 0, 0}
	errCntManage := []int{0, 0, 0}
	for i := 0; i < len(userViews); i++ {
		policyView := policyOrig.View(userViews[i])
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
	assert.Equal(t, []int{0, 10 + bootstrapRulesCnt, 40 + bootstrapRulesCnt}, errCntManage, "PolicyView.ManageObject() should work correctly")

	// check CanConsume()
	errCntConsume := []int{0, 0, 0}
	for i := 0; i < len(userViews); i++ {
		policyView := policyOrig.View(userViews[i])
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
	// make empty policy with bootstrap ACL
	policy := NewPolicy()
	for _, rule := range ACLRulesBootstrap {
		err := policy.AddObject(rule)
		assert.NoError(t, err, "Bootstrap ACL rule should be added successfully")
	}

	// users which will be used for viewing policy
	userViews := []*User{
		{ID: "1", Name: "1", Labels: map[string]string{"role": "aptomi_domain_admin"}},
		{ID: "2", Name: "2", Labels: map[string]string{"role": "aptomi_main_ns_admin"}},
		{ID: "3", Name: "3", Labels: map[string]string{"role": "aptomi_main_ns_consumer"}},
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
	for i := 0; i < len(userViews); i++ {
		// make empty policy with bootstrap ACL and add objects to it
		policyView := makeEmptyPolicy(t).View(userViews[i])
		for _, obj := range customRules {
			if policyView.AddObject(obj) != nil {
				errCnt[i]++
			}
		}
	}
	assert.Equal(t, []int{0, 1, 1}, errCnt, "PolicyView.AddObject() should work correctly for ACL rules")
}
