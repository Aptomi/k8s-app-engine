package bolt

import (
	lang "github.com/Aptomi/aptomi/pkg/slinga/language"
	"github.com/Aptomi/aptomi/pkg/slinga/object"
	"github.com/Aptomi/aptomi/pkg/slinga/object/codec/yaml"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

func TestBoltStore(t *testing.T) {
	catalog := object.NewObjectCatalog(lang.ServiceObject, lang.ContextObject, lang.ClusterObject, lang.RuleObject, lang.DependencyObject)
	db := NewBoltStore(catalog, yaml.NewCodec(catalog))

	f, err := ioutil.TempFile("", t.Name())
	assert.Nil(t, err, "Temp file should be successfully created")
	defer os.Remove(f.Name())

	err = db.Open(f.Name())
	if err != nil {
		panic(err)
	}

	policy := lang.LoadUnitTestsPolicy("../../../testdata/unittests")

	services := make([]object.BaseObject, 0, len(policy.Services))

	for _, service := range policy.Services {
		err = db.Save(service)
		if err != nil {
			panic(err)
		}
		services = append(services, service)
	}

	assert.Equal(t, 4, len(services), "Len!")

	for _, service := range services {
		obj, err := db.GetByKey(service.GetKey())
		if err != nil {
			panic(err)
		}

		assert.Exactly(t, service, obj, "fail!")
	}
}
