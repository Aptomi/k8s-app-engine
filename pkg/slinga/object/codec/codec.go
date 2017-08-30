package codec

import (
	"github.com/Aptomi/aptomi/pkg/slinga/object"
)

type MarshalUnmarshaler interface {
	GetName() string
	MarshalOne(object object.BaseObject) ([]byte, error)
	MarshalMany(objects []object.BaseObject) ([]byte, error)
	UnmarshalOne(data []byte) (object.BaseObject, error)
	UnmarshalOneOrMany(data []byte) ([]object.BaseObject, error)
}
