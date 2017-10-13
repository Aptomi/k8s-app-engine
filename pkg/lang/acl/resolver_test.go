package acl

import (
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/object"
	"github.com/stretchr/testify/assert"
	"testing"
)

// This is an initial POC implementation of ACL, not integrated with the rest of Aptomi codebase yet
// TODO: add tests for admins/consumers with specific set of namespaces
// TODO: introduce its own data struct for ACL rule instead of relying on lang.Rule
// TODO: run linter

func TestAclResolverBootstrapRules(t *testing.T) {
	resolver := NewResolver(BootstrapAclRules)
	role, err := resolver.GetUserRole(&lang.User{ID: "1"})
	assert.NoError(t, err, "User role should be retrieved successfully")
	assert.Equal(t, DomainAdmin.ID, role.ID, "Everyone should be a domain admin in bootstrap")
}

func TestAclResolverCustomRules(t *testing.T) {
	var rules = []*Rule{
		{
			Metadata: lang.Metadata{
				Kind:      RuleObject.Kind,
				Namespace: object.SystemNS,
				Name:      "is_domain_admin",
			},
			Weight:   100,
			Criteria: &lang.Criteria{RequireAll: []string{"is_domain_admin"}},
			Actions: &lang.RuleActions{
				ChangeLabels: lang.ChangeLabelsAction(lang.NewLabelOperationsSetSingleLabel(LabelRole, DomainAdmin.ID)),
				Stop:         true,
			},
		},
		// every single person is a namespace admin
		{
			Metadata: lang.Metadata{
				Kind:      RuleObject.Kind,
				Namespace: object.SystemNS,
				Name:      "is_namespace_admin",
			},
			Weight:   200,
			Criteria: &lang.Criteria{RequireAll: []string{"is_namespace_admin"}},
			Actions: &lang.RuleActions{
				ChangeLabels: lang.ChangeLabelsAction(lang.NewLabelOperationsSetSingleLabel(LabelRole, NamespaceAdmin.ID)),
				Stop:         true,
			},
		},
		// every single person is a service consumer
		{
			Metadata: lang.Metadata{
				Kind:      RuleObject.Kind,
				Namespace: object.SystemNS,
				Name:      "is_consumer",
			},
			Weight:   300,
			Criteria: &lang.Criteria{RequireAll: []string{"is_consumer"}},
			Actions: &lang.RuleActions{
				ChangeLabels: lang.ChangeLabelsAction(lang.NewLabelOperationsSetSingleLabel(LabelRole, ServiceConsumer.ID)),
				Stop:         true,
			},
		},
	}

	resolver := NewResolver(rules)

	testCases := []struct {
		user *lang.User
		role *Role
	}{
		{
			user: &lang.User{ID: "1", Labels: map[string]string{"is_domain_admin": "true"}},
			role: DomainAdmin,
		},
		{
			user: &lang.User{ID: "2", Labels: map[string]string{"is_namespace_admin": "true"}},
			role: NamespaceAdmin,
		},
		{
			user: &lang.User{ID: "3", Labels: map[string]string{"is_consumer": "true"}},
			role: ServiceConsumer,
		},
		{
			user: &lang.User{ID: "4", Labels: map[string]string{"name": "value"}},
			role: Nobody,
		},
	}

	for _, tc := range testCases {
		role, err := resolver.GetUserRole(tc.user)
		assert.NoError(t, err, "User role should be retrieved successfully")
		assert.Equal(t, tc.role.ID, role.ID, "Roles should match")
	}
}
