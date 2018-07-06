package store

import (
	"bytes"
	"encoding/gob"
	"encoding/json"

	"gopkg.in/yaml.v2"
)

type Codec interface {
	Marshal(value interface{}) ([]byte, error)
	Unmarshal(data []byte, value interface{}) error
}

type jsonCodec struct {
}

func NewJsonCodec() Codec {
	return &jsonCodec{}
}

func (c *jsonCodec) Marshal(value interface{}) ([]byte, error) {
	return json.Marshal(value)
}

func (c *jsonCodec) Unmarshal(data []byte, value interface{}) error {
	return json.Unmarshal(data, value)
}

type yamlCodec struct {
}

func NewYamlCodec() Codec {
	return &yamlCodec{}
}

func (c *yamlCodec) Marshal(value interface{}) ([]byte, error) {
	return yaml.Marshal(value)
}

func (c *yamlCodec) Unmarshal(data []byte, value interface{}) error {
	return yaml.Unmarshal(data, value)
}

type gobCodec struct {
}

func NewGobCodec() Codec {
	return &gobCodec{}
}

func (c *gobCodec) Marshal(value interface{}) ([]byte, error) {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)

	err := encoder.Encode(value)
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func (c *gobCodec) Unmarshal(data []byte, value interface{}) error {
	var buffer bytes.Buffer
	decoder := gob.NewDecoder(&buffer)

	_, err := buffer.Write(data)
	if err != nil {
		return err
	}

	return decoder.Decode(value)
}
