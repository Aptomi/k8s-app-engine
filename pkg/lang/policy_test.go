package lang

import (
	"github.com/Aptomi/aptomi/pkg/object"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

func TestPolicyGetObjects(t *testing.T) {
	namespace, policy := makePolicy()

	// retrieve objects
	for _, kind := range []string{ServiceObject.Kind, ContractObject.Kind} {
		assert.Equal(t, 10, len(policy.GetObjectsByKind(kind)), "Number of '%s' objects in the policy should be correct", kind)

		for i := 0; i < 10; i++ {
			name := kind + strconv.Itoa(i)
			getObject(t, policy, kind, name, namespace)
		}
	}

	for _, kind := range []string{ClusterObject.Kind} {
		assert.Equal(t, 10, len(policy.GetObjectsByKind(kind)), "Number of '%s' objects in the policy should be correct", kind)

		for i := 0; i < 10; i++ {
			name := kind + strconv.Itoa(i)
			getObject(t, policy, kind, name, object.SystemNS)
		}
	}

	for _, kind := range []string{RuleObject.Kind, DependencyObject.Kind} {
		assert.Equal(t, 10, len(policy.GetObjectsByKind(kind)), "Number of '%s' objects in the policy should be correct", kind)

		for i := 0; i < 10; i++ {
			name := kind + strconv.Itoa(i)
			_, err := policy.GetObject(kind, kind+strconv.Itoa(i), namespace)
			assert.Error(t, err, "Get object by kind '%s' should return an error", name)
		}
	}
}

func getObject(t *testing.T, policy *Policy, kind string, name string, namespace string) {
	// get within current namespace
	obj1, err := policy.GetObject(kind, name, namespace)
	assert.NoError(t, err, "Get object '%s/%s' should be successful", kind, name)
	assert.NotNil(t, obj1, "Get object '%s/%s' should return an object", kind, name)

	// get by absolute path
	obj2, err := policy.GetObject(kind, namespace+"/"+name, "")
	assert.NoError(t, err, "Get object '%s/%s/%s' should be successful", namespace, kind, name)
	assert.NotNil(t, obj2, "Get object '%s/%s/%s' should return an object", namespace, kind, name)

	// get by incorrect path (empty)
	obj3, err := policy.GetObject(kind, "", "")
	assert.Error(t, err, "Get object with incorrect locator (zero parts) should return an error")
	assert.Nil(t, obj3)

	// get by incorrect path (too many parts)
	obj4, err := policy.GetObject(kind, "extrapart"+"/"+namespace+"/"+name, "")
	assert.Error(t, err, "Get object with incorrect locator (too many parts) should return an error")
	assert.Nil(t, obj4)

	// get by incorrect namespace
	obj5, err := policy.GetObject(kind, name, "non-existing-namespace")
	assert.Error(t, err, "Get object with a non-existing namespace should return an error")
	assert.Nil(t, obj5)
}

func makePolicy() (string, *Policy) {
	namespace := "main"
	policy := NewPolicy()
	for i := 0; i < 10; i++ {
		policy.AddObject(&Service{
			Metadata: Metadata{
				Kind:      ServiceObject.Kind,
				Namespace: namespace,
				Name:      "service" + strconv.Itoa(i),
			},
		})
		policy.AddObject(&Contract{
			Metadata: Metadata{
				Kind:      ContractObject.Kind,
				Namespace: namespace,
				Name:      "contract" + strconv.Itoa(i),
			},
		})
		policy.AddObject(&Cluster{
			Metadata: Metadata{
				Kind:      ClusterObject.Kind,
				Namespace: object.SystemNS,
				Name:      "cluster" + strconv.Itoa(i),
			},
		})
		policy.AddObject(&Rule{
			Metadata: Metadata{
				Kind:      RuleObject.Kind,
				Namespace: namespace,
				Name:      "rule" + strconv.Itoa(i),
			},
		})
		policy.AddObject(&Dependency{
			Metadata: Metadata{
				Kind:      DependencyObject.Kind,
				Namespace: namespace,
				Name:      "dependency" + strconv.Itoa(i),
			},
		})
	}
	return namespace, policy
}
