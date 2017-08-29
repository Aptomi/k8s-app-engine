package codec

import (
	. "github.com/Aptomi/aptomi/pkg/slinga/object"
)

type MarshalUnmarshaler interface {
	GetName() string
	MarshalOne(object BaseObject) ([]byte, error)
	MarshalMany(objects []BaseObject) ([]byte, error)
	UnmarshalOne(data []byte) (BaseObject, error)
	UnmarshalOneOrMany(data []byte) ([]BaseObject, error)
}
