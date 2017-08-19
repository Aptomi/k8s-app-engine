package object

import (
	"github.com/Aptomi/aptomi/pkg/slinga/object/codec/yaml"
	"github.com/stretchr/testify/assert"
	"testing"
)

type t1 struct {
	Metadata

	T1 string
}

type t2 struct {
	Metadata

	T2 string
}

func newTestRegistry() *Registry {
	reg := Registry{yaml.YamlCodec, make(map[string]Constructor)}
	reg.AddKind("t1", func() BaseObject { return new(t1) })
	reg.AddKind("t2", func() BaseObject { return new(t2) })

	return &reg
}

var (
	testObjs = []BaseObject{
		&t1{
			Metadata{
				Kind: "t1",
				Name: "t1-name",
			},
			"t1p",
		},
		&t2{
			Metadata{
				Kind: "t2",
				Name: "t2-name",
			},
			"t2p",
		},
	}
	testObjsMarshaled = []string{
		`metadata:
  kind: t1
  uid: ""
  generation: 0
  name: t1-name
  namespace: ""
t1: t1p
`,
		`metadata:
  kind: t2
  uid: ""
  generation: 0
  name: t2-name
  namespace: ""
t2: t2p
`,
	}
	testObjsSliceMarshaled = `- metadata:
    kind: t1
    uid: ""
    generation: 0
    name: t1-name
    namespace: ""
  t1: t1p
- metadata:
    kind: t2
    uid: ""
    generation: 0
    name: t2-name
    namespace: ""
  t2: t2p
`
)

func TestRegistry_AddKind(t *testing.T) {
	reg := newTestRegistry()

	assert.Contains(t, reg.kinds, "t1", "Registry should know kind t1")
	assert.Contains(t, reg.kinds, "t2", "Registry should know kind t2")
}

func TestRegistry_MarshalOne(t *testing.T) {
	reg := newTestRegistry()

	data, err := reg.MarshalOne(testObjs[0])
	assert.Nil(t, err, "Object should be marshaled w/o errors")
	assert.Equal(t, testObjsMarshaled[0], string(data), "Correct marshaled data expected")
}

func TestRegistry_MarshalMany(t *testing.T) {
	reg := newTestRegistry()

	data, err := reg.MarshalMany(testObjs)
	assert.Nil(t, err, "Objects should be marshaled w/o errors")
	assert.Equal(t, testObjsSliceMarshaled, string(data), "Correct marshaled data expected")
}

func TestRegistry_UnmarshalOne(t *testing.T) {
	reg := newTestRegistry()

	obj, err := reg.UnmarshalOne([]byte(testObjsMarshaled[0]))
	assert.Nil(t, err, "Object should be unmarshaled w/o errors")
	assert.Exactly(t, testObjs[0], obj, "Unmarshaled object should be deep equal to initial one")
}

func TestRegistry_UnmarshalMany(t *testing.T) {
	reg := newTestRegistry()

	obj, err := reg.UnmarshalMany([]byte(testObjsSliceMarshaled))
	assert.Nil(t, err, "Objects should be unmarshaled w/o errors")
	assert.Exactly(t, testObjs, obj, "Unmarshaled objects should be deep equal to initial ones")
}
