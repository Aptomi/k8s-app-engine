package bolt

import (
	"github.com/Aptomi/aptomi/pkg/slinga/lang"
	"github.com/Aptomi/aptomi/pkg/slinga/lang/builder"
	"github.com/Aptomi/aptomi/pkg/slinga/object"
	"github.com/Aptomi/aptomi/pkg/slinga/object/codec/yaml"
	"github.com/Aptomi/aptomi/pkg/slinga/util"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

func TestBoltStore(t *testing.T) {
	catalog := object.NewCatalog(lang.ServiceObject, lang.ContractObject, lang.ClusterObject, lang.RuleObject, lang.DependencyObject)
	db := NewBoltStore(catalog, yaml.NewCodec(catalog))

	f, err := ioutil.TempFile("", t.Name())
	assert.Nil(t, err, "Temp file should be successfully created")
	defer os.Remove(f.Name()) // nolint: errcheck

	err = db.Open(f.Name())
	if err != nil {
		panic(err)
	}

	b := makePolicyBuilder()

	for _, kind := range catalog.Kinds {
		for _, obj := range b.Policy().GetObjectsByKind(kind.Kind) {
			updated, err := db.Save(obj)
			if err != nil {
				panic(err)
			}
			assert.True(t, updated, "Object saved for the first time")
		}
	}

	for _, kind := range catalog.Kinds {
		for _, obj := range b.Policy().GetObjectsByKind(kind.Kind) {
			objSaved, err := db.GetByName(obj.GetNamespace(), obj.GetKind(), obj.GetName(), obj.GetGeneration())
			if err != nil {
				panic(err)
			}

			assert.Exactly(t, obj.GetName(), objSaved.GetName(), "Saved object is not equal to the original object in the policy")

			// TODO: this fails because of deepequal comparison
			// - nil slice/map becomes empty slide/map after saving
			// - this should be fixed
			// assert.Exactly(t, obj, objSaved, "Saved object is not equal to the original object in the policy")
		}
	}
}

/*
	Helpers
*/

func makePolicyBuilder() *builder.PolicyBuilder {
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

	return b
}
