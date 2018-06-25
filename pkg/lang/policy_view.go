package lang

import (
	"fmt"
)

// PolicyView allows to view/manage policy objects on behalf on a certain user
// It will enforce all ACLs, allowing the user to only perform actions which he is entitled to perform.
type PolicyView struct {
	// Policy which gets viewed
	Policy *Policy

	// User who is viewing policy
	User *User

	// Resolver is access control rules (who can access which objects in which policy namespaces)
	Resolver *ACLResolver
}

// NewPolicyView creates a new PolicyView
func NewPolicyView(policy *Policy, user *User, resolver *ACLResolver) *PolicyView {
	return &PolicyView{
		Policy:   policy,
		User:     user,
		Resolver: resolver,
	}
}

// AddObject adds an object into the policy. When you add objects to the policy, they get added to the corresponding
// Namespace. If error occurs (e.g. object has an unknown kind, etc) then the error will be returned
func (view *PolicyView) AddObject(obj Base) error {
	privilege, err := view.Resolver.GetUserPrivileges(view.User, obj)
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
func (view *PolicyView) ViewObject(obj Base) error {
	privilege, err := view.Resolver.GetUserPrivileges(view.User, obj)
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
func (view *PolicyView) ManageObject(obj Base) error {
	privilege, err := view.Resolver.GetUserPrivileges(view.User, obj)
	if err != nil {
		return err
	}
	if !privilege.Manage {
		return fmt.Errorf("user '%s' doesn't have ACL permissions to manage object '%s/%s/%s'", view.User.Name, obj.GetNamespace(), obj.GetKind(), obj.GetName())
	}
	return nil
}

// CanConsume returns if user has permissions to consume a given bundle.
// If a user can declare a claim in a given namespace, then he can essentially can consume the bundle
func (view *PolicyView) CanConsume(bundle *Bundle) (bool, error) {
	obj := &Claim{
		TypeKind: ClaimObject.GetTypeKind(),
		Metadata: Metadata{
			Namespace: bundle.GetNamespace(),
		},
	}
	privilege, err := view.Resolver.GetUserPrivileges(view.User, obj)
	if err != nil {
		return false, err
	}
	if !privilege.Manage {
		return false, fmt.Errorf("user '%s' doesn't have ACL permissions to consume bundle '%s/%s'", view.User.Name, bundle.GetNamespace(), bundle.GetName())
	}
	return true, nil
}
