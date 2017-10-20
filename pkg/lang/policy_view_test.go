package lang

import (
	"github.com/Aptomi/aptomi/pkg/object"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPolicyViewCommonObjects(t *testing.T) {
	// make policy with objects
	_, policyOrig := makePolicy()

	// get the list of objects
	objList := []object.Base{}
	for _, obj := range Objects {
		objList = append(objList, policyOrig.GetObjectsByKind(obj.Kind)...)
	}
	assert.NotEmpty(t, objList, "Object list should not be empty")

	// make empty policy with bootstrap ACL
	policy := NewPolicy()
	for _, rule := range ACLRulesBootstrap {
		policy.AddObject(rule)
	}

	policyViews := []*PolicyView{
		policy.View(&User{ID: "1", Name: "1", Labels: map[string]string{"role": "aptomi_domain_admin"}}),
		policy.View(&User{ID: "2", Name: "2", Labels: map[string]string{"role": "aptomi_main_ns_admin"}}),
		policy.View(&User{ID: "3", Name: "3", Labels: map[string]string{"role": "aptomi_main_ns_consumer"}}),
	}

	// check AddObject()
	errCnt := []int{0, 0, 0}
	for _, obj := range objList {
		for i := 0; i < len(policyViews); i++ {
			if policyViews[i].AddObject(obj) != nil {
				errCnt[i]++
			}
		}
	}
	assert.Equal(t, errCnt, []int{0, 10, 40}, "PolicyView.AddObject() should work correctly")

	// check ViewObject() and ManageObject()
	errCntView := []int{0, 0, 0}
	errCntManage := []int{0, 0, 0}
	for _, obj := range objList {
		for i := 0; i < len(policyViews); i++ {
			if _, err := policyViews[i].ViewObject(obj.GetKind(), obj.GetName(), obj.GetNamespace()); err != nil {
				if obj.GetKind() == RuleObject.Kind || obj.GetKind() == ACLRuleObject.Kind || obj.GetKind() == DependencyObject.Kind {
					assert.Contains(t, err.Error(), "not supported")
				} else {
					errCntView[i]++
				}
			}
			if _, err := policyViews[i].ManageObject(obj.GetKind(), obj.GetName(), obj.GetNamespace()); err != nil {
				if obj.GetKind() == RuleObject.Kind || obj.GetKind() == ACLRuleObject.Kind || obj.GetKind() == DependencyObject.Kind {
					assert.Contains(t, err.Error(), "not supported")
				} else {
					errCntManage[i]++
				}
			}
		}
	}
	assert.Equal(t, errCntView, []int{0, 0, 0}, "PolicyView.ViewObject() should work correctly")
	assert.Equal(t, errCntManage, []int{0, 10, 30}, "PolicyView.ManageObject() should work correctly")

	// check CanConsume()
	errCntConsume := []int{0, 0, 0}
	for _, obj := range objList {
		if obj.GetKind() == ServiceObject.Kind {
			service := obj.(*Service)
			for i := 0; i < len(policyViews); i++ {
				if _, err := policyViews[i].CanConsume(service); err != nil {
					errCntConsume[i]++
				}
			}
		}
	}
	assert.Equal(t, errCntConsume, []int{0, 0, 0}, "PolicyView.CanConsume() should work correctly")
}

func TestPolicyViewManageACLRules(t *testing.T) {
	// make empty policy with bootstrap ACL
	policy := NewPolicy()
	for _, rule := range ACLRulesBootstrap {
		policy.AddObject(rule)
	}

	policyViews := []*PolicyView{
		policy.View(&User{ID: "1", Name: "1", Labels: map[string]string{"role": "aptomi_domain_admin"}}),
		policy.View(&User{ID: "2", Name: "2", Labels: map[string]string{"role": "aptomi_main_ns_admin"}}),
		policy.View(&User{ID: "3", Name: "3", Labels: map[string]string{"role": "aptomi_main_ns_consumer"}}),
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
	for _, obj := range customRules {
		for i := 0; i < len(policyViews); i++ {
			if policyViews[i].AddObject(obj) != nil {
				errCnt[i]++
			}
		}
	}
	assert.Equal(t, errCnt, []int{0, 1, 1}, "PolicyView.AddObject() should work correctly for ACL rules")
}
