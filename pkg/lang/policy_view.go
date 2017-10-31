package lang

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/object"
)

// PolicyView allows to view/manage policy objects on behalf on a certain user
// It will enforce all ACLs, allowing the user to only perform actions which he is entitled to perform.
type PolicyView struct {
	// Policy which gets viewed
	Policy *Policy

	// User who is viewing policy
	User *User
}

// NewPolicyView creates a new PolicyView
func NewPolicyView(policy *Policy, user *User) *PolicyView {
	return &PolicyView{
		Policy: policy,
		User:   user,
	}
}

// AddObject adds an object into the policy. When you add objects to the policy, they get added to the corresponding
// Namespace. If error occurs (e.g. object validation error, objects with duplicate names, etc) then the error
// will be returned
func (view *PolicyView) AddObject(obj object.Base) error {
	privilege, err := view.Policy.aclResolver.GetUserPrivileges(view.User, obj)
	if err != nil {
		return err
	}
	if !privilege.Manage {
		return fmt.Errorf("user '%s' doesn't have ACL permissions to manage object '%s/%s/%s'", view.User.Name, obj.GetNamespace(), obj.GetKind(), obj.GetName())
	}
	return view.Policy.AddObject(obj)
}

// ViewObject checks if user has permissions to view a given object. If user has no permissions, then ACL error
// will be returned
func (view *PolicyView) ViewObject(obj object.Base) error {
	privilege, err := view.Policy.aclResolver.GetUserPrivileges(view.User, obj)
	if err != nil {
		return err
	}
	if !privilege.View {
		return fmt.Errorf("user '%s' doesn't have ACL permissions to view object '%s/%s/%s'", view.User.Name, obj.GetNamespace(), obj.GetKind(), obj.GetName())
	}
	return nil
}

// ManageObject checks if user has permissions to manage a given object. If user has no permissions, then ACL error
// will be returned
func (view *PolicyView) ManageObject(obj object.Base) error {
	privilege, err := view.Policy.aclResolver.GetUserPrivileges(view.User, obj)
	if err != nil {
		return err
	}
	if !privilege.Manage {
		return fmt.Errorf("user '%s' doesn't have ACL permissions to manage object '%s/%s/%s'", view.User.Name, obj.GetNamespace(), obj.GetKind(), obj.GetName())
	}
	return nil
}

// CanConsume returns if user has permissions to consume a given service.
// If a user can declare a dependency in a given namespace, then he can essentially can consume the service
func (view *PolicyView) CanConsume(service *Service) (bool, error) {
	obj := &Metadata{Namespace: service.GetNamespace(), Kind: DependencyObject.Kind}
	privilege, err := view.Policy.aclResolver.GetUserPrivileges(view.User, obj)
	if err != nil {
		return false, err
	}
	if !privilege.Manage {
		return false, fmt.Errorf("user '%s' doesn't have ACL permissions to consume service '%s/%s'", view.User.Name, service.GetNamespace(), service.GetName())
	}
	return true, nil
}
