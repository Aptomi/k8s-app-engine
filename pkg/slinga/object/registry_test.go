package object

import (
	"github.com/Aptomi/aptomi/pkg/slinga/object/codec/fake"
	"github.com/Aptomi/aptomi/pkg/slinga/object/codec/yaml"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestRegistry(t *testing.T) {
	reg := Registry{yaml.YamlCodec, make(map[string]Constructor)}
	reg.AddKind("test", func() BaseObject { return new(testObject) })

	data := reg.MarshalOne(testObjectInstance)
	expectedYaml := `metadata:
  kind: test
  uid: 4d0e5391-83ef-11e7-876b-784f435b826b
  generation: 42
  name: name
  namespace: namespace
testfield: test
`
	assert.Equal(t, expectedYaml, string(data), "Correct marshal yaml string expected")

	newObj := reg.UnmarshalOne(data)
	assertTestObjectInstance(t, newObj)
	assert.True(t, reflect.DeepEqual(testObjectInstance, newObj), "Objects should be deep equal")
}

func TestRegistryCodecErrors(t *testing.T) {
	reg := Registry{fake.ErrCodec, make(map[string]Constructor)}
	reg.AddKind("test", func() BaseObject { return new(testObject) })

	assert.Panics(t, func() { reg.MarshalOne(testObjectInstance) })
	assert.Panics(t, func() { reg.UnmarshalOne(make([]byte, 0)) })
}

func TestRegistryUnknownKind(t *testing.T) {
	reg := Registry{yaml.YamlCodec, make(map[string]Constructor)}

	assert.Panics(t, func() { reg.UnmarshalOne(make([]byte, 0)) })
}
