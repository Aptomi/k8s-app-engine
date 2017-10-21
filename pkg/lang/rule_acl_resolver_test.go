package lang

import (
	"github.com/Aptomi/aptomi/pkg/object"
	"github.com/stretchr/testify/assert"
	"testing"
)

type aclTestCase struct {
	user             *User
	role             *ACLRole
	namespace        string
	expected         bool
	objectPrivileges []testCaseObjPrivileges
}

type testCaseObjPrivileges struct {
	obj      object.Base
	expected *Privilege
}

func runACLTests(testCases []aclTestCase, rules []*ACLRule, t *testing.T) {
	globalRules := NewGlobalRules()
	globalRules.addRule(rules...)
	resolver := NewACLResolver(globalRules)
	for _, tc := range testCases {
		roleMap, err := resolver.getUserRoleMap(tc.user)
		assert.NoError(t, err, "User role map should be retrieved successfully. Test = %s", tc)
		assert.Equal(t, tc.expected, roleMap[tc.role.ID][tc.namespace], "User role map should be correct. Test = %s", tc)

		for _, tcObj := range tc.objectPrivileges {
			privilege, errPrivilege := resolver.GetUserPrivileges(tc.user, tcObj.obj)
			assert.NoError(t, errPrivilege, "User privileges should be retrieved successfully. Test = %s", tcObj)
			assert.Equal(t, tcObj.expected, privilege, "User privilege should be correct. Test = %s", tcObj)
		}
	}
}

func TestAclResolverBootstrapRules(t *testing.T) {
	testCases := []aclTestCase{
		{
			user:      &User{ID: "1", Labels: map[string]string{"role": "aptomi_domain_admin"}},
			role:      domainAdmin,
			namespace: namespaceAll,
			expected:  true,
			objectPrivileges: []testCaseObjPrivileges{
				{obj: &Metadata{Namespace: object.SystemNS, Kind: ClusterObject.Kind}, expected: fullAccess},
				{obj: &Metadata{Namespace: "somens", Kind: ServiceObject.Kind}, expected: fullAccess},
			},
		},
		{
			user:      &User{ID: "2", Labels: map[string]string{"role": "aptomi_main_ns_admin"}},
			role:      namespaceAdmin,
			namespace: "main",
			expected:  true,
			objectPrivileges: []testCaseObjPrivileges{
				{obj: &Metadata{Namespace: object.SystemNS, Kind: ClusterObject.Kind}, expected: viewAccess},
				{obj: &Metadata{Namespace: "somens", Kind: ServiceObject.Kind}, expected: viewAccess},
				{obj: &Metadata{Namespace: "main", Kind: ServiceObject.Kind}, expected: fullAccess},
			},
		},
		{
			user:      &User{ID: "3", Labels: map[string]string{"role": "aptomi_main_ns_consumer"}},
			role:      serviceConsumer,
			namespace: "main",
			expected:  true,
			objectPrivileges: []testCaseObjPrivileges{
				{obj: &Metadata{Namespace: object.SystemNS, Kind: ClusterObject.Kind}, expected: viewAccess},
				{obj: &Metadata{Namespace: "somens", Kind: ServiceObject.Kind}, expected: viewAccess},
				{obj: &Metadata{Namespace: "main", Kind: ServiceObject.Kind}, expected: viewAccess},
				{obj: &Metadata{Namespace: "somens", Kind: DependencyObject.Kind}, expected: viewAccess},
				{obj: &Metadata{Namespace: "main", Kind: DependencyObject.Kind}, expected: fullAccess},
			},
		},
		{
			user:      &User{ID: "4", Labels: map[string]string{"name": "value"}},
			role:      nobody,
			namespace: "main",
			expected:  false,
			objectPrivileges: []testCaseObjPrivileges{
				{obj: &Metadata{Namespace: object.SystemNS, Kind: ClusterObject.Kind}, expected: viewAccess},
				{obj: &Metadata{Namespace: "somens", Kind: ContractObject.Kind}, expected: viewAccess},
				{obj: &Metadata{Namespace: "main", Kind: ContractObject.Kind}, expected: viewAccess},
			},
		},
	}

	runACLTests(testCases, ACLRulesBootstrap, t)
}

func TestAclResolverCustomRules(t *testing.T) {
	var rules = []*ACLRule{
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
		// namespace admins for 'test' namespace
		{
			Metadata: Metadata{
				Kind:      ACLRuleObject.Kind,
				Namespace: object.SystemNS,
				Name:      "is_namespace_admin",
			},
			Weight:   200,
			Criteria: &Criteria{RequireAll: []string{"is_namespace_admin"}},
			Actions: &RuleActions{
				AddRole: map[string]string{namespaceAdmin.ID: "test"},
			},
		},
		// service consumers for 'test2' namespace
		{
			Metadata: Metadata{
				Kind:      ACLRuleObject.Kind,
				Namespace: object.SystemNS,
				Name:      "is_consumer",
			},
			Weight:   300,
			Criteria: &Criteria{RequireAll: []string{"is_consumer"}},
			Actions: &RuleActions{
				AddRole: map[string]string{serviceConsumer.ID: "test2"},
			},
		},
		// bogus rule
		{
			Metadata: Metadata{
				Kind:      ACLRuleObject.Kind,
				Namespace: object.SystemNS,
				Name:      "some_bogus_rule",
			},
			Weight:   400,
			Criteria: &Criteria{RequireAll: []string{"true"}},
			Actions: &RuleActions{
				AddRole: map[string]string{"unknown-role": "some-value"},
			},
		},
	}

	testCases := []aclTestCase{
		{
			user:      &User{ID: "1", Labels: map[string]string{"is_domain_admin": "true"}},
			role:      domainAdmin,
			namespace: namespaceAll,
			expected:  true,
		},
		{
			user:      &User{ID: "2", Labels: map[string]string{"is_namespace_admin": "true"}},
			role:      namespaceAdmin,
			namespace: "test",
			expected:  true,
		},
		{
			user:      &User{ID: "3", Labels: map[string]string{"is_consumer": "true"}},
			role:      serviceConsumer,
			namespace: "test2",
			expected:  true,
		},
		{
			user:      &User{ID: "4", Labels: map[string]string{"name": "value"}},
			role:      nobody,
			namespace: "test",
			expected:  false,
		},
	}

	runACLTests(testCases, rules, t)
}
