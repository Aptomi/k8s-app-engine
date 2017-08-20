package yaml

import (
	. "github.com/Aptomi/aptomi/pkg/slinga/object/codec"
	. "github.com/Aptomi/aptomi/pkg/slinga/object/codec/codectest"
	"github.com/stretchr/testify/assert"
	"testing"
)

func newTestYamlCodec() MarshalUnmarshaler {
	codec := &YamlCodec{}
	codec.SetObjectCatalog(CodecTestObjectsCatalog)

	return codec
}

var (
	CodecTestObjectsYamls = []string{`metadata:
  kind: t1
  uid: uid-1
  generation: 1
  name: name-1
  namespace: namespace-1
str: str-1
number: 1
`,
		`metadata:
  kind: t1
  uid: uid-2
  generation: 2
  name: name-2
  namespace: namespace-1
str: str-2
number: 2
`,
		`metadata:
  kind: t2
  uid: uid-3
  generation: 3
  name: name-3
  namespace: namespace-2
nested:
  nestedstrs:
  - "1"
  - "2"
  - "3"
map:
  k-1:
  - uid-1$1
  - uid-2$2
  k-2:
  - uid-3$1
  - uid-4$2
`,
		`metadata:
  kind: t2
  uid: uid-4
  generation: 4
  name: name-4
  namespace: namespace-2
nested:
  nestedstrs:
  - "4"
  - "5"
  - "6"
map:
  k-3:
  - uid-5$1
  - uid-6$2
  k-4:
  - uid-7$1
  - uid-8$2
`,
	}
	CodecTestObjectsYaml = `- metadata:
    kind: t1
    uid: uid-1
    generation: 1
    name: name-1
    namespace: namespace-1
  str: str-1
  number: 1
- metadata:
    kind: t1
    uid: uid-2
    generation: 2
    name: name-2
    namespace: namespace-1
  str: str-2
  number: 2
- metadata:
    kind: t2
    uid: uid-3
    generation: 3
    name: name-3
    namespace: namespace-2
  nested:
    nestedstrs:
    - "1"
    - "2"
    - "3"
  map:
    k-1:
    - uid-1$1
    - uid-2$2
    k-2:
    - uid-3$1
    - uid-4$2
- metadata:
    kind: t2
    uid: uid-4
    generation: 4
    name: name-4
    namespace: namespace-2
  nested:
    nestedstrs:
    - "4"
    - "5"
    - "6"
  map:
    k-3:
    - uid-5$1
    - uid-6$2
    k-4:
    - uid-7$1
    - uid-8$2
`
)

func TestYamlCodec_GetName(t *testing.T) {
	codec := newTestYamlCodec()
	assert.Equal(t, "yaml", codec.GetName(), "Correct codec name expected")
}

func TestYamlCodec_MarshalOne(t *testing.T) {
	codec := newTestYamlCodec()

	for idx, obj := range CodecTestObjects {
		data, err := codec.MarshalOne(obj)
		assert.Nil(t, err, "Error while marshaling test object #%d", idx)
		assert.Equal(t, CodecTestObjectsYamls[idx], string(data), "Correct marshaled bytes expected for test object #%d", idx)
	}
}

func TestYamlCodec_MarshalMany(t *testing.T) {
	codec := newTestYamlCodec()

	data, err := codec.MarshalMany(CodecTestObjects)
	assert.Nil(t, err, "Error while marshaling test objects")
	assert.Equal(t, CodecTestObjectsYaml, string(data), "Correct marshaled bytes expected for test objects")
}

func TestYamlCodec_UnmarshalOne(t *testing.T) {
	codec := newTestYamlCodec()

	for idx, str := range CodecTestObjectsYamls {
		obj, err := codec.UnmarshalOne([]byte(str))

		assert.Nil(t, err, "Error while unmarshaling test object #%d", idx)
		assert.Exactly(t, CodecTestObjects[idx], obj, "Deep equal unmarshaled object of the same type expected for test object #%d", idx)
	}
}

func TestYamlCodec_UnmarshalOneOrMany(t *testing.T) {
	codec := newTestYamlCodec()
	obj, err := codec.UnmarshalOneOrMany([]byte(CodecTestObjectsYaml))

	assert.Nil(t, err, "Error while unmarshaling test objects")
	assert.Exactly(t, CodecTestObjects, obj, "Deep equal unmarshaled object of the same type expected for test objects")
}

// TODO(slukjanov): add tests for hidden values
// TODO(slukjanov): add tests for bad yaml
// TODO(slukjanov): add tests for incorrect one/many passed to unmarshal
