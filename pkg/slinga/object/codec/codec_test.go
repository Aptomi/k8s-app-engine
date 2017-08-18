package codec

import (
	"github.com/Aptomi/aptomi/pkg/slinga/object/codec/yaml"
	"github.com/stretchr/testify/assert"
	"testing"
)

type test struct {
	S string
}

func TestMarshalUnmarshal(t *testing.T) {
	obj := test{"test"}
	to := &test{}

	MarshalUnmarshal(yaml.YamlCodec, obj, to)
	assert.Equal(t, obj, *to)
}
