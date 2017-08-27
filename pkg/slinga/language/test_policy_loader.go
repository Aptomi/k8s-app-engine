package language

import (
	. "github.com/Aptomi/aptomi/pkg/slinga/object"
	"github.com/Aptomi/aptomi/pkg/slinga/object/codec/yaml"
	. "github.com/Aptomi/aptomi/pkg/slinga/object/store/file"
)

func LoadUnitTestsPolicy() *PolicyNamespace {
	catalog := NewObjectCatalog()
	catalog.Add(ServiceObject)
	catalog.Add(ContextObject)
	catalog.Add(ClusterObject)
	catalog.Add(RuleObject)
	catalog.Add(DependencyObject)

	store := FileStore{}
	store.Codec = &yaml.YamlCodec{}
	store.Codec.SetObjectCatalog(catalog)
	store.Open("../testdata/unittests")

	policy := NewPolicyNamespace()

	objects, err := store.LoadObjects()
	if err != nil {
		panic("Error while loading test Policy")
	}

	for _, object := range objects {
		policy.AddObject(object)
	}

	return policy

}
