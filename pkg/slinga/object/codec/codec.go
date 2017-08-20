package codec

import (
	. "github.com/Aptomi/aptomi/pkg/slinga/object"
)

type MarshalUnmarshaler interface {
	GetName() string
	SetObjectCatalog(catalog *ObjectCatalog) // todo encoders/decoders will register it internally only on this function call
	MarshalOne(object BaseObject) ([]byte, error)
	MarshalMany(objects []BaseObject) ([]byte, error)
	UnmarshalOne(data []byte) (BaseObject, error)
	UnmarshalOneOrMany(data []byte) ([]BaseObject, error)
}

type BaseMarshalUnmarshaler struct {
	Catalog *ObjectCatalog
}

func (codec *BaseMarshalUnmarshaler) SetObjectCatalog(catalog *ObjectCatalog) {
	codec.Catalog = catalog
}
