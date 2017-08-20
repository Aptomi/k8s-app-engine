package yaml

import (
	"fmt"
	. "github.com/Aptomi/aptomi/pkg/slinga/object"
	"github.com/Aptomi/aptomi/pkg/slinga/object/codec"
	"gopkg.in/yaml.v2"
)

type YamlCodec struct {
	codec.BaseMarshalUnmarshaler
}

// YamlCodecName is the name of Yaml MarshalUnmarshaler implementation
const YamlCodecName = "yaml"

func (c *YamlCodec) GetName() string {
	return YamlCodecName
}

func (c *YamlCodec) MarshalOne(object BaseObject) ([]byte, error) {
	return yaml.Marshal(&object)
}

func (c *YamlCodec) MarshalMany(objects []BaseObject) ([]byte, error) {
	return yaml.Marshal(&objects)
}

func (c *YamlCodec) UnmarshalOne(data []byte) (BaseObject, error) {
	objects, err := c.unmarshalOneOrMany(data, true)
	if err != nil {
		return nil, err
	}

	return objects[0], nil
}

func (c *YamlCodec) UnmarshalOneOrMany(data []byte) ([]BaseObject, error) {
	return c.unmarshalOneOrMany(data, false)
}

func (c *YamlCodec) unmarshalOneOrMany(data []byte, strictOne bool) ([]BaseObject, error) {
	raw := new(interface{})
	err := yaml.Unmarshal(data, raw)
	if err != nil {
		return nil, fmt.Errorf("Error while unmarshaling data to raw interface{}: %s", err)
	}

	result := make([]BaseObject, 0)

	if elem, ok := (*raw).(map[interface{}]interface{}); ok { // if it's a single object (map)
		obj, err := c.unmarshalRaw(elem, data)
		if err != nil {
			return nil, fmt.Errorf("Error while unmarshaling single object: %s", err)
		}

		result = append(result, obj)
	} else if strictOne { // if single object strictly required
		return nil, fmt.Errorf("Single object expected")
	} else if slice, ok := (*raw).([]interface{}); ok { // if it's an object slice
		for idx, rawElem := range slice {
			elem, ok := rawElem.(map[interface{}]interface{}) // each slice elem should be map
			if !ok {
				return nil, fmt.Errorf("Element #%d isn't an object", idx)
			}

			elemData, err := yaml.Marshal(elem) // get []byte for current elem only
			if err != nil {
				return nil, fmt.Errorf("Error while unmarshaling element #%d (marshal step): %s", idx, err)
			}

			obj, err := c.unmarshalRaw(elem, elemData) // unmarshal to kind type
			if err != nil {
				return nil, fmt.Errorf("Error while unmarshaling element #%d (final step): %s", idx, err)
			}

			result = append(result, obj)
		}
	} else { // if it's not an object or object slice
		return nil, fmt.Errorf("Unmarshalable data (not an object or object slice): %T", raw)
	}

	return result, nil
}

func (c *YamlCodec) unmarshalRaw(single map[interface{}]interface{}, data []byte) (BaseObject, error) {
	metaField, ok := single["metadata"]
	if !ok {
		return nil, fmt.Errorf("Can't find metadata field inside object: %v", single)
	}

	meta, ok := metaField.(map[interface{}]interface{})
	if !ok {
		return nil, fmt.Errorf("Metadata field isn't a map: %v", single)
	}

	kindField, ok := meta["kind"]
	if !ok {
		return nil, fmt.Errorf("Can't find kind field in metadata: %v", single)
	}

	kind, ok := kindField.(string)
	if !ok {
		return nil, fmt.Errorf("Kind field in metadata isn't a string: %v", single)
	}

	objectInfo := c.Catalog.Get(Kind(kind))
	if objectInfo == nil {
		return nil, fmt.Errorf("Unknown kind: %s", kind)
	}

	obj := objectInfo.New()
	err := yaml.Unmarshal(data, obj)
	if err != nil {
		return nil, err
	}

	return obj, nil
}
