package codec

import (
	"github.com/Aptomi/aptomi/pkg/slinga/object/codec/fake"
	"github.com/Aptomi/aptomi/pkg/slinga/object/codec/yaml"
	"github.com/stretchr/testify/assert"
	"testing"

	"reflect"
)

type test struct {
	S string
}

func TestMarshalUnmarshal(t *testing.T) {
	obj := test{"test"}
	to := &test{}

	MarshalUnmarshal(yaml.YamlCodec, obj, to)
	assert.True(t, reflect.DeepEqual(obj, *to), "Objects should be equal after Marshal and Unmarshal")

	assert.Panics(t, func() { MarshalUnmarshal(fake.ErrCodec, obj, to) })
}
