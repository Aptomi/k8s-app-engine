package yaml

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type test struct {
	Str     string
	Number  int
	Nested  nested
	Hidden1 string `yaml:"-"`
	hidden2 string
}

type nested struct {
	NestedStrs []string
}

func TestYamlCodec(t *testing.T) {
	assert.Equal(t, "yaml", YamlCodec.GetName(), "Correct codec name expected")

	object := test{
		"str",
		42,
		nested{
			[]string{
				"nstr1",
				"nstr2",
				"nstr3",
			},
		},
		"hidden1",
		"hidden2",
	}

	yaml := `str: str
number: 42
nested:
  nestedstrs:
  - nstr1
  - nstr2
  - nstr3
`

	data, err := YamlCodec.Marshal(object)
	assert.Nil(t, err, "There should be no errors marshalling provided object")
	assert.Equal(t, yaml, string(data), "Correct marshaled bytes expected")

	yamlWithHiddens := `str: str
number: 42
hidden1: hidden1
hidden2: hidden2
nested:
  nestedstrs:
  - nstr1
  - nstr2
  - nstr3
`

	unmarshObject := &test{}
	err = YamlCodec.Unmarshal([]byte(yamlWithHiddens), unmarshObject)
	assert.Nil(t, err, "There should be no errors unmarshaling provided bytes")
	assert.Equal(t, object.Str, unmarshObject.Str, "Correct field value expected")
	assert.Equal(t, object.Number, unmarshObject.Number, "Correct field value expected")
	assert.Equal(t, object.Nested, unmarshObject.Nested, "Correct nested struct field value expected")
	assert.Empty(t, unmarshObject.Hidden1, "Hidden field should be empty")
	assert.Empty(t, unmarshObject.hidden2, "Hidden field should be empty")

	badYaml := "str @ str"

	err = YamlCodec.Unmarshal([]byte(badYaml), &test{})
	assert.NotNil(t, err, "Unmarshaling should fail in case of invalid yaml")
}
