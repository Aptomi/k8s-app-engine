package diff

import (
	"github.com/Aptomi/aptomi/pkg/slinga/engine/apply/actions"
	. "github.com/Aptomi/aptomi/pkg/slinga/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/slinga/external"
	"github.com/Aptomi/aptomi/pkg/slinga/external/secrets"
	"github.com/Aptomi/aptomi/pkg/slinga/external/users"
	. "github.com/Aptomi/aptomi/pkg/slinga/language"
	"github.com/Aptomi/aptomi/pkg/slinga/language/yaml"
	"github.com/stretchr/testify/assert"
	"testing"
)

func getPolicy() *Policy {
	return LoadUnitTestsPolicy("../../testdata/unittests")
}

func getExternalData() *external.Data {
	return external.NewData(
		users.NewUserLoaderFromDir("../../testdata/unittests"),
		secrets.NewSecretLoaderFromDir("../../testdata/unittests"),
	)
}

func resolvePolicy(t *testing.T, policy *Policy, externalData *external.Data) *PolicyResolution {
	resolver := NewPolicyResolver(policy, externalData)
	result, _, err := resolver.ResolveAllDependencies()
	if !assert.Nil(t, err, "Policy should be resolved without errors") {
		t.FailNow()
	}
	return result
}

// TODO: this has to be changed to use the new serialization code instead of serializing to YAML
func emulateSaveAndLoadResolution(resolution *PolicyResolution) *PolicyResolution {
	policyNew := Policy{}
	yaml.DeserializeObject(yaml.SerializeObject(resolution), &policyNew)

	resolutionNew := PolicyResolution{}
	yaml.DeserializeObject(yaml.SerializeObject(resolution), &resolutionNew)

	return &resolutionNew
}

func verifyDiff(t *testing.T, diff *PolicyResolutionDiff, componentInstantiate int, componentDestruct int, componentUpdate int, componentAttachDependency int, componentDetachDependency int) {
	cnt := struct {
		create   int
		update   int
		delete   int
		attach   int
		detach   int
		clusters int
	}{}
	for _, action := range diff.Actions {
		switch action.(type) {
		case *actions.ComponentCreate:
			cnt.create++
		case *actions.ComponentDelete:
			cnt.delete++
		case *actions.ComponentUpdate:
			cnt.update++
		case *actions.ComponentAttachDependency:
			cnt.attach++
		case *actions.ComponentDetachDependency:
			cnt.detach++
		case *actions.ClustersPostProcess:
			cnt.clusters++
		default:
			t.Fatalf("Incorrect action type: %T", action)
		}
	}

	assert.Equal(t, componentInstantiate, cnt.create, "Diff: component instantiations")
	assert.Equal(t, componentDestruct, cnt.delete, "Diff: component destructions")
	assert.Equal(t, componentUpdate, cnt.update, "Diff: component updates")
	assert.Equal(t, componentAttachDependency, cnt.attach, "Diff: dependencies attached to components")
	assert.Equal(t, componentDetachDependency, cnt.detach, "Diff: dependencies removed from components")
	assert.Equal(t, 1, cnt.clusters, "Diff: all clusters post processing")
}
