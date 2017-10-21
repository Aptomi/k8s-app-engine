package lang

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/object"
)

// PolicyView an struct which allows to view/manage policy on behalf on a certain user
// It will enforce all ACLs, allowing the user to only perform actions which he is allowed to perform
type PolicyView struct {
	Policy *Policy
	User   *User
}

// NewPolicyView creates a new PolicyView
func NewPolicyView(policy *Policy, user *User) *PolicyView {
	return &PolicyView{
		Policy: policy,
		User:   user,
	}
}

// AddObject adds an object into the policy, putting it into the corresponding namespace
// If an error occurs (e.g. user doesn't have ACL permissions to perform an operation, or object validation error), then it will be returned
func (view *PolicyView) AddObject(obj object.Base) error {
	privilege, err := view.Policy.aclResolver.GetUserPrivileges(view.User, obj)
	if err != nil {
		return err
	}
	if !privilege.Manage {
		return fmt.Errorf("user '%s' doesn't have ACL permissions to manage object '%s/%s/%s'", view.User.ID, obj.GetNamespace(), obj.GetKind(), obj.GetName())
	}
	return view.Policy.AddObject(obj)
}

// ViewObject checks if user has permissions to view an object. If user has no permissions, ACL error will be returned
func (view *PolicyView) ViewObject(obj object.Base) error {
	privilege, err := view.Policy.aclResolver.GetUserPrivileges(view.User, obj)
	if err != nil {
		return err
	}
	if !privilege.View {
		return fmt.Errorf("user '%s' doesn't have ACL permissions to view object '%s/%s/%s'", view.User.ID, obj.GetNamespace(), obj.GetKind(), obj.GetName())
	}
	return nil
}

// ManageObject checks if user has permissions to manage an object. If user has no permissions, ACL error will be returned
func (view *PolicyView) ManageObject(obj object.Base) error {
	privilege, err := view.Policy.aclResolver.GetUserPrivileges(view.User, obj)
	if err != nil {
		return err
	}
	if !privilege.Manage {
		return fmt.Errorf("user '%s' doesn't have ACL permissions to manage object '%s/%s/%s'", view.User.ID, obj.GetNamespace(), obj.GetKind(), obj.GetName())
	}
	return nil
}

// CanConsume returns if user has permissions to consume a service
// If a user can declare a dependency in a given namespace, then he can essentially can consume the service
func (view *PolicyView) CanConsume(service *Service) (bool, error) {
	obj := &Metadata{Namespace: service.GetNamespace(), Kind: DependencyObject.Kind}
	privilege, err := view.Policy.aclResolver.GetUserPrivileges(view.User, obj)
	if err != nil {
		return false, err
	}
	if !privilege.Manage {
		return false, fmt.Errorf("user '%s' doesn't have ACL permissions to consume service '%s/%s'", view.User.ID, service.GetNamespace(), service.GetName())
	}
	return true, nil
}
