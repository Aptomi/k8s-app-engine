package object

import (
	. "github.com/Aptomi/aptomi/pkg/slinga/log"
	"github.com/Aptomi/aptomi/pkg/slinga/object/codec"
	log "github.com/Sirupsen/logrus"
)

type Constructor func() BaseObject

type Registry struct {
	codec codec.MarshalUnmarshaler
	kinds map[string]Constructor
}

func (registry Registry) AddKind(kind string, specUnmarshaler Constructor) {
	registry.kinds[kind] = specUnmarshaler
}

func (registry Registry) MarshalOne(object BaseObject) []byte {
	data, err := registry.codec.Marshal(&object)
	if err != nil {
		Debug.WithFields(log.Fields{
			"meta":  object,
			"error": err,
		}).Panic("Can't marshal meta", err)
	}
	return data
}

func (registry Registry) UnmarshalOne(data []byte) BaseObject {
	meta := &emptyObject{}
	err := registry.codec.Unmarshal(data, meta)
	if err != nil {
		Debug.WithFields(log.Fields{
			"data":  data,
			"error": err,
		}).Panic("Can't unmarshal object metadata (empty object)", err)
	}

	kind := meta.Metadata.Kind.String()
	kindConstructor, ok := registry.kinds[kind]
	if !ok {
		Debug.WithFields(log.Fields{
			"data":  data,
			"kind":  kind,
			"error": err,
		}).Panic("Unknown object kind: %s", kind)
	}

	obj := kindConstructor()
	err = registry.codec.Unmarshal(data, obj)

	return obj
}
