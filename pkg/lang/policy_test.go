package lang

import (
	"github.com/Aptomi/aptomi/pkg/object"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

func TestPolicy(t *testing.T) {
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

	// retrieve objects
	catalog := object.NewCatalog().Append(ServiceObject, ContractObject)
	for _, kind := range catalog.Kinds {
		assert.Equal(t, 10, len(policy.GetObjectsByKind(kind.Kind)), "Number of '%s' objects in the policy should be correct", kind.Kind)

		for i := 0; i < 10; i++ {
			name := kind.Kind + strconv.Itoa(i)

			// get within current namespace
			obj1, err := policy.GetObject(kind.Kind, name, namespace)
			assert.NoError(t, err, "Get object by kind '%s' should be successful", name)
			assert.NotNil(t, obj1, "Get object by kind '%s' should return an object", name)

			// get by absolute path
			obj2, err := policy.GetObject(kind.Kind, namespace+"/"+name, "")
			assert.NoError(t, err, "Get object by kind '%s' should be successful", name)
			assert.NotNil(t, obj2, "Get object by kind '%s' should return an object", name)
		}
	}

	catalog = object.NewCatalog().Append(ClusterObject)
	for _, kind := range catalog.Kinds {
		assert.Equal(t, 10, len(policy.GetObjectsByKind(kind.Kind)), "Number of '%s' objects in the policy should be correct", kind.Kind)

		for i := 0; i < 10; i++ {
			name := kind.Kind + strconv.Itoa(i)

			// get within current namespace
			obj1, err := policy.GetObject(kind.Kind, name, object.SystemNS)
			assert.NoError(t, err, "Get object by kind '%s' should be successful", name)
			assert.NotNil(t, obj1, "Get object by kind '%s' should return an object", name)

			// get by absolute path
			obj2, err := policy.GetObject(kind.Kind, object.SystemNS+"/"+name, "")
			assert.NoError(t, err, "Get object by kind '%s' should be successful", name)
			assert.NotNil(t, obj2, "Get object by kind '%s' should return an object", name)
		}
	}

	catalog = object.NewCatalog().Append(RuleObject, DependencyObject)
	for _, kind := range catalog.Kinds {
		assert.Equal(t, 10, len(policy.GetObjectsByKind(kind.Kind)), "Number of '%s' objects in the policy should be correct", kind.Kind)

		for i := 0; i < 10; i++ {
			name := kind.Kind + strconv.Itoa(i)
			_, err := policy.GetObject(kind.Kind, kind.Kind+strconv.Itoa(i), namespace)
			assert.Error(t, err, "Get object by kind '%s' should return an error", name)
		}
	}
}
