package bolt

import (
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/lang/builder"
	"github.com/Aptomi/aptomi/pkg/object"
	"github.com/Aptomi/aptomi/pkg/object/codec/yaml"
	"github.com/Aptomi/aptomi/pkg/util"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

var (
	testCatalog = object.NewCatalog().Append(lang.Objects...)
	testCodec   = yaml.NewCodec(testCatalog)
)

func TestBoltStore(t *testing.T) {
	db := NewBoltStore(testCatalog, testCodec)

	f, err := ioutil.TempFile("", t.Name())
	assert.NoError(t, err, "Temp file should be successfully created")
	defer os.Remove(f.Name()) // nolint: errcheck

	err = db.Open(f.Name())
	if err != nil {
		panic(err)
	}

	policy := makePolicyObjects(t)

	for _, obj := range policy {
		updated, err := db.Save(obj)
		if err != nil {
			panic(err)
		}
		assert.True(t, updated, "Object saved for the first time")
	}

	for _, obj := range policy {
		objSaved, err := db.GetByName(obj.GetNamespace(), obj.GetKind(), obj.GetName(), obj.GetGeneration())
		if err != nil {
			panic(err)
		}

		assert.Exactly(t, obj, objSaved, "Saved object is not equal to the original object in the policy")
	}

	services, err := db.GetAll("main", lang.ServiceObject.Kind)
	assert.NoError(t, err, "Error while getting all services from main namespace")
	assert.Len(t, services, 10, "Get all should return correct number of services")
	for _, obj := range policy {
		if obj.GetKind() == lang.ServiceObject.Kind {
			assert.Contains(t, services, obj, "Get all should return all services")
		}
	}
}

/*
	Helpers
*/

func makePolicyObjects(t *testing.T) []object.Base {
	t.Helper()
	b := builder.NewPolicyBuilder()

	for i := 0; i < 10; i++ {
		// create a service
		service := b.AddService(b.AddUser())
		b.AddServiceComponent(service,
			b.CodeComponent(
				util.NestedParameterMap{
					"param":   "{{ .Labels.param }}",
					"cluster": "{{ .Labels.cluster }}",
				},
				nil,
			),
		)
		contract := b.AddContract(service, b.CriteriaTrue())

		// add rules to allow all dependencies
		clusterObj := b.AddCluster()
		b.AddRule(b.CriteriaTrue(), b.RuleActions(lang.Allow, lang.Allow, lang.NewLabelOperationsSetSingleLabel(lang.LabelCluster, clusterObj.Name)))

		// add dependency
		dependency := b.AddDependency(b.AddUser(), contract)
		dependency.Labels["param"] = "value1"
	}

	// This hack is needed to make sure that we'll get test data in the same way like after marshaling objects
	// and storing them in DB. Example: empty fields will be stored anyway, while we omitting them in test data.

	objects := make([]object.Base, 0)
	policy := b.Policy()
	for _, kind := range testCatalog.Kinds {
		objects = append(objects, policy.GetObjectsByKind(kind.Kind)...)
	}

	data, err := testCodec.MarshalMany(objects)
	if err != nil {
		t.Errorf("Error marshaling policy objects: %s", err)
	}
	objects, err = testCodec.UnmarshalOneOrMany(data)
	if err != nil {
		t.Errorf("Error unmarshaling policy objects: %s", err)
	}

	return objects
}
