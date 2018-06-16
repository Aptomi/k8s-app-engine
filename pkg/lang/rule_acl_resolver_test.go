package lang

import (
	"testing"

	"github.com/Aptomi/aptomi/pkg/runtime"
	"github.com/stretchr/testify/assert"
)

type aclTestCase struct {
	user             *User
	role             *ACLRole
	namespace        string
	expected         bool
	objectPrivileges []testCaseObjPrivileges
}

type testCaseObjPrivileges struct {
	obj      Base
	expected *Privilege
}

func (testCase aclTestCase) print(t *testing.T) {
	if testCase.expected {
		t.Logf("User '%s': expected role '%s' in namespace '%s'", testCase.user.Name, testCase.role.Name, testCase.namespace)
	} else {
		t.Logf("User '%s': expected NOT to have role '%s' in namespace '%s'", testCase.user.Name, testCase.role.Name, testCase.namespace)
	}
}

func (privileges testCaseObjPrivileges) print(t *testing.T, testCase aclTestCase) {
	t.Logf("Object '%s' in namespace '%s', accessed by user '%s'", privileges.obj.GetKind(), privileges.obj.GetNamespace(), testCase.user.Name)
}

func runACLTests(testCases []aclTestCase, rules []*ACLRule, t *testing.T) {
	aclRules := make(map[string]*Rule)
	for _, rule := range rules {
		aclRules[rule.GetName()] = rule
	}
	resolver := NewACLResolver(aclRules)
	for _, tc := range testCases {
		roleMap, err := resolver.GetUserRoleMap(tc.user)
		if !assert.NoError(t, err, "User role map should be retrieved successfully") {
			continue
		}
		if !assert.Equal(t, tc.expected, roleMap[tc.role.ID][tc.namespace], "User role map should be correct") {
			tc.print(t)
		}

		for _, tcObj := range tc.objectPrivileges {
			privilege, errPrivilege := resolver.GetUserPrivileges(tc.user, tcObj.obj)
			if !assert.NoError(t, errPrivilege, "User privileges should be retrieved successfully") {
				continue
			}
			if !assert.Equal(t, tcObj.expected, privilege, "User privilege should be correct") {
				tcObj.print(t, tc)
			}
		}
	}
}

func TestAclResolver(t *testing.T) {
	var rules = []*ACLRule{
		// domain admins
		{
			TypeKind: ACLRuleObject.GetTypeKind(),
			Metadata: Metadata{
				Namespace: runtime.SystemNS,
				Name:      "is_domain_admin",
			},
			Weight:   100,
			Criteria: &Criteria{RequireAll: []string{"is_domain_admin"}},
			Actions: &RuleActions{
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
			Actions: &RuleActions{
				AddRole: map[string]string{NamespaceAdmin.ID: "main"},
			},
		},
		// service consumers for 'main2' namespace
		{
			TypeKind: ACLRuleObject.GetTypeKind(),
			Metadata: Metadata{
				Namespace: runtime.SystemNS,
				Name:      "is_consumer",
			},
			Weight:   300,
			Criteria: &Criteria{RequireAll: []string{"is_consumer"}},
			Actions: &RuleActions{
				AddRole: map[string]string{ServiceConsumer.ID: "main1, main2 ,main3,main4"},
			},
		},
		// bogus rule
		{
			TypeKind: ACLRuleObject.GetTypeKind(),
			Metadata: Metadata{
				Namespace: runtime.SystemNS,
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
			user:      &User{Name: "1", Labels: map[string]string{"is_domain_admin": "true"}},
			role:      DomainAdmin,
			namespace: namespaceAll,
			expected:  true,
			objectPrivileges: []testCaseObjPrivileges{
				{obj: &Cluster{TypeKind: ClusterObject.GetTypeKind(), Metadata: Metadata{Namespace: runtime.SystemNS}}, expected: fullAccess},
				{obj: &Service{TypeKind: ServiceObject.GetTypeKind(), Metadata: Metadata{Namespace: "somens"}}, expected: fullAccess},
			},
		},
		{
			user:      &User{Name: "2", Labels: map[string]string{"is_namespace_admin": "true"}},
			role:      NamespaceAdmin,
			namespace: "main",
			expected:  true,
			objectPrivileges: []testCaseObjPrivileges{
				{obj: &Cluster{TypeKind: ClusterObject.GetTypeKind(), Metadata: Metadata{Namespace: runtime.SystemNS}}, expected: viewAccess},
				{obj: &Service{TypeKind: ServiceObject.GetTypeKind(), Metadata: Metadata{Namespace: "somens"}}, expected: viewAccess},
				{obj: &Service{TypeKind: ServiceObject.GetTypeKind(), Metadata: Metadata{Namespace: "main"}}, expected: fullAccess},
			},
		},
		{
			user:      &User{Name: "3", Labels: map[string]string{"is_consumer": "true"}},
			role:      ServiceConsumer,
			namespace: "main2",
			expected:  true,
			objectPrivileges: []testCaseObjPrivileges{
				{obj: &Cluster{TypeKind: ClusterObject.GetTypeKind(), Metadata: Metadata{Namespace: runtime.SystemNS}}, expected: viewAccess},
				{obj: &Service{TypeKind: ServiceObject.GetTypeKind(), Metadata: Metadata{Namespace: "somens"}}, expected: viewAccess},
				{obj: &Service{TypeKind: ServiceObject.GetTypeKind(), Metadata: Metadata{Namespace: "main"}}, expected: viewAccess},
				{obj: &Dependency{TypeKind: DependencyObject.GetTypeKind(), Metadata: Metadata{Namespace: "somens"}}, expected: viewAccess},
				{obj: &Dependency{TypeKind: DependencyObject.GetTypeKind(), Metadata: Metadata{Namespace: "main2"}}, expected: fullAccess},
			},
		},
		{
			user:      &User{Name: "4", Labels: map[string]string{"name": "value"}},
			role:      nobody,
			namespace: "main",
			expected:  false,
			objectPrivileges: []testCaseObjPrivileges{
				{obj: &Cluster{TypeKind: ClusterObject.GetTypeKind(), Metadata: Metadata{Namespace: runtime.SystemNS}}, expected: viewAccess},
				{obj: &Contract{TypeKind: ContractObject.GetTypeKind(), Metadata: Metadata{Namespace: "somens"}}, expected: viewAccess},
				{obj: &Contract{TypeKind: ContractObject.GetTypeKind(), Metadata: Metadata{Namespace: "main"}}, expected: viewAccess},
			},
		},
	}

	runACLTests(testCases, rules, t)
}

func TestAclResolverAdminUser(t *testing.T) {
	var rules = []*ACLRule{}
	testCases := []aclTestCase{
		{
			user:      &User{Name: "1", DomainAdmin: true},
			role:      DomainAdmin,
			namespace: namespaceAll,
			expected:  true,
		},
	}
	runACLTests(testCases, rules, t)
}
