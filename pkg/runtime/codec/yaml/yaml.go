package yaml

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/runtime"
	utilyaml "github.com/ghodss/yaml"
	"gopkg.in/yaml.v2"
)

type yamlCodec struct {
	registry *runtime.Registry
	json     bool
}

func NewCodec(registry *runtime.Registry) runtime.Codec {
	return &yamlCodec{
		registry: registry,
		json:     false,
	}
}

func NewJSONCodec(registry *runtime.Registry) runtime.Codec {
	return &yamlCodec{
		registry: registry,
		json:     true,
	}
}

// yamlCodec implements runtime.Codec
var _ runtime.Codec = &yamlCodec{}

func (cod *yamlCodec) EncodeOne(obj runtime.Object) ([]byte, error) {
	return cod.encode(obj)
}

func (cod *yamlCodec) EncodeMany(objs []runtime.Object) ([]byte, error) {
	return cod.encode(objs)
}

func (cod *yamlCodec) DecodeOne(data []byte) (runtime.Object, error) {
	objects, err := cod.decodeOneOrMany(data, true)
	if err != nil {
		return nil, err
	}

	return objects[0], nil
}

func (cod *yamlCodec) DecodeOneOrMany(data []byte) ([]runtime.Object, error) {
	return cod.decodeOneOrMany(data, false)
}

func (cod *yamlCodec) encode(obj interface{}) ([]byte, error) {
	data, err := yaml.Marshal(obj)
	if err != nil {
		return nil, err
	}

	if cod.json {
		data, err = utilyaml.YAMLToJSON(data)
	}
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (cod *yamlCodec) decodeOneOrMany(data []byte, strictOne bool) ([]runtime.Object, error) {
	raw := new(interface{})
	err := yaml.Unmarshal(data, raw)
	if err != nil {
		return nil, fmt.Errorf("error while decoding data to raw interface{}: %s", err)
	}

	result := make([]runtime.Object, 0)

	if elem, single := (*raw).(map[interface{}]interface{}); single { // if it's a single object (map)
		obj, rawErr := cod.decodeRaw(elem, data)
		if rawErr != nil {
			return nil, fmt.Errorf("error while decoding single object: %s", rawErr)
		}

		result = append(result, obj)
	} else if strictOne { // if single object strictly required
		return nil, fmt.Errorf("single object expected, but found more")
	} else if rawSlice, slice := (*raw).([]interface{}); slice { // if it's an object slice
		for idx, rawElem := range rawSlice {
			sliceElem, isMap := rawElem.(map[interface{}]interface{}) // each slice elem should be map
			if !isMap {
				return nil, fmt.Errorf("element #%d isn't an object", idx)
			}

			elemData, elemErr := yaml.Marshal(sliceElem) // get []byte for current elem only
			if elemErr != nil {
				return nil, fmt.Errorf("error while decoding element #%d (decode step): %s", idx, elemErr)
			}

			obj, elemErr := cod.decodeRaw(sliceElem, elemData) // decode to kind type
			if elemErr != nil {
				return nil, fmt.Errorf("error while decoding element #%d (final step): %s", idx, elemErr)
			}

			result = append(result, obj)
		}
	} else { // if it's not an object or object slice
		return nil, fmt.Errorf("decoding data (not an object or object rawSlice): %T", raw)
	}

	return result, nil
}

func (cod *yamlCodec) decodeRaw(single map[interface{}]interface{}, data []byte) (runtime.Object, error) {
	kindField, ok := single["kind"]
	if !ok {
		return nil, fmt.Errorf("can't find kind field in metadata: %v", single)
	}

	kind, ok := kindField.(string)
	if !ok {
		return nil, fmt.Errorf("kind field in metadata isn't a string: %v", single)
	}

	if len(kind) == 0 {
		return nil, fmt.Errorf("empty kind")
	}

	info := cod.registry.Get(kind)
	if info == nil {
		return nil, fmt.Errorf("unknown kind: %s", kind)
	}

	obj := info.New()
	err := yaml.Unmarshal(data, obj)
	if err != nil {
		return nil, err
	}

	return obj, nil
}
