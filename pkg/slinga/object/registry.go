package object

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/slinga/object/codec"
)

type Constructor func() BaseObject

type Registry struct {
	codec codec.MarshalUnmarshaler
	kinds map[string]Constructor
}

func (reg *Registry) AddKind(kind string, constructor Constructor) {
	// todo support default kind?
	reg.kinds[kind] = constructor
}

func (reg *Registry) MarshalOne(object BaseObject) ([]byte, error) {
	return reg.codec.Marshal(&object)
}

func (reg *Registry) MarshalMany(objects []BaseObject) ([]byte, error) {
	return reg.codec.Marshal(&objects)
}

func (reg *Registry) UnmarshalOne(data []byte) (BaseObject, error) {
	objects, err := reg.UnmarshalMany(data)
	if err != nil {
		return nil, err
	}

	objectsLen := len(objects)
	if objectsLen != 1 {
		return nil, fmt.Errorf("Single object expected, but %d unmarshaled", objectsLen)
	}

	return objects[0], nil
}

func (reg *Registry) UnmarshalMany(data []byte) ([]BaseObject, error) {
	raw := new(interface{})
	err := reg.codec.Unmarshal(data, raw)
	if err != nil {
		return nil, fmt.Errorf("Error while unmarshaling data to raw interface{}: %s", err)
	}

	result := make([]BaseObject, 0)

	if elem, ok := (*raw).(map[interface{}]interface{}); ok { // if it's a single object (map)
		obj, err := reg.unmarshalRaw(elem, data)
		if err != nil {
			return nil, fmt.Errorf("Error while unmarshaling single object: %s", err)
		}

		result = append(result, obj)
	} else if slice, ok := (*raw).([]interface{}); ok { // if it's an object slice
		for idx, rawElem := range slice {
			elem, ok := rawElem.(map[interface{}]interface{}) // each slice elem should be map
			if !ok {
				return nil, fmt.Errorf("Element #%d isn't an object", idx)
			}

			elemData, err := reg.codec.Marshal(elem) // get []byte for current elem only
			if err != nil {
				return nil, fmt.Errorf("Error while unmarshaling element #%d (marshal step): %s", idx, err)
			}

			obj, err := reg.unmarshalRaw(elem, elemData) // unmarshal to kind type
			if err != nil {
				return nil, fmt.Errorf("Error while unmarshaling element #%d (final step): %s", idx, err)
			}

			result = append(result, obj)
		}
	}

	return result, nil
}

func (reg *Registry) unmarshalRaw(single map[interface{}]interface{}, data []byte) (BaseObject, error) {
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

	kindConstructor, ok := reg.kinds[kind]
	if !ok {
		// todo support default kind?
		return nil, fmt.Errorf("Unknown kind: %s", kind)
	}

	obj := kindConstructor()
	err := reg.codec.Unmarshal(data, obj)
	if err != nil {
		return nil, err
	}

	return obj, nil
}
