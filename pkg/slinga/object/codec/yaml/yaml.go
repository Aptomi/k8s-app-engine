package yaml

import (
	"gopkg.in/yaml.v2"
)

type yamlCodec struct {
}

// YamlCodecName is the name of Yaml MarshalUnmarshaler implementation
const YamlCodecName = "yaml"

// YamlCodec is the instance of Yaml MarshalUnmarshaler
var YamlCodec = yamlCodec{}

func (c yamlCodec) GetName() string {
	return YamlCodecName
}

func (c yamlCodec) Marshal(value interface{}) ([]byte, error) {
	return yaml.Marshal(&value)
}

func (c yamlCodec) Unmarshal(data []byte, value interface{}) error {
	return yaml.Unmarshal(data, value)
}
